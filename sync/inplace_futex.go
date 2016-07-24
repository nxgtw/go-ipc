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
	"golang.org/x/sys/unix"
)

const (
	inplaceMutexSize             = int(unsafe.Sizeof(uint32(0)))
	cFutexSpinCount              = 100
	cFutexMutexUnlocked          = 0
	cFutexMutexLockedNoWaiters   = 1
	cFutexMutexLockedHaveWaiters = 2
)

// InplaceMutex must implement sync.Locker on all platforms.
var (
	_ sync.Locker = (*InplaceMutex)(nil)
)

// InplaceMutex is a linux futex, which can be placed into a shared memory region.
type InplaceMutex uint32

// NewInplaceMutex creates a futex object on the given memory location.
//	ptr - memory location for the futex.
func NewInplaceMutex(ptr unsafe.Pointer) *InplaceMutex {
	return (*InplaceMutex)(ptr)
}

// Init writes initial value into futex's memory location.
func (f *InplaceMutex) Init() {
	*f.uint32Ptr() = cFutexMutexUnlocked
}

// Lock locks the locker.
func (f *InplaceMutex) Lock() {
	if err := f.lockTimeout(-1); err != nil {
		panic(err)
	}
}

// TryLock tries to lock the locker. Return true, if it was locked.
func (f *InplaceMutex) TryLock() bool {
	addr := f.uint32Ptr()
	return atomic.CompareAndSwapUint32(addr, cFutexMutexUnlocked, cFutexMutexLockedNoWaiters)
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
func (f *InplaceMutex) LockTimeout(timeout time.Duration) bool {
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
func (f *InplaceMutex) Unlock() {
	addr := f.uint32Ptr()
	if old := atomic.LoadUint32(addr); old == cFutexMutexLockedHaveWaiters {
		*addr = cFutexMutexUnlocked
	} else if atomic.SwapUint32(addr, cFutexMutexUnlocked) == cFutexMutexLockedNoWaiters {
		return
	}
	for i := 0; i < cFutexSpinCount; i++ {
		if *addr != cFutexMutexUnlocked {
			if atomic.CompareAndSwapUint32(addr, cFutexMutexLockedNoWaiters, cFutexMutexLockedHaveWaiters) {
				return
			}
		}
		runtime.Gosched()
	}
	if _, err := FutexWake(unsafe.Pointer(f), 1, 0); err != nil {
		panic(err)
	}
}

func (f *InplaceMutex) uint32Ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(f))
}

func (f *InplaceMutex) lockTimeout(timeout time.Duration) error {
	addr := f.uint32Ptr()
	for i := 0; i < cFutexSpinCount; i++ {
		if atomic.CompareAndSwapUint32(addr, cFutexMutexUnlocked, cFutexMutexLockedNoWaiters) {
			return nil
		}
		runtime.Gosched()
	}
	old := atomic.LoadUint32(addr)
	if old != cFutexMutexLockedHaveWaiters {
		old = atomic.SwapUint32(addr, cFutexMutexLockedHaveWaiters)
	}
	for old != cFutexMutexUnlocked {
		if err := FutexWait(unsafe.Pointer(f), cFutexMutexLockedHaveWaiters, timeout, 0); err != nil {
			if !common.SyscallErrHasCode(err, unix.EWOULDBLOCK) {
				return err
			}
		}
		old = atomic.SwapUint32(addr, cFutexMutexLockedHaveWaiters)
	}
	return nil
}
