// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin

package sync

import (
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	"github.com/pkg/errors"
)

type event struct {
	name   string
	region *mmf.MemoryRegion
	waiter *uint32
}

func newEvent(name string, flag int, perm os.FileMode, initial bool) (*event, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}
	internalName := eventName(name)
	obj, created, resultErr := shm.NewMemoryObjectSize(internalName, flag, perm, 4)
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
	if region, resultErr = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, int(spinImplSize)); resultErr != nil {
		return nil, errors.Wrap(resultErr, "failed to create shm region")
	}
	waiter := (*uint32)(allocator.ByteSliceData(region.Data()))
	if created && initial {
		*waiter = 1
	}
	return &event{waiter: waiter, name: name, region: region}, nil
}

func (e *event) set() {
	atomic.StoreUint32(e.waiter, 1)
}

func (e *event) wait() {
	for !atomic.CompareAndSwapUint32(e.waiter, 1, 0) {
		runtime.Gosched()
	}
}

func (e *event) waitTimeout(timeout time.Duration) bool {
	var attempt uint64
	start := time.Now()
	for !atomic.CompareAndSwapUint32(e.waiter, 1, 0) {
		runtime.Gosched()
		if attempt%1000 == 0 { // do not call time.Since too often.
			if timeout >= 0 && time.Since(start) >= timeout {
				return false
			}
		}
		attempt++
	}
	return true
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
