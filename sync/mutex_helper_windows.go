// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import "os"

// this is to ensure, that all implementations of ipc mutex satisfy the same minimal interface.
var (
	_ TimedIPCLocker = (*EventMutex)(nil)
)

func newMutex(name string, flag int, perm os.FileMode) (TimedIPCLocker, error) {
	l, err := NewEventMutex(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func destroyMutex(name string) error {
	return DestroyEventMutex(name)
}
