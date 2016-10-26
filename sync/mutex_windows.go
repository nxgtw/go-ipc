// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

// all implementations must satisfy IPCLocker interface.
var (
	_ IPCLocker = (*EventMutex)(nil)
)

// EventMutex is a mutex built on named windows events.
// It is not possible to use native windows named mutex, because
// goroutines migrate between threads, and windows mutex must
// be released by the same thread it was locked.
type EventMutex struct {
	handle  windows.Handle
	state   *mmf.MemoryRegion
	name    string
	inplace *inplaceMutex
}

// NewEventMutex creates a new event-basedmutex.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
func NewEventMutex(name string, flag int, perm os.FileMode) (*EventMutex, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}
	region, created, err := createWritableRegion(mutexSharedStateName(name, "e"), flag, perm, inplaceMutexSize, cInplaceMutexUnlocked)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create shared state")
	}
	handle, err := openOrCreateEvent(name, flag, 1)
	if err != nil {
		region.Close()
		if created {
			shm.DestroyMemoryObject(mutexSharedStateName(name, "e"))
		}
		return nil, errors.Wrap(err, "failed to open/create event mutex")
	}
	result := &EventMutex{
		handle: handle,
		state:  region,
		name:   name,
	}
	//	result.inplace = newInplaceMutex(allocator.ByteSliceData(region.Data()), result.wake, result.wait)
	return result, nil
}

// Lock locks the mutex. It panics on an error.
func (m *EventMutex) Lock() {
	ev, err := windows.WaitForSingleObject(m.handle, windows.INFINITE)
	if ev != windows.WAIT_OBJECT_0 {
		if err != nil {
			panic(err)
		} else {
			panic(errors.Errorf("invalid wait state for a mutex: %d", ev))
		}
	}
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
func (m *EventMutex) LockTimeout(timeout time.Duration) bool {
	waitMillis := uint32(timeout.Nanoseconds() / 1e6)
	ev, err := windows.WaitForSingleObject(m.handle, waitMillis)
	switch ev {
	case windows.WAIT_OBJECT_0:
		return true
	case windows.WAIT_TIMEOUT:
		return false
	default:
		if err != nil {
			panic(err)
		} else {
			panic(errors.Errorf("invalid wait state for a mutex: %d", ev))
		}
	}
}

// Unlock releases the mutex. It panics on an error.
func (m *EventMutex) Unlock() {
	if err := windows.SetEvent(m.handle); err != nil {
		panic("failed to unlock mutex: " + err.Error())
	}
}

// Close closes event's handle.
func (m *EventMutex) Close() error {
	m.Close()
	return windows.CloseHandle(m.handle)
}

// DestroyEventMutex destroys shared mutex state.
// The event object is destroyed, when its last handle is closed.
func DestroyEventMutex(name string) error {
	return shm.DestroyMemoryObject(mutexSharedStateName(name, "e"))
}
