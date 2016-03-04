// Copyright 2015 Aleksandr Demakin. All rights reserved.
// ignore this for a while, as linux rw mutexes don't work,
// and windows mutexes are not ready yes.

// +build windows

package sync

import (
	"os"
	"testing"
)

func mutexCtor(name string, mode int, perm os.FileMode) (IPCLocker, error) {
	return NewMutex(name, mode, perm)
}

/*
func mutexDtor(name string) error {
	return DestroySpinMutex(name)
}*/

func TestMutexOpenMode(t *testing.T) {
	testLockerOpenMode(t, mutexCtor, nil)
}

func TestMutexOpenMode2(t *testing.T) {
	testLockerOpenMode2(t, mutexCtor, nil)
}

func TestMutexOpenMode3(t *testing.T) {
	testLockerOpenMode3(t, mutexCtor, nil)
}

func TestMutexOpenMode4(t *testing.T) {
	testLockerOpenMode4(t, mutexCtor, nil)
}

func TestMutexLock(t *testing.T) {
	testLockerLock(t, mutexCtor, nil)
}
