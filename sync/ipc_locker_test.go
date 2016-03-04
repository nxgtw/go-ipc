// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"sync"
	"testing"

	ipc "bitbucket.org/avd/go-ipc"

	"github.com/stretchr/testify/assert"
)

const (
	testLockerName = "go-ipc.locker"
)

type lockerCtor func(name string, mode int, perm os.FileMode) (IPCLocker, error)
type lockerDtor func(name string) error

func testLockerOpenMode(t *testing.T, ctor lockerCtor, dtor lockerDtor) bool {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return false
		}
	}
	lk, err := ctor(testLockerName, ipc.O_READWRITE, 0666)
	a.Error(err)
	lk, err = ctor(testLockerName, ipc.O_CREATE_ONLY|ipc.O_READ_ONLY, 0666)
	a.Error(err)
	lk, err = ctor(testLockerName, ipc.O_OPEN_OR_CREATE|ipc.O_WRITE_ONLY, 0666)
	a.Error(err)
	lk, err = ctor(testLockerName, ipc.O_OPEN_ONLY|ipc.O_WRITE_ONLY, 0666)
	a.Error(err)
	lk, err = ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.NoError(err) {
		return false
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	lk, err = ctor(testLockerName, ipc.O_OPEN_ONLY, 0666)
	if !a.NoError(err) {
		return false
	}
	a.NoError(lk.Close())
	return true
}

func testLockerOpenMode2(t *testing.T, ctor lockerCtor, dtor lockerDtor) bool {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return false
		}
	}
	lk, err := ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.NoError(err) {
		return false
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	lk, err = ctor(testLockerName, ipc.O_OPEN_ONLY, 0666)
	if !a.NoError(err) {
		return false
	}
	defer func(lk IPCLocker) {
		a.NoError(lk.Close())
	}(lk)
	lk, err = ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.Error(err) {
		return false
	}
	return true
}

func testLockerOpenMode3(t *testing.T, ctor lockerCtor, dtor lockerDtor) bool {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return false
		}
	}
	lk, err := ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return false
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	lk, err = ctor(testLockerName, ipc.O_OPEN_OR_CREATE, 0666)
	if !a.NoError(err) {
		return false
	}
	a.NoError(lk.Close())
	return true
}

func testLockerOpenMode4(t *testing.T, ctor lockerCtor, dtor lockerDtor) bool {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return false
		}
	}
	lk, err := ctor(testLockerName, ipc.O_OPEN_OR_CREATE, 0666)
	if !a.NoError(err) {
		return false
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	lk, err = ctor(testLockerName, ipc.O_OPEN_ONLY, 0666)
	if !a.NoError(err) {
		return false
	}
	defer func(lk IPCLocker) {
		a.NoError(lk.Close())
	}(lk)
	lk, err = ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.Error(err) {
		return false
	}
	return true
}

func testLockerLock(t *testing.T, ctor lockerCtor, dtor lockerDtor) bool {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return false
		}
	}
	lk, err := ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.NoError(err) || !a.NotNil(lk) {
		return false
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	var wg sync.WaitGroup
	sharedValue := 0
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			lk.Lock()
			for i := 0; i < 1000; i++ {
				sharedValue++
			}
			lk.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
	a.Equal(30000, sharedValue)
	return true
}
