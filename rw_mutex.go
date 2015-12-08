// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import "os"

type RwMutex struct {
	*rwMutexImpl
}

func NewRwMutex(name string, mode int, perm os.FileMode) (*RwMutex, error) {
	return nil, nil
}
