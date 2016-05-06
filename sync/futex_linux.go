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

type Futex struct {
	uaddr  unsafe.Pointer
	region *ipc.MemoryRegion
	name   string
}

// NewFutex creates a new futex.
// name - object name.
// mode - object creation mode. must be one of the following:
//		O_CREATE_ONLY
//		O_OPEN_ONLY
//		O_OPEN_OR_CREATE
//	perm - file's mode and permission bits.
//	initial - initial futex value. it is set only if the futex was created.
func NewFutex(name string, mode int, perm os.FileMode, initial uint32) (*Futex, error) {
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
	futex := &Futex{uaddr: allocator.ByteSliceData(region.Data()), region: region, name: name}
	if created {
		*(*uint32)(futex.uaddr) = initial
	}
	return futex, nil
}

func (f *Futex) Wait(value uint32, timeout time.Duration) error {
	fun := func() error {
		ptr := unsafe.Pointer(common.TimeoutToTimeSpec(timeout))
		_, err := futex(f.uaddr, cFUTEX_WAIT, value, ptr, nil, 0)
		return err
	}
	return common.UninterruptedSyscall(fun)
}

func (f *Futex) Wake(count uint32) (int, error) {
	var woken int32
	fun := func() error {
		var err error
		woken, err = futex(f.uaddr, cFUTEX_WAKE, count, nil, nil, 0)
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
func (f *Futex) Close() error {
	return f.region.Close()
}

// Destroy removes the futex object.
func (f *Futex) Destroy() error {
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
