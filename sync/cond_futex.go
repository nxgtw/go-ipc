// Copyright 2016 Aleksandr Demakin. All rights reserved.

//+build freebsd linux

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

// cond is a futex-based convar.
type cond struct {
	L      IPCLocker
	name   string
	region *mmf.MemoryRegion
	waiter *inplaceWaiter
}

func newCond(name string, flag int, perm os.FileMode, l IPCLocker) (*cond, error) {
	openFlags := common.FlagsForOpen(flag)
	// create a shared memory object for the queue.
	obj, created, err := shm.NewMemoryObjectSize(condSharedStateName(name), openFlags, perm, int64(inplaceWaiterSize))
	if err != nil {
		return nil, errors.Wrap(err, "cond: failed to open/create shm object")
	}
	defer obj.Close()

	result := &cond{L: l, name: name}

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
	result.region, err = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, inplaceWaiterSize)
	if err != nil {
		return nil, errors.Wrap(err, "cond: failed to create new shm region")
	}

	result.waiter = newInplaceWaiter(allocator.ByteSliceData(result.region.Data()))

	return result, nil
}

func (c *cond) signal() {
	c.waiter.add(1)
	_, err := c.waiter.wake(1)
	if err != nil {
		panic(err)
	}
}

func (c *cond) broadcast() {
	c.waiter.add(1)
	_, err := c.waiter.wakeAll()
	if err != nil {
		panic(err)
	}
}

func (c *cond) wait() {
	seq := *c.waiter.addr()
	c.L.Unlock()
	if err := c.waiter.wait(seq, time.Duration(-1)); err != nil {
		panic(err)
	}
	c.L.Lock()
}

func (c *cond) waitTimeout(timeout time.Duration) bool {
	seq := *c.waiter.addr()
	var success bool
	c.L.Unlock()
	if err := c.waiter.wait(seq, timeout); err == nil {
		success = true
	} else if !common.IsTimeoutErr(err) {
		panic(err)
	}
	c.L.Lock()
	return success
}

func (c *cond) close() error {
	var result error
	if err := c.region.Close(); err != nil {
		result = errors.Wrap(err, "failed to close waiters list memory region")
	}
	return result
}

func (c *cond) destroy() error {
	var result error
	if err := c.close(); err != nil {
		result = errors.Wrap(err, "destroy failed")
	}
	if err := shm.DestroyMemoryObject(condSharedStateName(c.name)); err != nil {
		result = errors.Wrap(err, "failed to close waiters list memory object")
	}
	return result
}

func destroyCond(name string) error {
	return shm.DestroyMemoryObject(condSharedStateName(name))
}

func condSharedStateName(name string) string {
	return name + ".shared"
}
