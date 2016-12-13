// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

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
	defer func() {
		a.NoError(DestroySemaphore(testSemaName))
	}()
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
