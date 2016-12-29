// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

const (
	condWaiterSize = 8
)

var (
	pid = os.Getpid()
	seq uint32
)

type waiter struct {
	id *uint64
	e  *Event
}

func nextID() uint32 {
	return atomic.AddUint32(&seq, 1)
}

func newWaiter(ptr unsafe.Pointer) *waiter {
	for {
		id := uint64(pid)<<32 | uint64(nextID())
		e, err := NewEvent(condWaiterEventName(id), os.O_CREATE|os.O_EXCL, 0666, false)
		if err == nil {
			result := &waiter{id: (*uint64)(ptr), e: e}
			*result.id = id
			return result
		}
		if !os.IsExist(errors.Cause(err)) {
			panic(errors.Wrap(err, "cond: failed to create an event"))
		}
	}
}

func openWaiter(ptr unsafe.Pointer) *waiter {
	return &waiter{id: (*uint64)(ptr)}
}

func (w *waiter) signal() bool {
	ev, err := NewEvent(condWaiterEventName(*w.id), 0, 0, false)
	if err != nil {
		if os.IsNotExist(errors.Cause(err)) {
			return false
		}
		panic(err)
	}
	ev.Set()
	ev.Close()
	return true
}

func (w *waiter) isSame(ptr unsafe.Pointer) bool {
	return unsafe.Pointer(w.id) == ptr
}

func (w *waiter) destroy() {
	w.e.Destroy()
}

func (w *waiter) waitTimeout(timeout time.Duration) bool {
	return w.e.WaitTimeout(timeout)
}

func condWaiterEventName(id uint64) string {
	return fmt.Sprintf("cev.%d", id)
}
