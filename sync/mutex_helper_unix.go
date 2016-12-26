// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin linux,sysv_mutex_linux freebsd,sysv_mutex_freebsd

package sync

import "os"

// this is to ensure, that all implementations of ipc mutex
// satisfy the same minimal interface
var (
	_ TimedIPCLocker = (*SemaMutex)(nil)
)

func newMutex(name string, flag int, perm os.FileMode) (TimedIPCLocker, error) {
	l, err := NewSemaMutex(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func destroyMutex(name string) error {
	return DestroySemaMutex(name)
}
