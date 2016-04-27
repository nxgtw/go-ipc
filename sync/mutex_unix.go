// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package sync

import (
	"os"

	"bitbucket.org/avd/go-ipc"
)

// SemaMutex is a semaphore-based mutex for unix.
type SemaMutex struct {
	s *Semaphore
}

// NewSemaMutex creates a new mutex.
func NewSemaMutex(name string, mode int, perm os.FileMode) (*SemaMutex, error) {
	s, err := NewSemaphore(name, mode, perm, 1)
	if err != nil {
		return nil, err
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
	m, err := NewSemaMutex(name, ipc.O_OPEN_ONLY, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return err
	}
	return m.Destroy()
}
