// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package sync

import (
	"os"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/common"
)

type mutexImpl struct {
	name string
	id   int
}

func newMutexImpl(name string, mode int, perm os.FileMode) (*mutexImpl, error) {
	k, err := common.KeyForName(name)
	if err != nil {
		return nil, err
	}
	var flags int
	switch mode {
	case ipc.O_OPEN_ONLY:
	case ipc.O_CREATE_ONLY:
		flags = common.IpcCreate | common.IpcExcl
	case ipc.O_OPEN_OR_CREATE:
		flags = common.IpcCreate
	}
	id, err := semget(k, 1, flags)
	if err != nil {
		return nil, err
	}
	if err = semAdd(id, 1); err != nil {
		return nil, err
	}
	return &mutexImpl{name: name, id: id}, nil
}

func (m *mutexImpl) Lock() {
	semAdd(m.id, -1)
}

func (m *mutexImpl) Unlock() {
	semAdd(m.id, 1)
}

// Close is a no-op for unix semaphore
func (m *mutexImpl) Close() error {
	return nil
}

// Destroy closes the mutex and removes it permanently
func (m *mutexImpl) Destroy() error {
	m.Close()
	err := semctl(m.id, 0, common.IpcRmid)
	if err == nil {
		if err = os.Remove(common.TmpFilename(m.name)); os.IsNotExist(err) {
			err = nil
		}
	} else if os.IsNotExist(err) {
		err = nil
	}
	return err
}

// DestroyMutex permanently removes mutex with a given name
func DestroyMutex(name string) error {
	m, err := NewMutex(name, ipc.O_OPEN_ONLY, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return err
	}
	return m.Destroy()
}

func semAdd(id, value int) error {
	b := sembuf{semnum: 0, semop: int16(value), semflg: 0}
	return semtimedop(id, []sembuf{b}, nil)
}
