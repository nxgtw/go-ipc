// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"

	"github.com/pkg/errors"
)

// IPCFutex is a linux futex, placed into a shared memory region.
type IPCFutex struct {
	futex  *Futex
	region *mmf.MemoryRegion
	name   string
}

// NewIPCFutex creates a new futex, placing it in a shared memory region with the given name.
//	name - shared memory region name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
//	initial - initial futex value. it is set only if the futex was created.
func NewIPCFutex(name string, flag int, perm os.FileMode, initial uint32) (*IPCFutex, error) {
	if !checkMutexFlags(flag) {
		return nil, errors.New("invalid open flags")
	}
	name = futexName(name)
	obj, created, resultErr := newMemoryObjectSize(name, flag, perm, futexSize)
	if resultErr != nil {
		return nil, errors.Wrap(resultErr, "failed to create shm object")
	}
	var region *mmf.MemoryRegion
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
	if region, resultErr = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, int(futexSize)); resultErr != nil {
		return nil, errors.Wrap(resultErr, "failed to create shm region")
	}
	result := &IPCFutex{
		futex:  NewFutex(allocator.ByteSliceData(region.Data())),
		region: region, name: name,
	}
	if created {
		*result.Addr() = initial
	}
	return result, nil
}

// Addr returns address of the futex's value.
func (f *IPCFutex) Addr() *uint32 {
	return f.futex.Addr()
}

// Wait checks if the the value equals futex's value.
// If it doesn't, Wait returns EWOULDBLOCK.
// Otherwise, it waits for the Wake call on the futex for not longer, than timeout.
func (f *IPCFutex) Wait(value uint32, timeout time.Duration) error {
	return f.futex.Wait(value, timeout, 0)
}

// Wake wakes count threads waiting on the futex.
func (f *IPCFutex) Wake(count uint32) (int, error) {
	return f.futex.Wake(count, 0)
}

// Close indicates, that the object is no longer in use,
// and that the underlying resources can be freed.
func (f *IPCFutex) Close() error {
	return f.region.Close()
}

// Destroy removes the futex object.
func (f *IPCFutex) Destroy() error {
	if err := f.Close(); err != nil {
		return errors.Wrap(err, "failed to close shm region")
	}
	f.region = nil
	f.futex = nil
	err := shm.DestroyMemoryObject(f.name)
	f.name = ""
	return err
}

func futexName(name string) string {
	return "go-ipc.futex." + name
}
