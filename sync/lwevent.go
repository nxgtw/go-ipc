// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import "unsafe"

// lwEvent is a lightweight event implementation operating on a uint32 memory cell.
// it tries to minimize amount of syscalls.
// actual wait/wake must be implemented by a waitWaker object.
type lwEvent struct {
	state *uint32
	ww    waitWaker
}

func newLightweightEvent(state unsafe.Pointer, ww waitWaker) *lwEvent {
	return &lwEvent{state: (*uint32)(state), ww: ww}
}
