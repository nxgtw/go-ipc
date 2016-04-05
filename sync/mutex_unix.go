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
	creator := func(create bool) error {
		var creatorErr error
		flags := int(perm)
		if create {
			flags |= common.IpcCreate | common.IpcExcl
		}
		id, creatorErr = semget(k, 1, flags)
		return creatorErr
	}
	created, err := common.OpenOrCreate(creator, mode)
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
	return semop(id, []sembuf{b})
}

func isInterruptedSyscallErr(err error) bool {
	return syscallErrHasCode(err, syscall.EINTR)
}

func isTimeoutErr(err error) bool {
	return syscallErrHasCode(err, syscall.EAGAIN)
}

func syscallErrHasCode(err error, code syscall.Errno) bool {
	if sysErr, ok := err.(*os.SyscallError); ok {
		if errno, ok := sysErr.Err.(syscall.Errno); ok {
			return errno == code
		}
	}
	return false
}
