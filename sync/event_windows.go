// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

// WindowsEvent gives access to system event object.
type WindowsEvent struct {
	handle windows.Handle
}

// NewWindowsEvent returns new WindowsEvent.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
//	initial - if truem the event will be set after creation.
func NewWindowsEvent(name string, flag int, perm os.FileMode, initial bool) (*WindowsEvent, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}
	var init int
	if initial {
		init = 1
	}
	handle, err := openOrCreateEvent(name, flag, init)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open/create event")
	}
	return &WindowsEvent{handle: handle}, nil
}

// Set sets the specified event object to the signaled state.
func (e *WindowsEvent) Set() {
	if err := windows.SetEvent(e.handle); err != nil {
		panic("failed to set an event: " + err.Error())
	}
}

// Wait waits for the event to be signaled.
func (e *WindowsEvent) Wait() {
	e.WaitTimeout(-1)
}

// WaitTimeout waits until the event is signaled or the timeout elapses.
func (e *WindowsEvent) WaitTimeout(timeout time.Duration) bool {
	waitMillis := uint32(windows.INFINITE)
	if timeout >= 0 {
		waitMillis = uint32(timeout.Nanoseconds() / 1e6)
	}
	ev, err := windows.WaitForSingleObject(e.handle, waitMillis)
	switch ev {
	case windows.WAIT_OBJECT_0:
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

// Close closes the event.
func (e *WindowsEvent) Close() error {
	if e.handle == windows.InvalidHandle {
		return nil
	}
	err := windows.CloseHandle(e.handle)
	e.handle = windows.InvalidHandle
	return err
}
