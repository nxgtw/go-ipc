// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import "os"

type RwMutex struct {
	*rwMutexImpl
}

// TODO(avd) - mode check
func NewRwMutex(name string, mode int, perm os.FileMode) (*RwMutex, error) {
	impl, err := newRwMutexImpl(name, mode, perm)
	if err != nil {
		return nil, err
	}
	result := &RwMutex{impl}
	return result, nil
}
