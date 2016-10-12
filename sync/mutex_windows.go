// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

// all implementations must satisfy IPCLocker interface.
var (
	_ IPCLocker = (*EventMutex)(nil)
)

// EventMutex is a mutex built on named windows events.
// It is not possible to use native windows named mutex, because
// goroutines migrate between threads, and windows mutex must
// be released by the same thread it was locked.
type EventMutex struct {
	handle windows.Handle
}

// NewEventMutex creates a new event-basedmutex.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
func NewEventMutex(name string, flag int, perm os.FileMode) (*EventMutex, error) {
	handle, err := openOrCreateEvent(name, flag, 1)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open/create event mutex")
	}
	return &EventMutex{handle: handle}, nil
}

// Lock locks the mutex. It panics on an error.
func (m *EventMutex) Lock() {
	ev, err := windows.WaitForSingleObject(m.handle, windows.INFINITE)
	if ev != windows.WAIT_OBJECT_0 {
		if err != nil {
			panic(err)
		} else {
			panic(errors.Errorf("invalid wait state for a mutex: %d", ev))
		}
	}
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
func (m *EventMutex) LockTimeout(timeout time.Duration) bool {
	waitMillis := uint32(timeout.Nanoseconds() / 1e6)
	ev, err := windows.WaitForSingleObject(m.handle, waitMillis)
	switch ev {
	case windows.WAIT_OBJECT_0:
		return true
	case windows.WAIT_TIMEOUT:
		return false
	default:
		if err != nil {
			panic(err)
		} else {
			panic(errors.Errorf("invalid wait state for a mutex: %d", ev))
		}
	}
}

// Unlock releases the mutex. It panics on an error.
func (m *EventMutex) Unlock() {
	if err := windows.SetEvent(m.handle); err != nil {
		panic("failed to unlock mutex: " + err.Error())
	}
}

// Close closes event's handle.
func (m *EventMutex) Close() error {
	return windows.CloseHandle(m.handle)
}

// DestroyEventMutex is a no-op on windows, as the mutex is destroyed,
// when its last handle is closed.
func DestroyEventMutex(name string) error {
	return nil
}
