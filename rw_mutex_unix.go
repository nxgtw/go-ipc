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

// used for rwMutexImpl.RLocker().
// we can't use internal mutex's rlocker, as
// it may be used longer, then the rwMutexImpl object,
// which can be garbage collected and
// its shared memory region be removed
type rlocker rwMutexImpl

func (r *rlocker) Lock()   { (*rwMutexImpl)(r).RLock() }
func (r *rlocker) Unlock() { (*rwMutexImpl)(r).RUnlock() }

// TODO(avd) - handle errors more carefully!
func newRwMutexImpl(name string, mode int, perm os.FileMode) (*rwMutexImpl, error) {
	name = "go-ipc.rwm." + name
	obj, err := NewMemoryObject(name, mode|O_READWRITE, perm)
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
		region.Close()
		obj.Destroy()
		return nil, err
	}
	m := (*sync.RWMutex)(unsafe.Pointer(byteSliceAddress(region.Data())))
	return &rwMutexImpl{m, region, name}, nil
}

func (rw *rwMutexImpl) Lock() {
	rw.m.Lock()
}

func (rw *rwMutexImpl) RLock() {
	rw.m.RLock()
}

func (rw *rwMutexImpl) RLocker() sync.Locker {
	return (*rlocker)(rw)
}

func (rw *rwMutexImpl) RUnlock() {
	rw.m.RUnlock()
}

func (rw *rwMutexImpl) Unlock() {
	rw.m.Unlock()
}

func (rw *rwMutexImpl) Destroy() error {
	rw.m = nil
	if err := rw.region.Close(); err != nil {
		return err
	}
	rw.region = nil
	return DestroyMemoryObject(rw.name)
}

func DestroyRwMutex(name string) error {
	return DestroyMemoryObject(name)
}
