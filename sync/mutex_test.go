// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"testing"
)

func mutexCtor(name string, flag int, perm os.FileMode) (IPCLocker, error) {
	return NewMutex(name, flag, perm)
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

func TestMutexOpenMode5(t *testing.T) {
	testLockerOpenMode5(t, mutexCtor, mutexDtor)
}

func TestMutexLock(t *testing.T) {
	testLockerLock(t, mutexCtor, mutexDtor)
}

func TestMutexMemory(t *testing.T) {
	testLockerMemory(t, defaultMutexType, mutexCtor, mutexDtor)
}

func TestMutexValueInc(t *testing.T) {
	testLockerValueInc(t, defaultMutexType, mutexCtor, mutexDtor)
}

func TestMutexLockTimeout(t *testing.T) {
	testLockerLockTimeout(t, defaultMutexType, mutexCtor, mutexDtor)
}

func TestMutexLockTimeout2(t *testing.T) {
	testLockerLockTimeout2(t, defaultMutexType, mutexCtor, mutexDtor)
}

func TestMutexPanicsOnDoubleUnlock(t *testing.T) {
	testLockerTwiceUnlock(t, mutexCtor, mutexDtor)
}
