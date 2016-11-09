// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

type event struct {
	handle windows.Handle
}

func newEvent(name string, flag int, perm os.FileMode, initial bool) (*event, error) {
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
	return &event{handle: handle}, nil
}

func (e *event) set() {
	if err := windows.SetEvent(e.handle); err != nil {
		panic("failed to set an event: " + err.Error())
	}
}

func (e *event) wait() {
	e.waitTimeout(-1)
}

func (e *event) waitTimeout(timeout time.Duration) bool {
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

func (e *event) close() error {
	if e.handle == windows.InvalidHandle {
		return nil
	}
	err := windows.CloseHandle(e.handle)
	e.handle = windows.InvalidHandle
	return err
}

func (e *event) destroy() error {
	if err := e.close(); err != nil {
		return errors.Wrap(err, "failed to close windows handle")
	}
	return nil
}

// windows event is destroyed when all its instances are closed.
func destroyEvent(name string) error {
	return nil
}
