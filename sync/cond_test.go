// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/nxgtw/go-ipc/internal/test"

	"github.com/stretchr/testify/assert"
)

const (
	testCondName    = "ipccond"
	testCondMutName = "ipccondmut"
)

func makeTestCond(a *assert.Assertions) (cond *Cond, l IPCLocker, err error) {
	a.NoError(DestroyCond(testCondName))
	a.NoError(DestroyMutex(testCondMutName))
	l, err = NewMutex(testCondMutName, os.O_CREATE|os.O_EXCL, 0666)
	if !a.NoError(err) {
		return
	}
	cond, err = NewCond(testCondName, os.O_CREATE|os.O_EXCL, 0666, l)
	if !a.NoError(err) {
		return
	}
	return
}

func destroyTestCond(a *assert.Assertions, cond *Cond, l IPCLocker) {
	a.NoError(cond.Destroy())
	a.NoError(l.Close())
	a.NoError(DestroyMutex(testCondMutName))
}

func TestCondWait(t *testing.T) {
	a := assert.New(t)
	cond, l, err := makeTestCond(a)
	if err != nil {
		return
	}
	defer destroyTestCond(a, cond, l)
	l.Lock()
	endCh := make(chan struct{})
	go func() {
		time.Sleep(time.Millisecond * 50)
		cond.Signal()
		endCh <- struct{}{}
	}()
	cond.Wait()
	<-endCh
}

func TestCondWaitTimeout(t *testing.T) {
	a := assert.New(t)
	cond, l, err := makeTestCond(a)
	if err != nil {
		return
	}
	defer destroyTestCond(a, cond, l)
	l.Lock()
	a.False(cond.WaitTimeout(time.Millisecond * 50))
}

func TestCondBroadcast(t *testing.T) {
	a := assert.New(t)
	cond, l, err := makeTestCond(a)
	if err != nil {
		return
	}
	defer destroyTestCond(a, cond, l)
	var wg1, wg2 sync.WaitGroup
	wg1.Add(8)
	wg2.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			l.Lock()
			wg1.Done()
			cond.Wait()
			l.Unlock()
			wg2.Done()
		}()
	}
	wg1.Wait()
	time.Sleep(time.Millisecond * 250)
	cond.Broadcast()
	wg2.Wait()
}

func TestCondMissedSignal(t *testing.T) {
	a := assert.New(t)
	cond, l, err := makeTestCond(a)
	if err != nil {
		return
	}
	defer destroyTestCond(a, cond, l)
	cond.Signal()
	l.Lock()
	a.False(cond.WaitTimeout(0))
}

func TestCondSignalAnotherProcess(t *testing.T) {
	a := assert.New(t)
	cond, l, err := makeTestCond(a)
	if err != nil {
		return
	}
	defer destroyTestCond(a, cond, l)
	ch := make(chan struct{})
	go func() {
		l.Lock()
		cond.Wait()
		l.Unlock()
		ch <- struct{}{}
	}()
	args := argsForCondSignalCommand(testCondName)
	result := testutil.RunTestApp(args, nil)
	if !a.NoError(result.Err) {
		t.Logf("test app error. the output is: %s", result.Output)
	}
	select {
	case <-ch:
	case <-time.After(time.Second * 3):
		t.Errorf("timeout")
	}
}

func TestCondBroadcastAnotherProcess(t *testing.T) {
	a := assert.New(t)
	cond, l, err := makeTestCond(a)
	if err != nil {
		return
	}
	defer destroyTestCond(a, cond, l)
	ch := make(chan struct{})
	for i := 0; i < 8; i++ {
		go func() {
			l.Lock()
			cond.Wait()
			l.Unlock()
			ch <- struct{}{}
		}()
	}
	args := argsForCondBroadcastCommand(testCondName)
	result := testutil.RunTestApp(args, nil)
	if !a.NoError(result.Err) {
		t.Logf("test app error. the output is: %s", result.Output)
	}
	for i := 0; i < 8; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second * 3):
			t.Errorf("timeout")
		}
	}
}

func TestCondWaitAnotherProcess(t *testing.T) {
	const waitEvent = "condWaitEvent"
	a := assert.New(t)
	if !a.NoError(DestroyEvent(waitEvent)) {
		return
	}
	cond, l, err := makeTestCond(a)
	if err != nil {
		return
	}
	defer destroyTestCond(a, cond, l)
	ev, err := NewEvent(waitEvent, os.O_CREATE|os.O_EXCL, 0666, true)
	if a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(ev.Destroy())
	}()
	args := argsForCondWaitCommand(testCondName, testCondMutName, waitEvent)
	killCh := make(chan bool, 1)
	ch := testutil.RunTestAppAsync(args, nil)
	if !a.True(ev.WaitTimeout(time.Second*3), "wait event was not set") {
		return
	}
	time.Sleep(time.Millisecond * 100)
	cond.Signal()
	select {
	case res := <-ch:
		if res.Err != nil {
			t.Errorf("app error: %v. the output is %q", res.Err, res.Output)
		}
		killCh <- true
	case <-time.After(time.Second * 3):
		t.Errorf("timeout")
	}
}
