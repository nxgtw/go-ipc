// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"sync/atomic"
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/unix"
)

const (
	cFutexMutexUnlocked          = 0
	cFutexMutexLockedNoWaiters   = 1
	cFutexMutexLockedHaveWaiters = 2
)

// FutexMutex is a mutex based on linux futex object.
type FutexMutex struct {
	futex *IPCFutex
}

// NewFutexMutex creates a new futex-based mutex.
// This implementation is based on a paper 'Futexes Are Tricky' by Ulrich Drepper,
// this document can be found in 'docs' folder.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
func NewFutexMutex(name string, flag int, perm os.FileMode) (*FutexMutex, error) {
	futex, err := NewIPCFutex(name, flag, perm, cFutexMutexUnlocked)
	if err != nil {
		return nil, err
	}
	return &FutexMutex{futex: futex}, nil
}

// Lock locks the mutex. It panics on an error.
func (f *FutexMutex) Lock() {
	if err := f.lockTimeout(-1); err != nil {
		panic(err)
	}
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
func (f *FutexMutex) LockTimeout(timeout time.Duration) bool {
	err := f.lockTimeout(timeout)
	if err == nil {
		return true
	}
	if common.IsTimeoutErr(err) {
		return false
	}
	panic(err)
}

// Unlock releases the mutex. It panics on an error.
func (f *FutexMutex) Unlock() {
	addr := f.futex.Addr()
	if !atomic.CompareAndSwapUint32(addr, cFutexMutexLockedNoWaiters, cFutexMutexUnlocked) {
		*addr = cFutexMutexUnlocked
		if _, err := f.futex.Wake(1); err != nil {
			panic(err)
		}
	}
}

// Close indicates, that the object is no longer in use,
// and that the underlying resources can be freed.
func (f *FutexMutex) Close() error {
	return f.futex.Close()
}

// Destroy removes the mutex object.
func (f *FutexMutex) Destroy() error {
	return f.futex.Destroy()
}

// DestroyFutexMutex permanently removes mutex with the given name.
func DestroyFutexMutex(name string) error {
	m, err := NewFutexMutex(name, 0, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return err
	}
	return m.Destroy()
}

func (f *FutexMutex) lockTimeout(timeout time.Duration) error {
	addr := f.futex.Addr()
	if !atomic.CompareAndSwapUint32(addr, cFutexMutexUnlocked, cFutexMutexLockedNoWaiters) {
		old := atomic.LoadUint32(addr)
		if old != cFutexMutexLockedHaveWaiters {
			old = atomic.SwapUint32(addr, cFutexMutexLockedHaveWaiters)
		}
		for old != cFutexMutexUnlocked {
			if err := f.futex.Wait(cFutexMutexLockedHaveWaiters, timeout); err != nil {
				if !common.SyscallErrHasCode(err, unix.EWOULDBLOCK) {
					return err
				}
			}
			old = atomic.SwapUint32(addr, cFutexMutexLockedHaveWaiters)
		}
	}
	return nil
}
