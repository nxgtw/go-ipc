// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"errors"
	"os"
	"time"
)

var (
	// MaxCondWaiters is the maximum length of the waiting queue for this type of a cond.
	// This limit is actual for waitlist-based condvars, currently on windows and darwin.
	// If this limit is exceeded, Wait/WaitTimeout will panic with ErrTooManyWaiters.
	MaxCondWaiters = 128
	// ErrTooManyWaiters is an error, that indicates, that the waiting queue is full.
	ErrTooManyWaiters = errors.New("waiters limit has been reached")
)

// Cond is a named interprocess condition variable.
type Cond cond

// NewCond returns new interprocess condvar.
//	name - unique condvar name.
//	flag - a combination of open flags from 'os' package.
//	perm - object's permission bits.
//	l - a locker, associated with the shared resource.
func NewCond(name string, flag int, perm os.FileMode, l IPCLocker) (*Cond, error) {
	c, err := newCond(name, flag, perm, l)
	if err != nil {
		return nil, err
	}
	return (*Cond)(c), nil
}

// Signal wakes one waiter.
func (c *Cond) Signal() {
	(*cond)(c).signal()
}

// Broadcast wakes all waiters.
func (c *Cond) Broadcast() {
	(*cond)(c).broadcast()
}

// Wait waits for the condvar to be signaled.
func (c *Cond) Wait() {
	(*cond)(c).wait()
}

// WaitTimeout waits for the condvar to be signaled for not longer, than timeout.
func (c *Cond) WaitTimeout(timeout time.Duration) bool {
	return (*cond)(c).waitTimeout(timeout)
}

// Close releases resources of the cond's shared state.
func (c *Cond) Close() error {
	return (*cond)(c).close()
}

// Destroy permanently removes condvar.
func (c *Cond) Destroy() error {
	return (*cond)(c).destroy()
}

// DestroyCond permanently removes condvar with the given name.
func DestroyCond(name string) error {
	return destroyCond(name)
}
