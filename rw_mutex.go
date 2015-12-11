// Copyright 2015 Aleksandr Demakin. All rights reserved.

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
	if !checkRwMutexOpenMode(mode) {
		return nil, fmt.Errorf("invalid open mode")
	}
	impl, err := newRwMutexImpl(name, mode, perm)
	if err != nil {
		return nil, err
	}
	result := &RwMutex{impl}
	return result, nil
}

func checkRwMutexOpenMode(mode int) bool {
	return /*mode == O_OPEN_OR_CREATE ||*/ mode == O_CREATE_ONLY || mode == O_OPEN_ONLY
}
