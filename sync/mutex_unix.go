// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux

package sync

import (
	"os"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"

	"github.com/pkg/errors"
)

// all implementations must satisfy IPCLocker interface.
var (
	_ IPCLocker = (*SemaMutex)(nil)
)

// SemaMutex is a semaphore-based mutex for unix.
type SemaMutex struct {
	s       *Semaphore
	state   *mmf.MemoryRegion
	name    string
	inplace *inplaceMutex
}

// NewSemaMutex creates a new semaphore-based mutex.
func NewSemaMutex(name string, flag int, perm os.FileMode) (*SemaMutex, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}
	s, err := NewSemaphore(name, flag, perm, 1)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a semaphore")
	}
	region, err := createWritableRegion(semaMutexSharedStateName(name), flag, perm, inplaceMutexSize, cInplaceMutexUnlocked)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create shared state")
	}
	result := &SemaMutex{
		s:     s,
		state: region,
		name:  name,
	}
	result.inplace = newInplaceMutex(allocator.ByteSliceData(region.Data()), result.wake, result.wait)
	return result, nil
}

// Lock locks the mutex. It panics on an error.
func (m *SemaMutex) Lock() {
	m.inplace.Lock()
}

// Unlock releases the mutex. It panics on an error, or if the mutex is not locked.
func (m *SemaMutex) Unlock() {
	m.inplace.Unlock()
}

func (m *SemaMutex) wake(ptr *uint32) {
	if err := m.s.Add(1); err != nil {
		panic(err)
	}
}

/*
func (m *SemaMutex) wait(ptr *uint32, timeout time.Duration) error {
	return m.s.Add(-1)
}*/

// Close closes shared state of the mutex.
func (m *SemaMutex) Close() error {
	return m.state.Close()
}

// Destroy closes the mutex and removes it permanently.
func (m *SemaMutex) Destroy() error {
	if err := m.Close(); err != nil {
		return errors.Wrap(err, "failed to close shared state")
	}
	if err := shm.DestroyMemoryObject(semaMutexSharedStateName(m.name)); err != nil {
		return errors.Wrap(err, "failed to destroy shared state")
	}
	return m.s.Destroy()
}

// DestroySemaMutex permanently removes mutex with the given name.
func DestroySemaMutex(name string) error {
	if err := shm.DestroyMemoryObject(semaMutexSharedStateName(name)); err != nil {
		return errors.Wrap(err, "failed to destroy shared state")
	}
	if err := DestroySemaphore(name); err != nil && !os.IsNotExist(errors.Cause(err)) {
		return err
	}
	return nil
}

func semaMutexSharedStateName(name string) string {
	return "go-ipc.ssema" + name
}
