// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"os"
)

// NewMutex creates a new interprocess mutex.
// It uses the default implementation on the current platform.
//	name - object name
//	mode - object creation mode. must be one of the following:
//		O_OPEN_OR_CREATE
//		O_CREATE_ONLY
//		O_OPEN_ONLY
//	perm - object permissions
func NewMutex(name string, mode int, perm os.FileMode) (IPCLocker, error) {
	if !checkMutexOpenMode(mode) {
		return nil, fmt.Errorf("invalid open mode")
	}
	return newMutex(name, mode, perm)
}

func DestroyMutex(name string) error {
	return destroyMutex(name)
}
