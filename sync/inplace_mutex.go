// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux freebsd

package sync

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	cInplaceSpinCount              = 100
	cInplaceMutexUnlocked          = uint32(0)
	cInplaceMutexLockedNoWaiters   = uint32(1)
	cInplaceMutexLockedHaveWaiters = uint32(2)
)

const (
	inplaceMutexSize = int(unsafe.Sizeof(inplaceMutex{}))
)

// inplaceMutex must implement sync.Locker on all platforms.
var (
	_ sync.Locker = (*inplaceMutex)(nil)
)

type wakeFunc func(ptr *uint32)
type waitFunc func(ptr *uint32, timeout time.Duration) error

// inplaceMutex is a mutex, which can be placed into a shared memory region.
type inplaceMutex struct {
	ptr      *uint32
	wakeImpl wakeFunc
	waitImpl waitFunc
}

// NewinplaceMutex creates a new mutex based on a memory location.
//	ptr - memory location for the state.
func newInplaceMutex(ptr unsafe.Pointer, wake wakeFunc, wait waitFunc) *inplaceMutex {
	return &inplaceMutex{ptr: (*uint32)(ptr), wakeImpl: wake, waitImpl: wait}
}

// Init writes initial value into futex's memory location.
func (im *inplaceMutex) Init() {
	*im.ptr = cInplaceMutexUnlocked
}

// Lock locks the locker.
func (im *inplaceMutex) Lock() {
	if err := im.lockTimeout(-1); err != nil {
		panic(err)
	}
}

// TryLock tries to lock the locker. Return true, if it was locked.
func (im *inplaceMutex) TryLock() bool {
	return atomic.CompareAndSwapUint32(im.ptr, cInplaceMutexUnlocked, cInplaceMutexLockedNoWaiters)
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
func (im *inplaceMutex) LockTimeout(timeout time.Duration) bool {
	err := im.lockTimeout(timeout)
	if err == nil {
		return true
	}
	if common.IsTimeoutErr(err) {
		return false
	}
	panic(err)
}

// Unlock releases the mutex. It panics on an error, or if the mutex is not locked.
func (im *inplaceMutex) Unlock() {
	if old := atomic.LoadUint32(im.ptr); old == cInplaceMutexLockedHaveWaiters {
		*im.ptr = cInplaceMutexUnlocked
	} else {
		if old == cInplaceMutexUnlocked {
			panic("unlock of unlocked mutex")
		}
		if atomic.SwapUint32(im.ptr, cInplaceMutexUnlocked) == cInplaceMutexLockedNoWaiters {
			return
		}
	}
	for i := 0; i < cInplaceSpinCount; i++ {
		if *im.ptr != cInplaceMutexUnlocked {
			if atomic.CompareAndSwapUint32(im.ptr, cInplaceMutexLockedNoWaiters, cInplaceMutexLockedHaveWaiters) {
				return
			}
		}
		runtime.Gosched()
	}
	im.wakeImpl(im.ptr)
}

func (im *inplaceMutex) lockTimeout(timeout time.Duration) error {
	for i := 0; i < cInplaceSpinCount; i++ {
		if atomic.CompareAndSwapUint32(im.ptr, cInplaceMutexUnlocked, cInplaceMutexLockedNoWaiters) {
			return nil
		}
		runtime.Gosched()
	}
	old := atomic.LoadUint32(im.ptr)
	if old != cInplaceMutexLockedHaveWaiters {
		old = atomic.SwapUint32(im.ptr, cInplaceMutexLockedHaveWaiters)
	}
	for old != cInplaceMutexUnlocked {
		if err := im.waitImpl(im.ptr, timeout); err != nil {
			return err
		}
		old = atomic.SwapUint32(im.ptr, cInplaceMutexLockedHaveWaiters)
	}
	return nil
}

func defaultWake(ptr *uint32) {

}

func defaultWait(ptr *uint32, timeout time.Duration) error {
	runtime.Gosched()
	return nil
}
