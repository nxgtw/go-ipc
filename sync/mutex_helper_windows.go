// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build windows

package sync

import "os"

// this is to ensure, that all implementations of ipc mutex
// satisfy the same minimal interface
var (
	_ IPCLocker = (*EventMutex)(nil)
)

func newMutex(name string, flag int, perm os.FileMode) (IPCLocker, error) {
	l, err := NewEventMutex(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func destroyMutex(name string) error {
	return DestroyEventMutex(name)
}
