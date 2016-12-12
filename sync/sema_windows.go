// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/sys/windows"
)

// semaphore is a platform specific semaphore implementation.
// on windows it uses system semaphore object.
type semaphore struct {
	handle windows.Handle
}

func newSemaphore(name string, flag int, perm os.FileMode, initial int) (*semaphore, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}
	handle, err := openOrCreateSemaphore(name, flag, initial, CSemMaxVal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open/create semaphore")
	}
	return &semaphore{handle: handle}, nil
}

// destroySemaphore is a no-op on windows.
func destroySemaphore(name string) error {
	return nil
}

func (s *semaphore) Signal(count int) {
	if _, err := sys_ReleaseSemaphore(s.handle, count); err != nil {
		panic(err)
	}
}

func (s *semaphore) Wait() {
	s.WaitTimeout(-1)
}

func (s *semaphore) Close() error {
	if err := windows.CloseHandle(s.handle); err != nil {
		return errors.Wrap(err, "failed to close windows handle")
	}
	return nil
}

func (s *semaphore) WaitTimeout(timeout time.Duration) bool {
	waitMillis := uint32(windows.INFINITE)
	if timeout >= 0 {
		waitMillis = uint32(timeout.Nanoseconds() / 1e6)
	}
	ev, err := windows.WaitForSingleObject(s.handle, waitMillis)
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
