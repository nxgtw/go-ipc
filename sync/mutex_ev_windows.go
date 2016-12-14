// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"syscall"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

// all implementations must satisfy IPCLocker interface.
var (
	_          IPCLocker = (*EventMutex)(nil)
	timeoutErr           = os.NewSyscallError("WaitForSingleObject", syscall.Errno(common.ERROR_TIMEOUT))
)

// EventMutex is a mutex built on named windows events.
// It is not possible to use native windows named mutex, because
// goroutines migrate between threads, and windows mutex must
// be released by the same thread it was locked.
type EventMutex struct {
	handle windows.Handle
	state  *mmf.MemoryRegion
	name   string
	lwm    *lwMutex
}

// NewEventMutex creates a new event-basedmutex.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
func NewEventMutex(name string, flag int, perm os.FileMode) (*EventMutex, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}
	region, created, err := createWritableRegion(mutexSharedStateName(name, "e"), flag, perm, lwmCellSize, nil)
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
		lwm:    newLightweightMutex(allocator.ByteSliceData(region.Data()), &eventWaiter{handle: handle}),
	}
	if created {
		result.lwm.init()
	}
	return result, nil
}

// Lock locks the mutex. It panics on an error.
func (m *EventMutex) Lock() {
	m.lwm.lock()
}

// TryLock makes one attempt to lock the mutex. It return true on succeess and false otherwise.
func (m *EventMutex) TryLock() bool {
	return m.lwm.tryLock()
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
func (m *EventMutex) LockTimeout(timeout time.Duration) bool {
	return m.lwm.lockTimeout(timeout)
}

// Unlock releases the mutex. It panics on an error.
func (m *EventMutex) Unlock() {
	m.lwm.unlock()
}

// Close closes event's handle.
func (m *EventMutex) Close() error {
	m.state.Close()
	return windows.CloseHandle(m.handle)
}

// DestroyEventMutex destroys shared mutex state.
// The event object is destroyed, when its last handle is closed.
func DestroyEventMutex(name string) error {
	return shm.DestroyMemoryObject(mutexSharedStateName(name, "e"))
}

type eventWaiter struct {
	handle windows.Handle
}

func (e *eventWaiter) wake(uint32) (int, error) {
	if err := windows.SetEvent(e.handle); err != nil {
		panic("failed to unlock mutex: " + err.Error())
	}
	return 1, nil
}

func (e *eventWaiter) wait(unused uint32, timeout time.Duration) error {
	waitMillis := uint32(windows.INFINITE)
	if timeout >= 0 {
		waitMillis = uint32(timeout.Nanoseconds() / 1e6)
	}
	ev, err := windows.WaitForSingleObject(e.handle, waitMillis)
	switch ev {
	case windows.WAIT_OBJECT_0:
		return nil
	case windows.WAIT_TIMEOUT:
		return timeoutErr
	default:
		if err != nil {
			panic(err)
		} else {
			panic(errors.Errorf("invalid wait state for a mutex: %d", ev))
		}
	}
}
