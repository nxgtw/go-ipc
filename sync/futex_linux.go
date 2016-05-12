// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"os"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"
	"bitbucket.org/avd/go-ipc/shm"
)

const (
	futexSize = 4
)

// IPCFutex is a linux ipc mechanism, which can be used to implement
// different synchronization objects.
type IPCFutex struct {
	uaddr  unsafe.Pointer
	region *ipc.MemoryRegion
	name   string
	flags  int32
}

// NewIPCFutex creates a new futex, placing it in a shared memory region with the given name.
// name - shared memory region name.
// mode - object creation mode. must be one of the following:
//		O_CREATE_ONLY
//		O_OPEN_ONLY
//		O_OPEN_OR_CREATE
//	perm - file's mode and permission bits.
//	initial - initial futex value. it is set only if the futex was created.
//	flags - OR'ed combination of FUTEX_PRIVATE_FLAG and FUTEX_CLOCK_REALTIME.
func NewIPCFutex(name string, mode int, perm os.FileMode, initial uint32, flags int32) (*IPCFutex, error) {
	if !checkMutexOpenMode(mode) {
		return nil, fmt.Errorf("invalid open mode")
	}
	name = futexName(name)
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
		if resultErr = obj.Truncate(futexSize); resultErr != nil {
			return nil, resultErr
		}
	} else if obj.Size() < futexSize {
		return nil, fmt.Errorf("existing object has invalid size %d", obj.Size())
	}
	if region, resultErr = ipc.NewMemoryRegion(obj, ipc.MEM_READWRITE, 0, int(futexSize)); resultErr != nil {
		return nil, resultErr
	}
	futex := &IPCFutex{uaddr: allocator.ByteSliceData(region.Data()), region: region, name: name}
	if created {
		*(*uint32)(futex.uaddr) = initial
	}
	return futex, nil
}

// Addr returns address of the futex's value.
func (f *IPCFutex) Addr() *uint32 {
	return (*uint32)(f.uaddr)
}

// Wait checks if the the value equals futex's value.
// If it doesn't, Wait returns EWOULDBLOCK.
// Otherwise, it waits for the Wake call on the futex for not longer, than timeout.
func (f *IPCFutex) Wait(value uint32, timeout time.Duration) error {
	ptr := unsafe.Pointer(common.TimeoutToTimeSpec(timeout))
	fun := func() error {
		_, err := futex(f.uaddr, cFUTEX_WAIT|f.flags, value, ptr, nil, 0)
		return err
	}
	return common.UninterruptedSyscall(fun)
}

// Wake wakes count threads waiting on the futex.
func (f *IPCFutex) Wake(count uint32) (int, error) {
	var woken int32
	fun := func() error {
		var err error
		woken, err = futex(f.uaddr, cFUTEX_WAKE|f.flags, count, nil, nil, 0)
		return err
	}
	err := common.UninterruptedSyscall(fun)
	if err == nil {
		return int(woken), nil
	}
	return 0, err
}

// Close indicates, that the object is no longer in use,
// and that the underlying resources can be freed.
func (f *IPCFutex) Close() error {
	return f.region.Close()
}

// Destroy removes the futex object.
func (f *IPCFutex) Destroy() error {
	if err := f.Close(); err != nil {
		return err
	}
	f.region = nil
	err := shm.DestroyMemoryObject(f.name)
	f.name = ""
	return err
}

func futexName(name string) string {
	return "go-ipc.futex." + name
}
