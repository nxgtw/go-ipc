// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"os"
	"sync"
	"unsafe"
)

type rwMutexImpl struct {
	m      *sync.RWMutex
	region *MemoryRegion
	name   string
}

func newRwMutexImpl(name string, mode int, perm os.FileMode) (*rwMutexImpl, error) {
	name = "go-ipc.rwm." + name
	obj, err := NewMemoryObject(name, mode, perm)
	if err != nil {
		return nil, err
	}
	size := unsafe.Sizeof(sync.RWMutex{})
	if err := obj.Truncate(int64(size)); err != nil {
		obj.Destroy()
		return nil, err
	}
	region, err := NewMemoryRegion(obj, SHM_READWRITE, 0, int(size))
	if err != nil {
		obj.Destroy()
		return nil, err
	}
	if err := alloc(region.Data(), sync.RWMutex{}); err != nil {
		return nil, err
	}
	m := (*sync.RWMutex)(unsafe.Pointer(byteSliceToUintPtr(region.Data())))
	return &rwMutexImpl{m, region, name}, nil
}

func (rw *rwMutexImpl) Lock() {
	rw.m.Lock()
}

func (rw *rwMutexImpl) RLock() {
	rw.m.RLock()
}

func (rw *rwMutexImpl) RLocker() sync.Locker {
	return rw.m.RLocker()
}

func (rw *rwMutexImpl) RUnlock() {
	rw.m.RUnlock()
}

func (rw *rwMutexImpl) Unlock() {
	rw.m.Unlock()
}

func (rw *rwMutexImpl) Destroy() {
	rw.m = nil
	rw.region.Close()
	DestroyMemoryObject(rw.name)
}
