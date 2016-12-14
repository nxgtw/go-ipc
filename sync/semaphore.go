// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"io"
	"os"
	"time"
)

const (
	// CSemMaxVal is the maximum semaphore value,
	// which is guaranteed to be supported on all platforms.
	CSemMaxVal = 32767
)

// Semaphore is a synchronization object with a resource counter,
// which can be used to control access to a common resource.
type Semaphore interface {
	// Signal increments the value of semaphore variable by 1, waking waiting process (if any).
	Signal(count int)
	// Wait decrements the value of semaphore variable by -1, and blocks if the value becomes negative.
	Wait()
	io.Closer
}

// TimedSemaphore is a semaphore, that supports timed waiting.
// Currently supported on all platforms, except darwin.
type TimedSemaphore interface {
	Semaphore
	// WaitTimeout decrements the value of semaphore variable by 1.
	// If the value becomes negative, it waites for not longer than timeout.
	WaitTimeout(timeout time.Duration) bool
}

// NewSemaphore creates new semaphore with the given name.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
//	initial - this value will be added to the semaphore's value, if it was created.
func NewSemaphore(name string, flag int, perm os.FileMode, initial int) (Semaphore, error) {
	result, err := newSemaphore(name, flag, perm, initial)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DestroySemaphore removes the semaphore permanently.
func DestroySemaphore(name string) error {
	return destroySemaphore(name)
}
