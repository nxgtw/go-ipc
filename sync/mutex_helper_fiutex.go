// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux,!sysv_mutex_linux freebsd,!sysv_mutex_freebsd

package sync

import "os"

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
