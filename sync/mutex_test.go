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

func mutexDtor(name string) error {
	return DestroyMutex(name)
}

func TestMutexOpenMode(t *testing.T) {
	testLockerOpenMode(t, mutexCtor, mutexDtor)
}

func TestMutexOpenMode2(t *testing.T) {
	testLockerOpenMode2(t, mutexCtor, mutexDtor)
}

func TestMutexOpenMode3(t *testing.T) {
	testLockerOpenMode3(t, mutexCtor, mutexDtor)
}

func TestMutexOpenMode4(t *testing.T) {
	testLockerOpenMode4(t, mutexCtor, mutexDtor)
}

func TestMutexLock(t *testing.T) {
	testLockerLock(t, mutexCtor, mutexDtor)
}

func TestMutexMemory(t *testing.T) {
	testLockerMemory(t, "m", mutexCtor, mutexDtor)
}

func TestMutexValueInc(t *testing.T) {
	testLockerValueInc(t, "m", mutexCtor, mutexDtor)
}
