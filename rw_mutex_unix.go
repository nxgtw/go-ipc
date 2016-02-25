// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build ignore

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

type internalRwMutex sync.RWMutex

// used for rwMutexImpl.RLocker().
// we can't use internal mutex's rlocker, as
// it may be used longer, then the rwMutexImpl object exists,
// which can be garbage collected and
// its shared memory region be removed
type rlocker rwMutexImpl

func (r *rlocker) Lock()   { (*rwMutexImpl)(r).RLock() }
func (r *rlocker) Unlock() { (*rwMutexImpl)(r).RUnlock() }

func newRwMutexImpl(name string, mode int, perm os.FileMode) (impl *rwMutexImpl, resultErr error) {
	name = rwmName(name)
	var obj *MemoryObject
	if obj, resultErr = NewMemoryObject(name, mode|O_READWRITE, perm); resultErr != nil {
		return
	}
	var region *MemoryRegion
	defer func() {
		if resultErr == nil {
			return
		}
		if region != nil {
			region.Close()
		}
		if obj != nil {
			if mode == O_CREATE_ONLY {
				obj.Destroy()
			}
		}
	}()
	size := unsafe.Sizeof(internalRwMutex{})
	if resultErr = obj.Truncate(int64(size)); resultErr != nil {
		return
	}
	if region, resultErr = NewMemoryRegion(obj, SHM_READWRITE, 0, int(size)); resultErr != nil {
		return
	}
	if mode == O_CREATE_ONLY {
		if resultErr = alloc(region.Data(), sync.RWMutex{}); resultErr != nil {
			return
		}
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
	rw.region.Close()
	rw.region = nil
	name := rw.name
	rw.name = ""
	return DestroyMemoryObject(name)
}

// DestroyRwMutex permanently removes the mutex
func DestroyRwMutex(name string) error {
	return DestroyMemoryObject(rwmName(name))
}

func rwmName(name string) string {
	return "go-ipc.rwm." + name
}
