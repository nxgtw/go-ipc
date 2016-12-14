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
	ftx    *futex
}

func newEvent(name string, flag int, perm os.FileMode, initial bool) (*event, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}
	internalName := eventName(name)
	obj, created, resultErr := shm.NewMemoryObjectSize(internalName, flag, perm, int64(futexSize))
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
	if region, resultErr = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, futexSize); resultErr != nil {
		return nil, errors.Wrap(resultErr, "failed to create shm region")
	}
	ftx := &futex{allocator.ByteSliceData(region.Data())}
	if created {
		if initial {
			*ftx.addr() = 1
		} else {
			*ftx.addr() = 0
		}
	}
	return &event{ftx: ftx, name: name, region: region}, nil
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
		if atomic.CompareAndSwapUint32(e.ftx.addr(), 1, 0) {
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
