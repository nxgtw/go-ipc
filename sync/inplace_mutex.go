// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
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

// waitWaker is an object, which implements wake/wait semantics.
type waitWaker interface {
	wake()
	wait(timeout time.Duration) error
}

// inplaceMutex is a mutex implementation operating on a uint32 memory cell.
// it tries to minimize amount of syscalls needed to do locking.
// actual sleeping must be implemented by a waitWaker object.
type inplaceMutex struct {
	ptr *uint32
	ww  waitWaker
}

func newInplaceMutex(ptr unsafe.Pointer, ww waitWaker) *inplaceMutex {
	return &inplaceMutex{ptr: (*uint32)(ptr), ww: ww}
}

// init writes initial value into mutex's memory location.
func (im *inplaceMutex) init() {
	*im.ptr = cInplaceMutexUnlocked
}

func (im *inplaceMutex) lock() {
	if err := im.doLock(-1); err != nil {
		panic(err)
	}
}

func (im *inplaceMutex) tryLock() bool {
	return atomic.CompareAndSwapUint32(im.ptr, cInplaceMutexUnlocked, cInplaceMutexLockedNoWaiters)
}

func (im *inplaceMutex) lockTimeout(timeout time.Duration) bool {
	err := im.doLock(timeout)
	if err == nil {
		return true
	}
	if common.IsTimeoutErr(err) {
		return false
	}
	panic(err)
}

func (im *inplaceMutex) doLock(timeout time.Duration) error {
	for i := 0; i < cInplaceSpinCount; i++ {
		if atomic.CompareAndSwapUint32(im.ptr, cInplaceMutexUnlocked, cInplaceMutexLockedNoWaiters) {
			return nil
		}
	}
	old := atomic.LoadUint32(im.ptr)
	if old != cInplaceMutexLockedHaveWaiters {
		old = atomic.SwapUint32(im.ptr, cInplaceMutexLockedHaveWaiters)
	}
	for old != cInplaceMutexUnlocked {
		if err := im.ww.wait(timeout); err != nil {
			return err
		}
		old = atomic.SwapUint32(im.ptr, cInplaceMutexLockedHaveWaiters)
	}
	return nil
}

func (im *inplaceMutex) unlock() {
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
	}
	im.ww.wake()
}
