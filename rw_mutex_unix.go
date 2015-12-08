// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"os"
	"sync"
	"unsafe"
)

type rwMutexImpl struct {
	m sync.RWMutex
}

func newRwMutexImpl(name string, mode int, perm os.FileMode) (*rwMutexImpl, error) {
	obj, err := NewMemoryObject("go-ipc.rwm."+name, mode, perm)
	if err != nil {
		return nil, err
	}
	region, err := NewMemoryRegion(obj, mode, 0, int(unsafe.Sizeof(rwMutexImpl{})))
	if err != nil {
		obj.Destroy()
		return nil, err
	}
	_ = region
	// TODO(avd) - finish this
	return &rwMutexImpl{}, nil
}
