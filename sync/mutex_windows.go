// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"

	ipc "bitbucket.org/avd/go-ipc"

	"golang.org/x/sys/windows"
)

type mutex struct {
	handle windows.Handle
}

func newMutex(name string, mode int, perm os.FileMode) (*mutex, error) {
	var handle windows.Handle
	var err error
	switch mode {
	case ipc.O_OPEN_ONLY:
		handle, err = openMutex(name)
	case ipc.O_CREATE_ONLY:
		handle, err = createMutex(name)
		if handle != windows.Handle(0) && os.IsExist(err) {
			windows.CloseHandle(handle)
		}
	case ipc.O_OPEN_OR_CREATE:
		handle, err = createMutex(name)
		if handle != windows.Handle(0) && os.IsExist(err) {
			err = nil
		}
	}
	if err != nil {
		return nil, err
	}
	return &mutex{handle: handle}, nil
}

func (m *mutex) Lock() {
	windows.WaitForSingleObject(m.handle, windows.INFINITE)
}

func (m *mutex) Unlock() {
	if err := releaseMutex(m.handle); err != nil {
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
