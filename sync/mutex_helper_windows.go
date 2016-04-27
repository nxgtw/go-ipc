// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import "os"

// this is to ensure, that all implementations of ipc mutex
// satisfy the same minimal interface
var (
	_ IPCLocker = (*EventMutex)(nil)
)

func newMutex(name string, mode int, perm os.FileMode) (IPCLocker, error) {
	return NewEventMutex(name, mode, perm)
}

func destroyMutex(name string) error {
	return DestroyEventMutex(name)
}
