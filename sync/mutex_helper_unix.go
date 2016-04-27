// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux

package sync

import "os"

// this is to ensure, that all implementations of ipc mutex
// satisfy the same minimal interface
var (
	_ IPCLocker = (*SemaMutex)(nil)
)

func newMutex(name string, mode int, perm os.FileMode) (IPCLocker, error) {
	return NewSemaMutex(name, mode, perm)
}

func destroyMutex(name string) error {
	return DestroySemaMutex(name)
}
