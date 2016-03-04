// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"

	"golang.org/x/sys/windows"
)

type mutexImpl struct {
	handle windows.Handle
}

func newMutexImpl(name string, mode int, perm os.FileMode) (*mutexImpl, error) {
	var handle windows.Handle
	var err error
	switch mode {
	case O_OPEN_ONLY:
		handle, err = openMutex(name)
	case O_CREATE_ONLY:
		handle, err = createMutex(name)
	case O_OPEN_OR_CREATE:
		handle, err = createMutex(name)
		if handle != windows.Handle(0) && os.IsExist(err) {
			err = nil
		}
	}
	if err != nil {
		return nil, err
	}
	return &mutexImpl{handle: handle}, nil
}

func (m *mutexImpl) Lock() {
	windows.WaitForSingleObject(m.handle, windows.INFINITE)
}

func (m *mutexImpl) Unlock() {
	releaseMutex(m.handle)
}

func (m *mutexImpl) Finish() error {
	return windows.CloseHandle(m.handle)
}
