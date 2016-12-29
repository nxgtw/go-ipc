// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build freebsd linux

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
	name   string
	region *mmf.MemoryRegion
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
	state := allocator.ByteSliceData(region.Data())
	result := &event{
		lwe:    newLightweightEvent(state, &futex{ptr: state}),
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
	return e.region.Close()
}

func (e *event) destroy() error {
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
