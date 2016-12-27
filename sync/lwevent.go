// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"sync/atomic"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	lweStateSize = 4
)

// lwEvent is a lightweight event implementation operating on a uint32 memory cell.
// it tries to minimize amount of syscalls.
// actual wait/wake must be implemented by a waitWaker object.
//	state == 1 - signaled
//	state == 0 - not signaled and no waiters
//	state == -n - have n waiters.
// this implementation is inspired by Jeff Preshing and his article at
// http://preshing.com/20150316/semaphores-are-surprisingly-versatile/
// and his c++ implementation (github.com/preshing/cpp11-on-multicore).
type lwEvent struct {
	state *int32
	ww    waitWaker
}

func newLightweightEvent(state unsafe.Pointer, ww waitWaker) *lwEvent {
	return &lwEvent{state: (*int32)(state), ww: ww}
}

func (e *lwEvent) init(set bool) {
	val := int32(0)
	if set {
		val = 1
	}
	*e.state = val
}

func (e *lwEvent) signal() {
	var old int32
	for {
		old = atomic.LoadInt32(e.state)
		new := old
		if new <= 0 {
			new++
		}
		if atomic.CompareAndSwapInt32(e.state, old, new) {
			break
		}
	}
	if old < 0 {
		e.ww.wake(1)
	}
}

func (e *lwEvent) waitTimeout(timeout time.Duration) bool {
	new := atomic.AddInt32(e.state, -1)
	if new < 0 {
		if err := e.ww.wait(old, timeout); err != nil {
			atomic.AddInt32(e.state, 1)
			if common.IsTimeoutErr(err) {
				return false
			}
			panic(err)
		}
	}
	return true
}
