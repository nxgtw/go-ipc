// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"sync/atomic"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	lwmStateSize = 4

	lwmSpinCount         = 100
	lwmUnlocked          = int32(0)
	lwmLockedNoWaiters   = int32(1)
	lwmLockedHaveWaiters = int32(2)
)

// lwMutex is a lightweight mutex implementation operating on a uint32 memory cell.
// it tries to minimize amount of syscalls needed to do locking.
// actual sleeping must be implemented by a waitWaker object.
type lwMutex struct {
	state *int32
	ww    waitWaker
}

func newLightweightMutex(state unsafe.Pointer, ww waitWaker) *lwMutex {
	return &lwMutex{state: (*int32)(state), ww: ww}
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
	return atomic.CompareAndSwapInt32(lwm.state, lwmUnlocked, lwmLockedNoWaiters)
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
	old := atomic.LoadInt32(lwm.state)
	if old != lwmLockedHaveWaiters {
		old = atomic.SwapInt32(lwm.state, lwmLockedHaveWaiters)
	}
	for old != lwmUnlocked {
		if err := lwm.ww.wait(lwmLockedHaveWaiters, timeout); err != nil {
			return err
		}
		old = atomic.SwapInt32(lwm.state, lwmLockedHaveWaiters)
	}
	return nil
}

func (lwm *lwMutex) unlock() {
	if old := atomic.LoadInt32(lwm.state); old == lwmLockedHaveWaiters {
		*lwm.state = lwmUnlocked
	} else {
		if old == lwmUnlocked {
			panic("unlock of unlocked mutex")
		}
		if atomic.SwapInt32(lwm.state, lwmUnlocked) == lwmLockedNoWaiters {
			return
		}
	}
	for i := 0; i < lwmSpinCount; i++ {
		if *lwm.state != lwmUnlocked {
			if atomic.CompareAndSwapInt32(lwm.state, lwmLockedNoWaiters, lwmLockedHaveWaiters) {
				return
			}
		}
	}
	lwm.ww.wake(1)
}
