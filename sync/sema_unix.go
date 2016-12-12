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

// semaphore is a sysV semaphore.
type semaphore struct {
	name string
	id   int
}

// newSemaphore creates a new sysV semaphore with the given name.
// It generates a key from the name, and then calls NewSemaphoreKey.
func newSemaphore(name string, flag int, perm os.FileMode, initial int) (*semaphore, error) {
	k, err := common.KeyForName(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate a key for the name")
	}
	result, err := newSemaphoreKey(uint64(k), flag, perm, initial)
	if err != nil {
		return nil, err
	}
	result.name = name
	return result, nil
}

// newSemaphoreKey creates a new sysV semaphore for the given key.
func newSemaphoreKey(key uint64, flag int, perm os.FileMode, initial int) (*semaphore, error) {
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
	result := &semaphore{id: id}
	if created && initial > 0 {
		if err = result.add(initial); err != nil {
			result.Destroy()
			return nil, errors.Wrap(err, "failed to add initial semaphore value")
		}
	}
	return result, nil
}

func (s *semaphore) Signal(count int) {
	if err := s.add(count); err != nil {
		panic(err)
	}
}

func (s *semaphore) Wait() {
	if err := s.add(-1); err != nil {
		panic(err)
	}
}

// Close is a no-op on unix.
func (s *semaphore) Close() error {
	return nil
}

func (s *semaphore) Destroy() error {
	return removeSysVSemaByID(s.id, s.name)
}

func (s *semaphore) add(value int) error {
	return common.UninterruptedSyscall(func() error { return semAdd(s.id, value) })
}

// destroySemaphore permanently removes semaphore with the given name.
func destroySemaphore(name string) error {
	k, err := common.KeyForName(name)
	if err != nil {
		return errors.Wrap(err, "failed to get a key for the name")
	}
	id, err := semget(k, 1, 0)
	if err != nil {
		return errors.Wrap(err, "failed to get semaphore id")
	}
	return removeSysVSemaByID(id, name)
}

func removeSysVSemaByID(id int, name string) error {
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
