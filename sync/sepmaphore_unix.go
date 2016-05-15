// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package sync

import (
	"os"

	"bitbucket.org/avd/go-ipc/internal/common"
)

// Semaphore is a sysV semaphore.
type Semaphore struct {
	name string
	id   int
}

// NewSemaphore creates a new sysV semaphore with the given name.
// It generates a key from the name, and then calls NewSemaphoreKey.
func NewSemaphore(name string, mode int, perm os.FileMode, initial int) (*Semaphore, error) {
	k, err := common.KeyForName(name)
	if err != nil {
		return nil, err
	}
	result, err := NewSemaphoreKey(uint64(k), mode, perm, initial)
	if err != nil {
		return nil, err
	}
	result.name = name
	return result, nil
}

// NewSemaphoreKey creates a new sysV semaphore for the given key.
//	key - object key. each semaphore object is identifyed by a unique key.
//	mode - object creation mode. must be one of the following:
//		O_OPEN_OR_CREATE
//		O_CREATE_ONLY
//		O_OPEN_ONLY
//	perm - object permissions
//	initial - this value will be added to the semaphore's value, if it was created.
func NewSemaphoreKey(key uint64, mode int, perm os.FileMode, initial int) (*Semaphore, error) {
	var id int
	creator := func(create bool) error {
		var creatorErr error
		flags := int(perm)
		if create {
			flags |= common.IpcCreate | common.IpcExcl
		}
		id, creatorErr = semget(common.Key(key), 1, flags)
		return creatorErr
	}
	created, err := common.OpenOrCreate(creator, mode)
	if err != nil {
		return nil, err
	}
	result := &Semaphore{id: id}
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
	if err == nil && len(s.name) > 0 {
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
