// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build !sysv_mutex_linux

package sync

import "os"

// this is to ensure, that all implementations of ipc mutex
// satisfy the same minimal interface
var (
	_ IPCLocker = (*FutexMutex)(nil)
)

func newMutex(name string, flag int, perm os.FileMode) (IPCLocker, error) {
	l, err := NewFutexMutex(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func destroyMutex(name string) error {
	return DestroyFutexMutex(name)
}
