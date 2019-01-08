// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nxgtw/go-ipc/internal/test"

	"github.com/stretchr/testify/assert"
)

const (
	testEventName = "testevent"
)

func TestEventOpenMode(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_RDWR, 0666, false)
	a.Error(err)
	ev, err = NewEvent(testEventName, os.O_WRONLY, 0666, false)
	a.Error(err)
	ev, err = NewEvent(testEventName, os.O_CREATE, 0666, false)
	if !a.NoError(err) {
		return
	}
	defer func(ev *Event) {
		a.NoError(ev.Destroy())
	}(ev)
	ev, err = NewEvent(testEventName, 0, 0666, false)
	if !a.NoError(err) {
		return
	}
	a.NoError(ev.Close())
}

func TestEventOpenMode2(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE, 0666, false)
	if !a.NoError(err) {
		return
	}
	defer func(ev *Event) {
		a.NoError(ev.Destroy())
	}(ev)
	ev, err = NewEvent(testEventName, 0, 0666, false)
	if !a.NoError(err) {
		return
	}
	defer func(ev *Event) {
		a.NoError(ev.Close())
	}(ev)
	ev, err = NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.Error(err) {
		ev.Destroy()
	}
}

func TestEventOpenMode3(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.NoError(err) {
		return
	}
	defer func(ev *Event) {
		a.NoError(ev.Destroy())
	}(ev)
	ev, err = NewEvent(testEventName, os.O_CREATE, 0666, false)
	if !a.NoError(err) {
		return
	}
	a.NoError(ev.Close())
}

func TestEventOpenMode4(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE, 0666, false)
	if !a.NoError(err) {
		return
	}
	defer func(ev *Event) {
		a.NoError(ev.Destroy())
	}(ev)
	ev, err = NewEvent(testEventName, 0, 0666, false)
	if !a.NoError(err) {
		return
	}
	defer func(ev *Event) {
		a.NoError(ev.Close())
	}(ev)
	ev, err = NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.Error(err) {
		ev.Destroy()
	}
}

func TestEventOpenMode5(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.NoError(err) {
		return
	}
	a.NoError(ev.Close())
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err = NewEvent(testEventName, 0, 0666, false)
	if !a.Error(err) {
		ev.Destroy()
	}
}

func TestEventWait(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.NoError(err) || !a.NotNil(ev) {
		return
	}
	defer func() {
		a.NoError(ev.Destroy())
	}()
	go func() {
		time.Sleep(time.Millisecond * 50)
		ev.Set()
	}()
	ch := make(chan struct{})
	go func() {
		ev.Wait()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Error("timeout")
	}
}

func TestEventWait2(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.NoError(err) || !a.NotNil(ev) {
		return
	}
	go func() {
		ev.Set()
	}()
	a.True(ev.WaitTimeout(time.Millisecond * 250))
	a.NoError(ev.Destroy())
}

func TestEventWait3(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.NoError(err) || !a.NotNil(ev) {
		return
	}
	a.False(ev.WaitTimeout(time.Millisecond * 50))
	a.NoError(ev.Destroy())
}

func TestEventWait4(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, true)
	if !a.NoError(err) || !a.NotNil(ev) {
		return
	}
	a.True(ev.WaitTimeout(0))
	a.NoError(ev.Destroy())
}

func TestEventWait5(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, true)
	if !a.NoError(err) || !a.NotNil(ev) {
		return
	}
	defer func() {
		a.NoError(ev.Destroy())
	}()
	var wg sync.WaitGroup
	for i := 0; i < 1024; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ev.Wait()
			ev.Set()
		}()
	}
	ev.Set()
	wg.Wait()
}

func TestEventWait6(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.NoError(err) || !a.NotNil(ev) {
		return
	}
	defer func() {
		a.NoError(ev.Destroy())
	}()
	a.False(ev.WaitTimeout(0))
	a.False(ev.WaitTimeout(0))
	ev.Set()
	a.True(ev.WaitTimeout(0))
}

func TestEventSetAnotherProcess(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.NoError(err) || !a.NotNil(ev) {
		return
	}
	defer func() {
		a.NoError(ev.Destroy())
	}()
	ch := make(chan struct{})
	go func() {
		ev.Wait()
		ch <- struct{}{}
	}()
	args := argsForEventSetCommand(testEventName)
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

func TestEventWaitAnotherProcess(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.NoError(err) || !a.NotNil(ev) {
		return
	}
	defer func() {
		a.NoError(ev.Destroy())
	}()
	args := argsForEventWaitCommand(testEventName, -1)
	killCh := make(chan bool, 1)
	resultCh := testutil.RunTestAppAsync(args, killCh)
	ev.Set()
	select {
	case result := <-resultCh:
		if !a.NoError(result.Err) {
			t.Logf("test app error. the output is: %s", result.Output)
		}
	case <-time.After(time.Second * 10):
		killCh <- true
		t.Errorf("timeout")
	}
}

func TestEventTimedWaitAnotherProcess(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroyEvent(testEventName)) {
		return
	}
	ev, err := NewEvent(testEventName, os.O_CREATE|os.O_EXCL, 0666, false)
	if !a.NoError(err) || !a.NotNil(ev) {
		return
	}
	defer func() {
		a.NoError(ev.Destroy())
	}()
	args := argsForEventWaitCommand(testEventName, 250)
	killCh := make(chan bool, 1)
	resultCh := testutil.RunTestAppAsync(args, killCh)
	select {
	case result := <-resultCh:
		if !a.Error(result.Err) {
			t.Logf("test app error was expected. the output is: %s", result.Output)
		} else if !a.True(strings.Contains(result.Output, "timeout exceeded")) {
			t.Logf("invalid error message: %s", result.Output)
		}
	case <-time.After(time.Second * 10):
		killCh <- true
		t.Errorf("timeout")
	}
}
