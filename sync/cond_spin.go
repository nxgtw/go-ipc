// Copyright 2016 Aleksandr Demakin. All rights reserved.

//+build windows darwin

package sync

import (
	"runtime"
	"sync/atomic"
	"time"
)

const (
	cSpinWaiterLocked        = 0
	cSpinWaiterWaitCancelled = 1
	cSpinWaiterUnlocked      = 2
	cSpinWaiterWaitDone      = 3
)

type waiter uint32

func (w *waiter) signal(old, new uint32) bool {
	return atomic.CompareAndSwapUint32((*uint32)(w), old, new)
}

func (w *waiter) wait(value uint32) {
	for atomic.LoadUint32((*uint32)(w)) != value {
		runtime.Gosched()
	}
}

func (w *waiter) waitTimeout(value, newValue, fallbackValue uint32, timeout time.Duration) bool {
	var attempt uint64
	start := time.Now()
	for !w.signal(value, newValue) {
		runtime.Gosched()
		if attempt%1000 == 0 { // do not call time.Since too often.
			if timeout >= 0 && time.Since(start) >= timeout {
				last := atomic.LoadUint32((*uint32)(w))
				for !w.signal(last, fallbackValue) {
					if last = atomic.LoadUint32((*uint32)(w)); last == value {
						return true
					}
				}
				return false
			}
		}
		attempt++
	}
	return true
}
