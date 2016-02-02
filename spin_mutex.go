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

// NewSpinMutex creates a new spinmutex
// name - object name
// mode - object creation mode. must be one of the following:
//	TODO(avd) - O_OPEN_OR_CREATE
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
	defer obj.Close()
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
	impl = &SpinMutex{m, region, name}
	return
}

// Finish indicates, that the object is no longer in use,
// and that the underlying resources can be freed
func (rw *SpinMutex) Finish() error {
	return rw.region.Close()
}

func (rw *SpinMutex) Destroy() error {
	if err := rw.Finish(); err != nil {
		return err
	}
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
