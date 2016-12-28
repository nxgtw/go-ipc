// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build windows darwin

package sync

import (
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/helper"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	"github.com/pkg/errors"
)

type event struct {
	s      *Semaphore
	region *mmf.MemoryRegion
	name   string
	lwe    *lwEvent
}

func newEvent(name string, flag int, perm os.FileMode, initial bool) (*event, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}

	region, created, err := helper.CreateWritableRegion(eventName(name), flag, perm, lweStateSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create shared state")
	}
	s, err := NewSemaphore(name, flag, perm, 0)
	if err != nil {
		region.Close()
		if created {
			shm.DestroyMemoryObject(mutexSharedStateName(name, "s"))
		}
		return nil, errors.Wrap(err, "failed to create a semaphore")
	}
	result := &event{
		lwe:    newLightweightEvent(allocator.ByteSliceData(region.Data()), newSemaWaiter(s)),
		name:   name,
		region: region,
	}
	if created {
		result.lwe.init(initial)
	}
	return result, nil
}

func (e *event) set() {
	e.lwe.set()
}

func (e *event) wait() {
	e.waitTimeout(-1)
}

func (e *event) waitTimeout(timeout time.Duration) bool {
	return e.lwe.waitTimeout(timeout)
}

func (e *event) close() error {
	e1, e2 := e.s.Close(), e.region.Close()
	if e1 != nil {
		return errors.Wrap(e1, "failed to close sema")
	}
	if e2 != nil {
		return errors.Wrap(e2, "failed to shared state")
	}
	return nil
}

func (e *event) destroy() error {
	if err := e.close(); err != nil {
		return errors.Wrap(err, "failed to close shm region")
	}
	return destroyEvent(e.name)
}

func destroyEvent(name string) error {
	e1, e2 := shm.DestroyMemoryObject(eventName(name)), destroySemaphore(name)
	if e1 != nil {
		return errors.Wrap(e1, "failed to destroy memory object")
	}
	if e2 != nil {
		return errors.Wrap(e2, "failed to destroy semaphore")
	}
	return nil
}
