// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build !sysv_mutex_linux

package sync

import "os"

// this is to ensure, that all implementations of ipc mutex
// satisfy the same minimal interface
var (
	_ IPCLocker = (*FutexMutex)(nil)
)

func newMutex(name string, mode int, perm os.FileMode) (IPCLocker, error) {
	return NewFutexMutex(name, mode, perm)
}

func destroyMutex(name string) error {
	return DestroyFutexMutex(name)
}
