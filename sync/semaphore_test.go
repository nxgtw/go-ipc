// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"bitbucket.org/avd/go-ipc/internal/test"

	"github.com/stretchr/testify/assert"
)

const (
	testSemaName = "testsema"
)

func TestSemaOpenMode(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_RDWR, 0666, 1)
	a.Error(err)
	s, err = NewSemaphore(testSemaName, os.O_WRONLY, 0666, 1)
	a.Error(err)
	s, err = NewSemaphore(testSemaName, os.O_CREATE, 0666, 1)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
		a.NoError(DestroySemaphore(testSemaName))
	}(s)
	s, err = NewSemaphore(testSemaName, 0, 0666, 1)
	if !a.NoError(err) {
		return
	}
	a.NoError(s.Close())
}

func TestSemaOpenMode2(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 1)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
		a.NoError(DestroySemaphore(testSemaName))
	}(s)
	s, err = NewSemaphore(testSemaName, 0, 0666, 1)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
	}(s)
	s, err = NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 1)
	if !a.Error(err) {
		s.Close()
		DestroySemaphore(testSemaName)
	}
}

func TestSemaOpenMode3(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 1)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
		a.NoError(DestroySemaphore(testSemaName))
	}(s)
	s, err = NewSemaphore(testSemaName, os.O_CREATE, 0666, 1)
	if !a.NoError(err) {
		return
	}
	a.NoError(s.Close())
}

func TestSemaOpenMode4(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_CREATE, 0666, 1)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
		a.NoError(DestroySemaphore(testSemaName))
	}(s)
	s, err = NewSemaphore(testSemaName, 0, 0666, 1)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
	}(s)
	s, err = NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 1)
	if !a.Error(err) {
		s.Close()
		DestroySemaphore(testSemaName)
	}
}

func TestSemaOpenMode5(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 1)
	if !a.NoError(err) {
		return
	}
	a.NoError(s.Close())
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err = NewSemaphore(testSemaName, 0, 0666, 1)
	if !a.Error(err) {
		s.Close()
		DestroySemaphore(testSemaName)
	}
}

func TestSemaCount(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 16)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
		a.NoError(DestroySemaphore(testSemaName))
	}(s)
	var wg sync.WaitGroup
	wg.Add(16)
	for i := 0; i < 16; i++ {
		go func() {
			s.Wait()
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestSemaCount2(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 0)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
		a.NoError(DestroySemaphore(testSemaName))
	}(s)
	var wg sync.WaitGroup
	wg.Add(16)
	for i := 0; i < 16; i++ {
		go func() {
			s.Wait()
			wg.Done()
		}()
		s.Signal(1)
	}
	wg.Wait()
}

func TestTimedSema(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 1)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
		a.NoError(DestroySemaphore(testSemaName))
	}(s)
	ts, ok := s.(TimedSemaphore)
	if !ok {
		t.Skipf("semaphore on %s aren't timed", runtime.GOARCH)
		return
	}
	ts.Wait()
	a.False(ts.WaitTimeout(time.Millisecond * 50))
}

func TestSemaSignalAnotherProcess(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 0)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
		a.NoError(DestroySemaphore(testSemaName))
	}(s)
	ch := make(chan struct{})
	go func() {
		s.Wait()
		s.Wait()
		ch <- struct{}{}
	}()
	args := argsForSemaSignalCommand(testSemaName, 2)
	result := testutil.RunTestApp(args, nil)
	if !a.NoError(result.Err) {
		t.Logf("test app error. the output is: %s", result.Output)
		return
	}
	select {
	case <-ch:
	case <-time.After(time.Second * 3):
		t.Errorf("timeout")
	}
}

func TestSemaWaitAnotherProcess(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 0)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
		a.NoError(DestroySemaphore(testSemaName))
	}(s)
	args := argsForSemaWaitCommand(testSemaName, -1)
	killCh := make(chan bool, 1)
	resultCh := testutil.RunTestAppAsync(args, killCh)
	s.Signal(1)
	select {
	case result := <-resultCh:
		if !a.NoError(result.Err) {
			t.Logf("test app error. the output is: %s", result.Output)
		}
	case <-time.After(time.Second * 3):
		killCh <- true
		t.Errorf("timeout")
	}
}

func TestSemaTimedWaitAnotherProcess(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(DestroySemaphore(testSemaName)) {
		return
	}
	s, err := NewSemaphore(testSemaName, os.O_CREATE|os.O_EXCL, 0666, 0)
	if !a.NoError(err) {
		return
	}
	defer func(s Semaphore) {
		a.NoError(s.Close())
		a.NoError(DestroySemaphore(testSemaName))
	}(s)
	ts, ok := s.(TimedSemaphore)
	if !ok {
		t.Skipf("semaphore on %s aren't timed", runtime.GOARCH)
		return
	}
	args := argsForSemaWaitCommand(testSemaName, 250)
	killCh := make(chan bool, 1)
	resultCh := testutil.RunTestAppAsync(args, killCh)
	select {
	case result := <-resultCh:
		if !a.Error(result.Err) {
			t.Logf("test app error was expected. the output is: %s", result.Output)
		} else if !a.True(strings.Contains(result.Output, "timeout exceeded")) {
			t.Logf("invalid error message: %s", result.Output)
		}
	case <-time.After(time.Second * 3):
		killCh <- true
		t.Errorf("timeout")
	}
}

func ExampleSemaphore() {
	// create new semaphore with initial count set to 3.
	DestroySemaphore("sema")
	sema, err := NewSemaphore("sema", os.O_CREATE|os.O_EXCL, 0666, 3)
	if err != nil {
		panic(err)
	}
	defer sema.Close()
	// in the following cycle we consume three units of the resource and won't block.
	for i := 0; i < 3; i++ {
		sema.Wait()
		fmt.Println("got one resource unit")
	}
	// the following two goroutines won't continue until we call Signal().
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			// open existing semaphore
			sema, err := NewSemaphore("sema", 0, 0666, 0)
			if err != nil {
				panic(err)
			}
			defer sema.Close()
			sema.Wait()
			fmt.Println("got one resource unit after waiting")
		}()
	}
	// wake up goroutines
	fmt.Println("waking up...")
	sema.Signal(2)
	wg.Wait()
	fmt.Println("done")
	// Output:
	// got one resource unit
	// got one resource unit
	// got one resource unit
	// waking up...
	// got one resource unit after waiting
	// got one resource unit after waiting
	// done
}
