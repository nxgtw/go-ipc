// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"testing"
)

func rwMutexCtor(name string, flag int, perm os.FileMode) (IPCLocker, error) {
	return NewRWMutex(name, flag, perm)
}

func rwRMutexCtor(name string, flag int, perm os.FileMode) (IPCLocker, error) {
	locker, err := NewRWMutex(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return locker.RLocker(), nil
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

func TestRWMutexMemory(t *testing.T) {
	testLockerMemory(t, "rw", false, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexMemory2(t *testing.T) {
	testLockerMemory(t, "rw", true, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexValueInc(t *testing.T) {
	testLockerValueInc(t, "rw", rwMutexCtor, rwMutexDtor)
}

func TestRWMutexPanicsOnDoubleUnlock(t *testing.T) {
	testLockerTwiceUnlock(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexPanicsOnDoubleRUnlock(t *testing.T) {
	testLockerTwiceUnlock(t, rwRMutexCtor, rwMutexDtor)
}
