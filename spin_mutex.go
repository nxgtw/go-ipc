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

// SpinMutex is a synchronization object which performs busy wait loop
type SpinMutex struct {
	*spinMutexImpl
	region *MemoryRegion
	name   string
}

// NewSpinMutex creates a new spinmutex
// name - object name.
// mode - object creation mode. must be one of the following:
//  O_CREATE_ONLY
//  O_OPEN_ONLY
//  O_OPEN_OR_CREATE
func NewSpinMutex(name string, mode int, perm os.FileMode) (*SpinMutex, error) {
	if !checkMutexOpenMode(mode) {
		return nil, fmt.Errorf("invalid open mode")
	}
	return newSpinMutex(name, mode, perm)
}

func newSpinMutex(name string, mode int, perm os.FileMode) (*SpinMutex, error) {
	const spinImplSize = int64(unsafe.Sizeof(spinMutexImpl{}))
	name = spinName(name)
	obj, created, resultErr := createMemoryObject(name, mode|O_READWRITE, perm)
	if resultErr != nil {
		return nil, resultErr
	}
	var region *MemoryRegion
	defer func() {
		obj.Close()
		if resultErr == nil {
			return
		}
		if region != nil {
			region.Close()
		}
		if created {
			obj.Destroy()
		}
	}()
	if created {
		if resultErr = obj.Truncate(int64(spinImplSize)); resultErr != nil {
			return nil, resultErr
		}
	} else {
		if obj.Size() < spinImplSize {
			return nil, fmt.Errorf("existing object has invalid size %d", obj.Size())
		}
	}
	if region, resultErr = NewMemoryRegion(obj, MEM_READWRITE, 0, int(spinImplSize)); resultErr != nil {
		return nil, resultErr
	}
	if created {
		if resultErr = alloc(region.Data(), spinMutexImpl{}); resultErr != nil {
			return nil, resultErr
		}
	}
	m := (*spinMutexImpl)(unsafe.Pointer(&region.data[0]))
	impl := &SpinMutex{m, region, name}
	return impl, nil
}

// Finish indicates, that the object is no longer in use,
// and that the underlying resources can be freed
func (spin *SpinMutex) Close() error {
	return spin.region.Close()
}

// Destroy removes the mutex object
func (spin *SpinMutex) Destroy() error {
	if err := spin.Close(); err != nil {
		return err
	}
	spin.region = nil
	err := DestroyMemoryObject(spin.name)
	spin.name = ""
	return err
}

// DestroySpinMutex removes the mutex object with a given name
func DestroySpinMutex(name string) error {
	return DestroyMemoryObject(spinName(name))
}

func spinName(name string) string {
	return "go-ipc.spin." + name
}

func createMemoryObject(name string, mode int, perm os.FileMode) (obj *MemoryObject, created bool, err error) {
	switch {
	case mode&(O_OPEN_ONLY|O_CREATE_ONLY) != 0:
		obj, err = NewMemoryObject(name, mode, perm)
		if err == nil && (mode&O_CREATE_ONLY) != 0 {
			created = true
		}
	case mode&O_OPEN_OR_CREATE != 0:
		const attempts = 16
		mode = mode & ^(O_OPEN_OR_CREATE)
		for attempt := 0; attempt < attempts; attempt++ {
			if obj, err = NewMemoryObject(name, mode|O_CREATE_ONLY, perm); !os.IsExist(err) {
				created = true
				break
			} else {
				if obj, err = NewMemoryObject(name, mode|O_OPEN_ONLY, perm); !os.IsNotExist(err) {
					break
				}
			}
		}
	default:
		err = fmt.Errorf("invalid open mode")
	}
	return
}
