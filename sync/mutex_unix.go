// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux

package sync

import (
	"os"

	"github.com/pkg/errors"
)

// all implementations must satisfy IPCLocker interface.
var (
	_ IPCLocker = (*SpinMutex)(nil)
)

// SemaMutex is a semaphore-based mutex for unix.
type SemaMutex struct {
	s *Semaphore
}

// NewSemaMutex creates a new mutex.
func NewSemaMutex(name string, flag int, perm os.FileMode) (*SemaMutex, error) {
	s, err := NewSemaphore(name, flag, perm, 1)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a semaphore")
	}
	return &SemaMutex{s: s}, nil
}

// Lock locks the mutex. It panics on an error.
func (m *SemaMutex) Lock() {
	if err := m.s.Add(-1); err != nil {
		panic(err)
	}
}

// Unlock releases the mutex. It panics on an error.
func (m *SemaMutex) Unlock() {
	if err := m.s.Add(1); err != nil {
		panic(err)
	}
}

// Close is a no-op for unix mutex.
func (m *SemaMutex) Close() error {
	return nil
}

// Destroy closes the mutex and removes it permanently.
func (m *SemaMutex) Destroy() error {
	return m.s.Destroy()
}

// DestroySemaMutex permanently removes mutex with the given name.
func DestroySemaMutex(name string) error {
	m, err := NewSemaMutex(name, 0, 0666)
	if err != nil {
		if os.IsNotExist(errors.Cause(err)) {
			return nil
		}
		return errors.Wrap(err, "failed to open sema mutex")
	}
	return m.Destroy()
}
