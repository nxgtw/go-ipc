// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"unsafe"
)

type spinMutexImpl struct {
	value uint32
}

func (impl *spinMutexImpl) Lock() {
	for !impl.TryLock() {
		runtime.Gosched()
	}
}

func (impl *spinMutexImpl) Unlock() {
	atomic.StoreUint32(&impl.value, 0)
}

func (impl *spinMutexImpl) TryLock() bool {
	return atomic.CompareAndSwapUint32(&impl.value, 0, 1)
}

type SpinMutex struct {
	*spinMutexImpl
	region *MemoryRegion
	name   string
}

// creates a new rwmutex
// name - object name
// mode - object creation mode. must be one of the following:
//	O_CREATE_ONLY
//	O_OPEN_ONLY
func NewSpinMutex(name string, mode int, perm os.FileMode) (*SpinMutex, error) {
	if !checkMutexOpenMode(mode) {
		return nil, fmt.Errorf("invalid open mode")
	}
	return newSpinMutex(name, mode, perm)
}

func newSpinMutex(name string, mode int, perm os.FileMode) (impl *SpinMutex, resultErr error) {
	var obj *MemoryObject
	name = spinName(name)
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
	size := unsafe.Sizeof(spinMutexImpl{})
	if resultErr = obj.Truncate(int64(size)); resultErr != nil {
		return
	}
	if region, resultErr = NewMemoryRegion(obj, MEM_READWRITE, 0, int(size)); resultErr != nil {
		return
	}
	if mode == O_CREATE_ONLY {
		if resultErr = alloc(region.Data(), spinMutexImpl{}); resultErr != nil {
			return
		}
	}
	m := (*spinMutexImpl)(unsafe.Pointer(&region.data[0]))
	return &SpinMutex{m, region, name}, nil
}

func (rw *SpinMutex) Destroy() error {
	rw.region.Close()
	rw.region = nil
	name := rw.name
	rw.name = ""
	return DestroyMemoryObject(name)
}

func DestroySpinMutex(name string) error {
	return DestroyMemoryObject(spinName(name))
}

func spinName(name string) string {
	return "go-ipc.spin." + name
}
