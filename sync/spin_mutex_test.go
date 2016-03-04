// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"testing"
)

const testSpinMutexName = "spin-test"

func spinCtor(name string, mode int, perm os.FileMode) (IPCLocker, error) {
	return NewSpinMutex(name, mode, perm)
}

func spinDtor(name string) error {
	return DestroySpinMutex(name)
}

func TestSpinMutexOpenMode(t *testing.T) {
	testLockerOpenMode(t, spinCtor, spinDtor)
}

func TestSpinMutexOpenMode2(t *testing.T) {
	testLockerOpenMode2(t, spinCtor, spinDtor)
}

func TestSpinMutexOpenMode3(t *testing.T) {
	testLockerOpenMode3(t, spinCtor, spinDtor)
}

func TestSpinMutexOpenMode4(t *testing.T) {
	testLockerOpenMode4(t, spinCtor, spinDtor)
}

func TestSpinMutexLock(t *testing.T) {
	testLockerLock(t, spinCtor, spinDtor)
}

func TestSpinMutexMemory(t *testing.T) {
	testLockerMemory(t, spinCtor, spinDtor)
}

func TestSpinMutexValueInc(t *testing.T) {
	testLockerValueInc(t, spinCtor, spinDtor)
}
