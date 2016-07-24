// Copyright 2016 Aleksandr Demakin. All rights reserved.

//+build linux

package sync

import (
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	"github.com/pkg/errors"
)

// Cond is a named interprocess condition variable.
type Cond struct {
	L      IPCLocker
	name   string
	region *mmf.MemoryRegion
	waiter *inplaceWaiter
}

// NewCond returns new interprocess condvar.
//	name - unique condvar name.
//	flag - a combination of open flags from 'os' package.
//	perm - object's permission bits.
//	l - a locker, associated with the shared resource.
func NewCond(name string, flag int, perm os.FileMode, l IPCLocker) (*Cond, error) {
	openFlags := common.FlagsForOpen(flag)
	// create a shared memory object for the queue.
	obj, created, err := shm.NewMemoryObjectSize(condSharedStateName(name), openFlags, perm, int64(inplaceMutexSize))
	if err != nil {
		return nil, errors.Wrap(err, "cond: failed to open/create shm object")
	}
	defer obj.Close()

	result := &Cond{L: l, name: name}

	defer func() {
		if err == nil {
			return
		}
		if result.region != nil {
			result.region.Close()
		}
		if created {
			obj.Destroy()
		}
	}()

	// mmap memory object.
	result.region, err = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, inplaceMutexSize)
	if err != nil {
		return nil, errors.Wrap(err, "cond: failed to create new shm region")
	}

	result.waiter = newInplaceWaiter(allocator.ByteSliceData(result.region.Data()))

	return result, nil
}

func (c *Cond) Signal() {
	c.waiter.add(1)
	_, err := c.waiter.wake(1)
	if err != nil {
		panic(err)
	}
}

func (c *Cond) Broadcast() {
	c.waiter.add(1)
	_, err := c.waiter.wakeAll()
	if err != nil {
		panic(err)
	}
}

func (c *Cond) Wait() {
	seq := *c.waiter.addr()
	c.L.Unlock()
	if err := c.waiter.wait(seq, time.Duration(-1)); err != nil {
		panic(err)
	}
	c.L.Lock()
}

func (c *Cond) WaitTimeout(timeout time.Duration) {
	seq := *c.waiter.addr()
	c.L.Unlock()
	if err := c.waiter.wait(seq, timeout); err != nil && !common.IsTimeoutErr(err) {
		panic(err)
	}
	c.L.Lock()
}

// Close releases resources of the cond's shared state.
func (c *Cond) Close() error {
	var result error
	if err := c.region.Close(); err != nil {
		result = errors.Wrap(err, "failed to close waiters list memory region")
	}
	return result
}

// Destroy permanently removes condvar.
func (c *Cond) Destroy() error {
	var result error
	if err := c.Close(); err != nil {
		result = errors.Wrap(err, "destroy failed")
	}
	if err := shm.DestroyMemoryObject(condSharedStateName(c.name)); err != nil {
		result = errors.Wrap(err, "failed to close waiters list memory object")
	}
	return result
}

func DestroyCond(name string) error {
	return shm.DestroyMemoryObject(condSharedStateName(name))
}

func condSharedStateName(name string) string {
	return name + ".shared"
}
