// Copyright 2016 Aleksandr Demakin. All rights reserved.

//+build windows darwin

package sync

import (
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/array"
	"bitbucket.org/avd/go-ipc/internal/common"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	"github.com/pkg/errors"
)

// cond is a condvar implemented as a shared queue of waiters.
type cond struct {
	L             IPCLocker
	listLock      IPCLocker
	name          string
	waitersRegion *mmf.MemoryRegion
	waiters       *array.SharedArray
}

func newCond(name string, flag int, perm os.FileMode, l IPCLocker) (*cond, error) {
	size := array.CalcSharedArraySize(MaxCondWaiters, condWaiterSize)
	openFlags := common.FlagsForOpen(flag)
	// create a shared memory object for the queue.
	obj, created, err := shm.NewMemoryObjectSize(condSharedStateName(name), openFlags, perm, int64(size))
	if err != nil {
		return nil, errors.Wrap(err, "cond: failed to open/create shm object")
	}

	result := &cond{L: l, name: name}

	defer func() {
		obj.Close()
		if err == nil {
			return
		}
		condCleanup(result, name, obj, created)
	}()

	// mmap memory object.
	result.waitersRegion, err = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, size)
	if err != nil {
		return nil, errors.Wrap(err, "cond: failed to create new shm region")
	}

	// cleanup previous mutex instances. it could be useful in a case,
	// when previous mutex owner crashed, and the mutex is in incosistient state.
	if created {
		if err = DestroyMutex(condMutexName(name)); err != nil {
			return nil, errors.Wrap(err, "cond: failed to access a locker")
		}
	}

	result.listLock, err = NewMutex(condMutexName(name), flag, perm)
	if err != nil {
		return nil, errors.Wrap(err, "cond: failed to obtain internal lock")
	}

	rawData := allocator.ByteSliceData(result.waitersRegion.Data())
	if created {
		result.waiters = array.NewSharedArray(rawData, MaxCondWaiters, condWaiterSize)
	} else {
		result.waiters = array.OpenSharedArray(rawData)
	}
	return result, nil
}

func (c *cond) wait() {
	c.doWait(time.Duration(-1))
}

func (c *cond) waitTimeout(timeout time.Duration) bool {
	return c.doWait(timeout)
}

func (c *cond) signal() {
	c.listLock.Lock()
	defer c.listLock.Unlock()
	c.signalN(1)
}

func (c *cond) broadcast() {
	c.listLock.Lock()
	defer c.listLock.Unlock()
	c.signalN(c.waiters.Len())
}

// signalN wakes n waiters. Must be run with the list mutex locked.
func (c *cond) signalN(count int) {
	var signaled int
	for i := 0; i < c.waiters.Len() && signaled < count; i++ {
		if w := openWaiter(c.waiters.AtPointer(i)); w.signal() {
			signaled++
		}
	}
}

func (c *cond) doWait(timeout time.Duration) bool {
	w := c.addToWaitersList()
	// unlock resource locker
	c.L.Unlock()
	result := w.waitTimeout(timeout)
	c.L.Lock()
	c.cleanupWaiter(w)
	return result
}

func (c *cond) cleanupWaiter(w *waiter) {
	c.listLock.Lock()
	defer c.listLock.Unlock()
	for i := 0; i < c.waiters.Len(); i++ {
		if w.isSame(c.waiters.AtPointer(i)) {
			w.destroy()
			c.waiters.RemoveAt(i)
			return
		}
	}
}

func (c *cond) addToWaitersList() *waiter {
	c.listLock.Lock()
	defer c.listLock.Unlock()
	if c.waiters.Len() >= MaxCondWaiters {
		panic(ErrTooManyWaiters)
	}
	c.waiters.PushBack(nil)
	return newWaiter(c.waiters.AtPointer(c.waiters.Len() - 1))
}

func (c *cond) close() error {
	var result error
	if err := c.listLock.Close(); err != nil {
		result = errors.Wrap(err, "failed to close waiters list locker")
	}
	if err := c.waitersRegion.Close(); err != nil {
		result = errors.Wrap(err, "failed to close waiters list memory region")
	}
	return result
}

func (c *cond) destroy() error {
	var result error
	if err := c.close(); err != nil {
		result = errors.Wrap(err, "destroy failed")
	}
	if err := DestroyMutex(condMutexName(c.name)); err != nil {
		result = errors.Wrap(err, "failed to destroy waiters list locker")
	}
	if err := shm.DestroyMemoryObject(condSharedStateName(c.name)); err != nil {
		result = errors.Wrap(err, "failed to destroy waiters list memory object")
	}
	return result
}

func condMutexName(name string) string {
	return name + ".m"
}

func condSharedStateName(name string) string {
	return name + ".st"
}

func condCleanup(result *cond, name string, obj shm.SharedMemoryObject, created bool) {
	if result.waitersRegion != nil {
		result.waitersRegion.Close()
	}
	if result.listLock != nil {
		result.listLock.Close()
		DestroyMutex(condMutexName(name))
	}
	if created {
		obj.Destroy()
	}
}

func destroyCond(name string) error {
	result := DestroyMutex(condMutexName(name))
	if result != nil {
		result = errors.Wrap(result, "failed to destroy cond list mutex")
	}
	if err := shm.DestroyMemoryObject(condSharedStateName(name)); err != nil {
		if result == nil {
			result = errors.Wrap(err, "failed to destroy shared cond state")
		}
	}
	return result
}
