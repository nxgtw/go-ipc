// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"time"
)

// Event is a synchronization primitive used for notification.
// If it is signaled by a call to Set(), it'll stay in this state,
// unless someone calls Wait(). After it the event is reset into non-signaled state.
type Event event

// NewEvent creates a new interprocess event.
// It uses the default implementation on the current platform.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
//	initial - if true, the event will be set after creation.
func NewEvent(name string, flag int, perm os.FileMode, initial bool) (*Event, error) {
	e, err := newEvent(name, flag, perm, initial)
	if err != nil {
		return nil, err
	}
	return (*Event)(e), nil
}

// Set sets the specified event object to the signaled state.
func (e *Event) Set() {
	(*event)(e).set()
}

// Wait waits for the event to be signaled.
func (e *Event) Wait() {
	(*event)(e).wait()
}

// WaitTimeout waits until the event is signaled or the timeout elapses.
func (e *Event) WaitTimeout(timeout time.Duration) bool {
	return (*event)(e).waitTimeout(timeout)
}

// Close closes the event.
func (e *Event) Close() error {
	return (*event)(e).close()
}

// Destroy permanently destroys the event.
func (e *Event) Destroy() error {
	return (*event)(e).destroy()
}

// DestroyEvent permanently destroys an event with the given name.
func DestroyEvent(name string) error {
	return destroyEvent(name)
}

func eventName(baseName string) string {
	return baseName + ".ev"
}
