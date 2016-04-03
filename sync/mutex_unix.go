// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package sync

import (
	"os"
	"syscall"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/common"
)

type mutex struct {
	name string
	id   int
}

func newMutex(name string, mode int, perm os.FileMode) (*mutex, error) {
	k, err := common.KeyForName(name)
	if err != nil {
		return nil, err
	}
	var id int
	opener := func() error {
		id, err = semget(k, 1, int(perm))
		return err
	}
	creator := func() error {
		id, err = semget(k, 1, common.IpcCreate|common.IpcExcl|int(perm))
		return err
	}
	created, err := common.OpenOrCreate(creator, opener, mode)
	if err != nil {
		return nil, err
	}
	if created {
		if err = semAdd(id, 1); err != nil {
			semctl(id, 0, common.IpcRmid)
			return nil, err
		}
	}
	return &mutex{name: name, id: id}, nil
}

func (m *mutex) Lock() {
	for {
		if err := semAdd(m.id, -1); err == nil {
			return
		} else if !isInterruptedSyscallErr(err) {
			panic(err)
		}
	}
}

func (m *mutex) Unlock() {
	for {
		if err := semAdd(m.id, 1); err == nil {
			return
		} else if !isInterruptedSyscallErr(err) {
			panic(err)
		}
	}
}

// Close is a no-op for unix semaphore
func (m *mutex) Close() error {
	return nil
}

// Destroy closes the mutex and removes it permanently
func (m *mutex) Destroy() error {
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

func isInterruptedSyscallErr(err error) bool {
	if sysErr, ok := err.(*os.SyscallError); ok {
		if errno, ok := sysErr.Err.(syscall.Errno); ok {
			return errno == syscall.Errno(syscall.EINTR)
		}
	}
	return false
}
