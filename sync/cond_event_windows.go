// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build ignore

package sync

import (
	"os"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/sys/windows"
)

type waiter struct {
	pid     int
	eventId uintptr
}

func newWaiter() waiter {
	handle, err := openOrCreateEvent("", os.O_CREATE|os.O_EXCL, 0)
	if err != nil {
		panic(errors.Wrap(err, "cond: failed to create an event"))
	}
	return waiter{pid: os.Getpid(), eventId: uintptr(handle)}
}

func (w *waiter) signal() {
	if w.pid == os.Getpid() {
		windows.SetEvent(windows.Handle(w.eventId))
	} else {
		panic("another process")
	}
}

func (w waiter) wait(timeout time.Duration) bool {
	millis := uint32(windows.INFINITE)
	if timeout >= 0 {
		millis = uint32(timeout.Nanoseconds() / 1e6)
	}
	ev, err := windows.WaitForSingleObject(windows.Handle(w.eventId), millis)
	switch ev {
	case windows.WAIT_OBJECT_0:
		// success, lock resource locker
		return true
	case windows.WAIT_TIMEOUT:
		return false
	default:
		if err != nil {
			panic(err)
		} else {
			panic(errors.Errorf("invalid wait state for an event: %d", ev))
		}
	}
}
