// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"unsafe"

	ipc "bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"
	"bitbucket.org/avd/go-ipc/shm"
)

type spinMutex struct {
	value uint32
}

func (spin *spinMutex) Lock() {
	for !spin.TryLock() {
		runtime.Gosched()
	}
}

func (spin *spinMutex) Unlock() {
	atomic.StoreUint32(&spin.value, 0)
}

func (spin *spinMutex) TryLock() bool {
	return atomic.CompareAndSwapUint32(&spin.value, 0, 1)
}

// SpinMutex is a synchronization object which performs busy wait loop
type SpinMutex struct {
	*spinMutex
	region *ipc.MemoryRegion
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
	const spinImplSize = int64(unsafe.Sizeof(spinMutex{}))
	name = spinName(name)
	var obj *shm.MemoryObject
	creator := func(create bool) error {
		var err error
		creatorMode := ipc.O_READWRITE
		if create {
			creatorMode |= ipc.O_CREATE_ONLY
		} else {
			creatorMode |= ipc.O_OPEN_ONLY
		}
		obj, err = shm.NewMemoryObject(name, creatorMode, perm)
		return err
	}
	created, resultErr := common.OpenOrCreate(creator, mode)
	if resultErr != nil {
		return nil, resultErr
	}
	var region *ipc.MemoryRegion
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
		if resultErr = obj.Truncate(spinImplSize); resultErr != nil {
			return nil, resultErr
		}
	} else if obj.Size() < spinImplSize {
		return nil, fmt.Errorf("existing object has invalid size %d", obj.Size())
	}
	if region, resultErr = ipc.NewMemoryRegion(obj, ipc.MEM_READWRITE, 0, int(spinImplSize)); resultErr != nil {
		return nil, resultErr
	}
	if created {
		if resultErr = allocator.Alloc(region.Data(), spinMutex{}); resultErr != nil {
			return nil, resultErr
		}
	}
	m := (*spinMutex)(allocator.ByteSliceData(region.Data()))
	impl := &SpinMutex{m, region, name}
	return impl, nil
}

// Close indicates, that the object is no longer in use,
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
	err := shm.DestroyMemoryObject(spin.name)
	spin.name = ""
	return err
}

// DestroySpinMutex removes the mutex object with a given name
func DestroySpinMutex(name string) error {
	return shm.DestroyMemoryObject(spinName(name))
}

func spinName(name string) string {
	return "go-ipc.spin." + name
}
