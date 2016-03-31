// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"sync"
	"fmt"
	"os"

	ipc "bitbucket.org/avd/go-ipc"

	"golang.org/x/sys/windows"
)

type mutexImpl struct {
	handle windows.Handle
}

func newMutexImpl(name string, mode int, perm os.FileMode) (*mutexImpl, error) {
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
	return &mutexImpl{handle: handle}, nil
}

func (m *mutexImpl) Lock() {
	windows.WaitForSingleObject(m.handle, windows.INFINITE)
}

func (m *mutexImpl) Unlock() {
	if err := releaseMutex(m.handle); err != nil {
		fmt.Printf("%v", err)
		var m sync.Mutex{}
		m.Lock()
		os.Exit(1)
	}
}

func (m *mutexImpl) Close() error {
	return windows.CloseHandle(m.handle)
}

// DestroyMutex is a no-op on windows, as the mutex is destroyed,
// when its last handle is closed.
func DestroyMutex(name string) error {
	return nil
}
