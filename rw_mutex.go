// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build ignore

package ipc

import (
	"fmt"
	"os"
)

type RwMutex struct {
	*rwMutexImpl
}

// creates a new rwmutex
// name - object name
// mode - object creation mode. must be one of the following:
//	O_OPEN_OR_CREATE
//	O_CREATE_ONLY
//	O_OPEN_ONLY
func NewRwMutex(name string, mode int, perm os.FileMode) (*RwMutex, error) {
	if !checkMutexOpenMode(mode) {
		return nil, fmt.Errorf("invalid open mode")
	}
	impl, err := newRwMutexImpl(name, mode, perm)
	if err != nil {
		return nil, err
	}
	result := &RwMutex{impl}
	return result, nil
}
