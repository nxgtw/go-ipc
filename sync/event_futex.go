// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux freebsd

package sync

import (
	"os"
	"sync/atomic"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	"github.com/pkg/errors"
)

type event struct {
	name   string
	region *mmf.MemoryRegion
	waiter *inplaceWaiter
}

func newEvent(name string, flag int, perm os.FileMode, initial bool) (*event, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}
	internalName := eventName(name)
	obj, created, resultErr := shm.NewMemoryObjectSize(internalName, flag, perm, int64(inplaceWaiterSize))
	if resultErr != nil {
		return nil, errors.Wrap(resultErr, "failed to create shm object")
	}
	var region *mmf.MemoryRegion
	defer func() {
		obj.Close()
		if resultErr == nil {
			return
		}
		if region != nil {
			region.Close()
		}
		if created {
			obj.Destroy()
		}
	}()
	if region, resultErr = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, inplaceWaiterSize); resultErr != nil {
		return nil, errors.Wrap(resultErr, "failed to create shm region")
	}
	waiter := newInplaceWaiter(allocator.ByteSliceData(region.Data()))
	if created {
		if initial {
			*waiter.addr() = 1
		} else {
			*waiter.addr() = 0
		}
	}
	return &event{waiter: waiter, name: name, region: region}, nil
}

func (e *event) set() {
	*e.waiter.addr() = 1
	if _, err := e.waiter.wake(1); err != nil {
		panic(err)
	}
}

func (e *event) wait() {
	for {
		if atomic.CompareAndSwapUint32(e.waiter.addr(), 1, 0) {
			return
		}
		if err := e.waiter.wait(0, time.Duration(-1)); err != nil {
			panic(err)
		}
	}
}

func (e *event) waitTimeout(timeout time.Duration) bool {
	for {
		if atomic.CompareAndSwapUint32(e.waiter.addr(), 1, 0) {
			return true
		}
		if err := e.waiter.wait(0, timeout); err != nil {
			if common.IsTimeoutErr(err) {
				return false
			}
			panic(err)
		}
	}
}

func (e *event) close() error {
	if e.region == nil {
		return nil
	}
	err := e.region.Close()
	e.region = nil
	e.waiter = nil
	return err
}

func (e *event) destroy() error {
	if e.region == nil {
		return nil
	}
	if err := e.close(); err != nil {
		return errors.Wrap(err, "failed to close shm region")
	}
	return destroyEvent(e.name)
}

func destroyEvent(name string) error {
	err := shm.DestroyMemoryObject(eventName(name))
	if err != nil {
		return errors.Wrap(err, "failed to destroy memory object")
	}
	return nil
}
