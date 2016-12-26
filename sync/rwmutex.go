// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/helper"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"

	"github.com/pkg/errors"
)

// all implementations must satisfy at least IPCLocker interface.
var (
	_ IPCLocker = (*RWMutex)(nil)
)

// RWMutex is a mutex, that can be held by any number of readers or one writer.
type RWMutex struct {
	lwm    *lwRWMutex
	region *mmf.MemoryRegion
	wR, wW waitWaker
	name   string
}

// NewRWMutex returns new RWMutex
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
func NewRWMutex(name string, flag int, perm os.FileMode) (*RWMutex, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}
	region, created, err := helper.CreateWritableRegion(mutexSharedStateName(name, "rw"), flag, perm, lwRWMStateSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create shared state")
	}
	result := &RWMutex{region: region, name: name}
	if result.wR, result.wW, err = makeRWMWaiters(name, flag, perm); err != nil {
		region.Close()
		if created {
			shm.DestroyMemoryObject(mutexSharedStateName(name, "rw"))
		}
		return nil, err
	}
	result.lwm = newRWLightweightMutex(allocator.ByteSliceData(region.Data()), result.wR, result.wW)
	if created {
		result.lwm.init()
	}
	return result, nil
}

// Lock locks the mutex exclusively. It panics on an error.
func (rw *RWMutex) Lock() {
	rw.lwm.lock()
}

// Unlock releases the mutex. It panics on an error, or if the mutex is not locked.
func (rw *RWMutex) Unlock() {
	rw.lwm.unlock()
}

// RLock locks the mutex for reading. It panics on an error.
func (rw *RWMutex) RLock() {
	rw.lwm.rlock()
}

// RUnlock desceases the number of mutex's readers. If it becomes 0, writers (if any) can proceed.
// It panics on an error, or if the mutex is not locked.
func (rw *RWMutex) RUnlock() {
	rw.lwm.runlock()
}

// Close closes shared state of the mutex.
func (rw *RWMutex) Close() error {
	e1, e2 := closeRWWaiters(rw.wR, rw.wW), rw.region.Close()
	if e1 != nil {
		return e1
	}
	if e2 != nil {
		return e2
	}
	return nil
}

// Destroy closes the mutex and removes it permanently.
func (rw *RWMutex) Destroy() error {
	if err := rw.Close(); err != nil {
		return errors.Wrap(err, "failed to close shared state")
	}
	return DestroyRWMutex(rw.name)
}

// DestroyRWMutex permanently removes mutex with the given name.
func DestroyRWMutex(name string) error {
	e1 := shm.DestroyMemoryObject(mutexSharedStateName(name, "rw"))
	e2 := destroyRWWaiters(name)
	if e1 != nil {
		return errors.Wrap(e1, "failed to destroy shared state")
	}
	if e2 != nil {
		return e2
	}
	return nil
}

// RLocker returns a Locker interface that implements
// the Lock and Unlock methods by calling rw.RLock and rw.RUnlock.
func (rw *RWMutex) RLocker() IPCLocker {
	return (*rlocker)(rw)
}

type rlocker RWMutex

func (r *rlocker) Lock()        { (*RWMutex)(r).RLock() }
func (r *rlocker) Unlock()      { (*RWMutex)(r).RUnlock() }
func (r *rlocker) Close() error { return (*RWMutex)(r).Close() }

func makeRWMWaiters(name string, flag int, perm os.FileMode) (waitWaker, waitWaker, error) {
	rSema, err := NewSemaphore(name+".rs", flag, perm, 0)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create r/sema")
	}
	wSema, err := NewSemaphore(name+".ws", flag, perm, 0)
	if err != nil {
		rSema.Close()
		DestroySemaphore(name + ".rs")
		return nil, nil, errors.Wrap(err, "failed to create w/sema")
	}
	return newSemaWaiter(rSema), newSemaWaiter(wSema), nil
}

func closeRWWaiters(wR, wW waitWaker) error {
	sR, sW := wR.(*semaWaiter).s, wW.(*semaWaiter).s
	e1, e2 := sR.Close(), sW.Close()
	if e1 != nil {
		return errors.Wrap(e1, "failed to close r/sema")
	}
	if e2 != nil {
		return errors.Wrap(e2, "failed to close w/sema")
	}
	return nil
}

func destroyRWWaiters(name string) error {
	e1, e2 := DestroySemaphore(name+".rs"), DestroySemaphore(name+".ws")
	if e1 != nil {
		return errors.Wrap(e1, "failed to deatroy r/sema")
	}
	if e2 != nil {
		return errors.Wrap(e2, "failed to deatroy w/sema")
	}
	return nil
}
