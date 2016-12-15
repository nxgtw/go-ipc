// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"testing"
)

func rwMutexCtor(name string, flag int, perm os.FileMode) (IPCLocker, error) {
	return NewRWMutex(name, flag, perm)
}

func rwMutexDtor(name string) error {
	return DestroyRWMutex(name)
}

func TestRWMutexOpenMode(t *testing.T) {
	testLockerOpenMode(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexOpenMode2(t *testing.T) {
	testLockerOpenMode2(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexOpenMode3(t *testing.T) {
	testLockerOpenMode3(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexOpenMode4(t *testing.T) {
	testLockerOpenMode4(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexOpenMode5(t *testing.T) {
	testLockerOpenMode5(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexLock(t *testing.T) {
	testLockerLock(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexPanicsOnDoubleUnlock(t *testing.T) {
	testLockerTwiceUnlock(t, rwMutexCtor, rwMutexDtor)
}
