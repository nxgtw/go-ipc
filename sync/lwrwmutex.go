// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"sync/atomic"
	"unsafe"
)

const (
	lwRWCellSize = 8
)

// wRWState is a shared rwmutex state.
//   63........62.........41...................20....0
//  unused      writers     waiting readers     readers
type lwRWState uint64

func (s *lwRWState) load() lwRWState {
	return (lwRWState)(atomic.LoadUint64((*uint64)(s)))
}

func (s *lwRWState) readers() int {
	return (int)(*(*uint64)(s) & 0xFFFFF)
}

type lwRWMutex struct {
	rWaiter waitWaker
	wWaiter waitWaker
	state   *lwRWState
}

func newRWLightweightMutex(state unsafe.Pointer, rWaiter, wWaiter waitWaker) *lwRWMutex {
	return &lwRWMutex{state: (*lwRWState)(state), rWaiter: rWaiter, wWaiter: wWaiter}
}

// init writes initial value into mutex's memory location.
func (lwrw *lwRWMutex) init() {
	*lwrw.state = 0
}

func (lwrw *lwRWMutex) lock() {
	old := lwrw.state.load()
	for {
		new := old
		_ = new
	}
}

func (lwrw *lwRWMutex) rlock() {

}

func (lwrw *lwRWMutex) runlock() {

}

func (lwrw *lwRWMutex) unlock() {

}
