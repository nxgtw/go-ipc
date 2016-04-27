// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package sync

import (
	"os"

	"bitbucket.org/avd/go-ipc/internal/common"
)

// NewSemaphore is a sysV semaphore.
type Semaphore struct {
	name string
	id   int
}

// NewSemaphore creates a new sysV semaphore.
//	name - object name
//	mode - object creation mode. must be one of the following:
//		O_OPEN_OR_CREATE
//		O_CREATE_ONLY
//		O_OPEN_ONLY
//	perm - object permissions
//	initial - this value will be added to the semaphore's value, if it was created.
func NewSemaphore(name string, mode int, perm os.FileMode, initial int) (*Semaphore, error) {
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
	result := &Semaphore{name: name, id: id}
	if created && initial > 0 {
		if err = result.Add(initial); err != nil {
			result.Destroy()
			return nil, err
		}
	}
	return result, nil
}

// Add adds the given value to the semaphore's value.
// It locks, if the operation cannot be done immediately.
func (s *Semaphore) Add(value int) error {
	f := func() error { return semAdd(s.id, value) }
	return common.UninterruptedSyscall(f)
}

// Destroy removes the semaphore permanently.
func (s *Semaphore) Destroy() error {
	err := semctl(s.id, 0, common.IpcRmid)
	if err == nil {
		if err = os.Remove(common.TmpFilename(s.name)); os.IsNotExist(err) {
			err = nil
		}
	} else if os.IsNotExist(err) {
		err = nil
	}
	return err
}

func semAdd(id, value int) error {
	b := sembuf{semnum: 0, semop: int16(value), semflg: 0}
	return semop(id, []sembuf{b})
}
