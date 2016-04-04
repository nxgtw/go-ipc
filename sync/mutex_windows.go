// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"os"

	ipc "bitbucket.org/avd/go-ipc"

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
	handle, err := windows.CreateEvent(nil, 0, 1, namep)
	if handle == windows.Handle(0) {
		return nil, err
	}
	var created bool
	switch mode {
	case ipc.O_OPEN_ONLY:
		if os.IsExist(err) {
			err = nil
		}
	case ipc.O_CREATE_ONLY:
		if !os.IsExist(err) {
			if err != nil {
				print(err.Error())
			}
			err = nil
			created = true
		}
	case ipc.O_OPEN_OR_CREATE:
		if !os.IsExist(err) {
			created = true
		}
		err = nil
	}
	if err != nil {
		return nil, err
	}
	if created {
		if err = windows.SetEvent(handle); err != nil {
			windows.Close(handle)
			return nil, err
		}
	}
	return &mutex{handle: handle}, nil
}

func (m *mutex) Lock() {
	ev, err := windows.WaitForSingleObject(m.handle, windows.INFINITE)
	if ev != windows.WAIT_OBJECT_0 {
		if err != nil {
			panic(err)
		} else {
			panic(fmt.Errorf("invalid wati state for a mutex: %d", ev))
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
