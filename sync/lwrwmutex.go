// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"sync/atomic"
	"unsafe"
)

const (
	lwRWMStateSize          = 8
	lwRWMMask               = 0x1FFFFF
	lwRWMWaitingReaderShift = 21
	lwRWMWriterShift        = 42
)

// wRWState is a shared rwmutex state with the following bits distribution:
//  ...63...|62.................42|41.................21|20.................0|
//  --------|---------------------|---------------------|--------------------|
//   unused |       writers       |   waiting readers   |      readers       |
// which gives us up to 2kk readers and writers.
type lwRWState int64

func (s lwRWState) readers() int64 {
	return (int64)(s) & lwRWMMask
}

func (s lwRWState) waitingReaders() int64 {
	return ((int64)(s) >> lwRWMWaitingReaderShift) & lwRWMMask
}

func (s lwRWState) writers() int64 {
	return ((int64)(s) >> lwRWMWriterShift) & lwRWMMask
}

func (s *lwRWState) addReaders(count int64) {
	*(*int64)(s) += count
}

func (s *lwRWState) addWaitingReaders(count int64) {
	*(*int64)(s) += count << lwRWMWaitingReaderShift
}

func (s *lwRWState) addWriters(count int64) {
	*(*int64)(s) += count << lwRWMWriterShift
}

// lwRWMutex is an optimized low-level rwmutex implementation,
// that doesn't have internal lock for its state.
// this implementation is inspired by Jeff Preshing and his article at
// http://preshing.com/20150316/semaphores-are-surprisingly-versatile/
// and his c++ implementation (github.com/preshing/cpp11-on-multicore).
type lwRWMutex struct {
	rWaiter waitWaker
	wWaiter waitWaker
	state   *int64
}

func newRWLightweightMutex(state unsafe.Pointer, rWaiter, wWaiter waitWaker) *lwRWMutex {
	return &lwRWMutex{state: (*int64)(state), rWaiter: rWaiter, wWaiter: wWaiter}
}

// init writes initial value into mutex's memory location.
func (lwrw *lwRWMutex) init() {
	*lwrw.state = 0
}

func (lwrw *lwRWMutex) lock() {
	new := (lwRWState)(atomic.AddInt64(lwrw.state, 1<<lwRWMWriterShift))
	if new.readers() > 0 || new.writers() > 1 {
		if err := lwrw.wWaiter.wait(0, -1); err != nil {
			panic(err)
		}
	}
}

func (lwrw *lwRWMutex) rlock() {
	var new lwRWState
	for {
		old := (lwRWState)(atomic.LoadInt64(lwrw.state))
		new = old
		if new.writers() == 0 {
			new.addReaders(1)
		} else {
			new.addWaitingReaders(1)
		}
		if atomic.CompareAndSwapInt64(lwrw.state, (int64)(old), (int64)(new)) {
			break
		}
	}
	if new.writers() > 0 {
		if err := lwrw.rWaiter.wait(0, -1); err != nil {
			panic(err)
		}
	}
}

func (lwrw *lwRWMutex) runlock() {
	new := (lwRWState)(atomic.AddInt64(lwrw.state, -1))
	if new.readers() == lwRWMMask {
		panic("unlock of unlocked mutex")
	}
	if new.readers() == 0 && new.writers() > 0 {
		lwrw.wWaiter.wake(1)
	}
}

func (lwrw *lwRWMutex) unlock() {
	var new lwRWState
	for {
		old := (lwRWState)(atomic.LoadInt64(lwrw.state))
		if old.writers() == 0 {
			panic("unlock of unlocked mutex")
		}
		new = old
		new.addWriters(-1)
		if wr := new.waitingReaders(); wr > 0 {
			new.addWaitingReaders(-wr)
			new.addReaders(wr)
		}
		if atomic.CompareAndSwapInt64(lwrw.state, (int64)(old), (int64)(new)) {
			break
		}
	}
	if new.readers() > 0 {
		lwrw.rWaiter.wake(int32(new.readers()))
	} else if new.writers() > 0 {
		lwrw.wWaiter.wake(1)
	}
}
