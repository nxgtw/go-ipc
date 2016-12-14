// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"sync/atomic"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	lwmSpinCount         = 100
	lwmUnlocked          = uint32(0)
	lwmLockedNoWaiters   = uint32(1)
	lwmLockedHaveWaiters = uint32(2)
)

const (
	lwmCellSize = 4
)

// waitWaker is an object, which implements wake/wait semantics.
type waitWaker interface {
	wake(count uint32) (int, error)
	wait(value uint32, timeout time.Duration) error
}

// lwMutex is a lightweight mutex implementation operating on a uint32 memory cell.
// it tries to minimize amount of syscalls needed to do locking.
// actual sleeping must be implemented by a waitWaker object.
type lwMutex struct {
	state *uint32
	ww    waitWaker
}

func newLightweightMutex(state unsafe.Pointer, ww waitWaker) *lwMutex {
	return &lwMutex{state: (*uint32)(state), ww: ww}
}

// init writes initial value into mutex's memory location.
func (lwm *lwMutex) init() {
	*lwm.state = lwmUnlocked
}

func (lwm *lwMutex) lock() {
	if err := lwm.doLock(-1); err != nil {
		panic(err)
	}
}

func (lwm *lwMutex) tryLock() bool {
	return atomic.CompareAndSwapUint32(lwm.state, lwmUnlocked, lwmLockedNoWaiters)
}

func (lwm *lwMutex) lockTimeout(timeout time.Duration) bool {
	err := lwm.doLock(timeout)
	if err == nil {
		return true
	}
	if common.IsTimeoutErr(err) {
		return false
	}
	panic(err)
}

func (lwm *lwMutex) doLock(timeout time.Duration) error {
	for i := 0; i < lwmSpinCount; i++ {
		if lwm.tryLock() {
			return nil
		}
	}
	old := atomic.LoadUint32(lwm.state)
	if old != lwmLockedHaveWaiters {
		old = atomic.SwapUint32(lwm.state, lwmLockedHaveWaiters)
	}
	for old != lwmUnlocked {
		if err := lwm.ww.wait(lwmLockedHaveWaiters, timeout); err != nil {
			return err
		}
		old = atomic.SwapUint32(lwm.state, lwmLockedHaveWaiters)
	}
	return nil
}

func (lwm *lwMutex) unlock() {
	if old := atomic.LoadUint32(lwm.state); old == lwmLockedHaveWaiters {
		*lwm.state = lwmUnlocked
	} else {
		if old == lwmUnlocked {
			panic("unlock of unlocked mutex")
		}
		if atomic.SwapUint32(lwm.state, lwmUnlocked) == lwmLockedNoWaiters {
			return
		}
	}
	for i := 0; i < lwmSpinCount; i++ {
		if *lwm.state != lwmUnlocked {
			if atomic.CompareAndSwapUint32(lwm.state, lwmLockedNoWaiters, lwmLockedHaveWaiters) {
				return
			}
		}
	}
	lwm.ww.wake(1)
}
