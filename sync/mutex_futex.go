// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux freebsd

package sync

import (
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"

	"github.com/pkg/errors"
)

// all implementations must satisfy at least IPCLocker interface.
var (
	_ TimedIPCLocker = (*FutexMutex)(nil)
)

// FutexMutex is a mutex based on linux futex object.
type FutexMutex struct {
	futex  *InplaceMutex
	region *mmf.MemoryRegion
	name   string
}

// NewFutexMutex creates a new futex-based mutex.
// This implementation is based on a paper 'Futexes Are Tricky' by Ulrich Drepper,
// this document can be found in 'docs' folder.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
func NewFutexMutex(name string, flag int, perm os.FileMode) (*FutexMutex, error) {
	if !checkMutexFlags(flag) {
		return nil, errors.New("invalid open flags")
	}
	name = futexName(name)
	obj, created, resultErr := shm.NewMemoryObjectSize(name, flag, perm, int64(inplaceMutexSize))
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
	if region, resultErr = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, inplaceMutexSize); resultErr != nil {
		return nil, errors.Wrap(resultErr, "failed to create shm region")
	}
	futex := NewInplaceMutex(allocator.ByteSliceData(region.Data()))
	if created {
		futex.Init()
	}
	return &FutexMutex{futex: futex, name: name, region: region}, nil
}

// Lock locks the mutex. It panics on an error.
func (f *FutexMutex) Lock() {
	f.futex.Lock()
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
func (f *FutexMutex) LockTimeout(timeout time.Duration) bool {
	return f.futex.LockTimeout(timeout)
}

// Unlock releases the mutex. It panics on an error.
func (f *FutexMutex) Unlock() {
	f.futex.Unlock()
}

// Close indicates, that the object is no longer in use,
// and that the underlying resources can be freed.
func (f *FutexMutex) Close() error {
	return f.region.Close()
}

// Destroy removes the mutex object.
func (f *FutexMutex) Destroy() error {
	if err := f.Close(); err != nil {
		return errors.Wrap(err, "failed to close shm region")
	}
	f.region = nil
	f.futex = nil
	err := shm.DestroyMemoryObject(f.name)
	f.name = ""
	return err
}

// DestroyFutexMutex permanently removes mutex with the given name.
func DestroyFutexMutex(name string) error {
	m, err := NewFutexMutex(name, 0, 0666)
	if err != nil {
		if os.IsNotExist(errors.Cause(err)) {
			return nil
		}
		return errors.Wrap(err, "failed to open mutex")
	}
	return m.Destroy()
}

func futexName(name string) string {
	return "go-ipc.futex." + name
}