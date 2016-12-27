// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build freebsd

package sync

import (
	"os"
	"sync/atomic"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"
	"bitbucket.org/avd/go-ipc/internal/helper"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	"github.com/pkg/errors"
)

type event struct {
	name   string
	region *mmf.MemoryRegion
	ftx    *futex
}

func newEvent(name string, flag int, perm os.FileMode, initial bool) (*event, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}

	region, created, err := helper.CreateWritableRegion(eventName(name), flag, perm, futexSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create shared state")
	}

	result := &event{
		ftx:    &futex{allocator.ByteSliceData(region.Data())},
		name:   name,
		region: region,
	}
	if created {
		if initial {
			*result.ftx.addr() = 1
		} else {
			*result.ftx.addr() = 0
		}
	}
	return result, nil
}

func (e *event) set() {
	*e.ftx.addr() = 1
	if _, err := e.ftx.wake(1); err != nil {
		panic(err)
	}
}

func (e *event) wait() {
	e.waitTimeout(-1)
}

func (e *event) waitTimeout(timeout time.Duration) bool {
	for {
		if atomic.CompareAndSwapInt32(e.ftx.addr(), 1, 0) {
			return true
		}
		if err := e.ftx.wait(0, timeout); err != nil {
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
	e.ftx = nil
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
