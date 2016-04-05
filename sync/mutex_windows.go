// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/windows"
)

type mutex struct {
	handle windows.Handle
}

func newMutex(name string, mode int, perm os.FileMode) (*mutex, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}
	var handle windows.Handle
	creator := func(create bool) error {
		var err error
		handle, err = openEvent(name, windows.SYNCHRONIZE, uint32(0))
		if create {
			if handle != windows.Handle(0) {
				// this is emulation of O_EXCL. despite wait MSDN for CreateEvent says:
				// "If the named event object existed before the function call, the function returns a handle
				// to the existing object and GetLastError returns ERROR_ALREADY_EXIST",
				// we cannot actually find out with CreateEvent
				// if the event has already existed if was created in the same process.
				// so, we just do a check with OpenEvent.
				// yes, there is a race condition.
				windows.CloseHandle(handle)
				return windows.ERROR_ALREADY_EXISTS
			} else {
				handle, err = windows.CreateEvent(nil, 0, 1, namep)
			}
		}
		if handle != windows.Handle(0) {
			return nil
		}
		return err
	}
	_, err = common.OpenOrCreate(creator, mode)
	if err != nil {
		return nil, err
	}
	return &mutex{handle: handle}, nil
}

func (m *mutex) Lock() {
	ev, err := windows.WaitForSingleObject(m.handle, windows.INFINITE)
	if ev != windows.WAIT_OBJECT_0 {
		if err != nil {
			panic(err)
		} else {
			panic(fmt.Errorf("invalid wait state for a mutex: %d", ev))
		}
	}
}

func (m *mutex) LockTimeout(timeout time.Duration) bool {
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
			panic(fmt.Errorf("invalid wait state for a mutex: %d", ev))
		}
	}
}

func (m *mutex) Unlock() {
	if err := windows.SetEvent(m.handle); err != nil {
		panic("failed to unlock mutex: " + err.Error())
	}
}

func (m *mutex) Close() error {
	return windows.CloseHandle(m.handle)
}

// DestroyMutex is a no-op on windows, as the mutex is destroyed,
// when its last handle is closed.
func DestroyMutex(name string) error {
	return nil
}
