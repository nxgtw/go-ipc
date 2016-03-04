// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"os"
)

// Mutex is an interprocess readwrite lock object
type Mutex struct {
	*mutexImpl
}

// NewMutex creates a new readwrite mutex
// name - object name
// mode - object creation mode. must be one of the following:
//	O_OPEN_OR_CREATE
//	O_CREATE_ONLY
//	O_OPEN_ONLY
func NewMutex(name string, mode int, perm os.FileMode) (*Mutex, error) {
	if !checkMutexOpenMode(mode) {
		return nil, fmt.Errorf("invalid open mode")
	}
	impl, err := newMutexImpl(name, mode, perm)
	if err != nil {
		return nil, err
	}
	result := &Mutex{impl}
	return result, nil
}
