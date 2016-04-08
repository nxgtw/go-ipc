// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"os"
)

// Mutex is an interprocess readwrite lock object
type Mutex struct {
	*mutex
}

// NewMutex creates a new readwrite mutex
//	name - object name
//	mode - object creation mode. must be one of the following:
//		O_OPEN_OR_CREATE
//		O_CREATE_ONLY
//		O_OPEN_ONLY
func NewMutex(name string, mode int, perm os.FileMode) (*Mutex, error) {
	if !checkMutexOpenMode(mode) {
		return nil, fmt.Errorf("invalid open mode")
	}
	m, err := newMutex(name, mode, perm)
	if err != nil {
		return nil, err
	}
	result := &Mutex{m}
	return result, nil
}

// Lock locks m. If the lock is already in use, the calling goroutine blocks until the mutex is available.
// Lock panics on any error.
func (m *Mutex) Lock() {
	m.mutex.Lock()
}

// Unlock unlocks m. Unlock panics on any error.
func (m *Mutex) Unlock() {
	m.mutex.Unlock()
}

// Close closes current instance so that it cannot be used anymore.
func (m *Mutex) Close() error {
	return m.mutex.Close()
}
