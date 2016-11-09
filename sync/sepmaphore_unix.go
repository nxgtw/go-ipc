// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux

package sync

import (
	"os"

	"bitbucket.org/avd/go-ipc/internal/common"
	"github.com/pkg/errors"
)

const (
	cSemUndo = 0x1000
)

type sembuf struct {
	semnum uint16
	semop  int16
	semflg int16
}

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
		return nil, errors.Wrap(err, "failed to generate a key for the name")
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
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
//	initial - this value will be added to the semaphore's value, if it was created.
func NewSemaphoreKey(key uint64, flag int, perm os.FileMode, initial int) (*Semaphore, error) {
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
	created, err := common.OpenOrCreate(creator, flag)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open/create sysv semaphore")
	}
	result := &Semaphore{id: id}
	if created && initial > 0 {
		if err = result.Add(initial); err != nil {
			result.Destroy()
			return nil, errors.Wrap(err, "failed to add initial semaphore value")
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
	return removeSemaByID(s.id, s.name)
}

// DestroySemaphore permanently removes semaphore with the given name.
func DestroySemaphore(name string) error {
	k, err := common.KeyForName(name)
	if err != nil {
		return errors.Wrap(err, "failed to get a key for the name")
	}
	id, err := semget(k, 1, 0)
	if err != nil {
		return errors.Wrap(err, "failed to get semaphore id")
	}
	return removeSemaByID(id, name)
}

func removeSemaByID(id int, name string) error {
	err := semctl(id, 0, common.IpcRmid)
	if err == nil && len(name) > 0 {
		if err = os.Remove(common.TmpFilename(name)); os.IsNotExist(err) {
			err = nil
		} else if err != nil {
			err = errors.Wrap(err, "failed to remove temporary file")
		}
	} else if os.IsNotExist(err) {
		err = nil
	} else {
		err = errors.Wrap(err, "semctl failed")
	}
	return err
}

func semAdd(id, value int) error {
	b := sembuf{semnum: 0, semop: int16(value), semflg: 0}
	return semop(id, []sembuf{b})
}
