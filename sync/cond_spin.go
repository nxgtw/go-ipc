// Copyright 2016 Aleksandr Demakin. All rights reserved.

//+build ignore

package sync

import (
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	condWaiterSize = 4

	cSpinWaiterUnset    = 0
	cSpinWaiterWaitDone = 1
	cSpinWaiterSet      = 2
)

type waiter uint32

func newWaiter(ptr unsafe.Pointer) *waiter {
	w := (*waiter)(ptr)
	*w = cSpinWaiterUnset
	return w
}

func openWaiter(ptr unsafe.Pointer) *waiter {
	return (*waiter)(ptr)
}

// signal wakes a spin waiter.
func (w *waiter) signal() (signaled bool) {
	return atomic.CompareAndSwapUint32((*uint32)(w), cSpinWaiterUnset, cSpinWaiterSet)
}

func (w *waiter) waitTimeout(timeout time.Duration) bool {
	var attempt uint64
	start := time.Now()
	ptr := (*uint32)(w)
	for !atomic.CompareAndSwapUint32(ptr, cSpinWaiterSet, cSpinWaiterWaitDone) {
		if attempt%1000 == 0 { // do not call time.Since too often.
			if timeout >= 0 && time.Since(start) >= timeout {
				// if we changed the value from 'unset' to 'done', than the waiter had not been set, return false.
				// otherwise, the value is 'set', we consider waiting to be successful and return true.
				ret := !atomic.CompareAndSwapUint32(ptr, cSpinWaiterUnset, cSpinWaiterWaitDone)
				if ret {
					atomic.StoreUint32(ptr, cSpinWaiterWaitDone)
				}
				return ret
			}
			runtime.Gosched()
		}
		attempt++
	}
	return true
}

func (w *waiter) isSame(ptr unsafe.Pointer) bool {
	return unsafe.Pointer(w) == ptr
}

// destroy is a no-op for a spin waiter.
func (w *waiter) destroy() {}
