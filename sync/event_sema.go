// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux

package sync

import (
	"os"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/helper"
	"bitbucket.org/avd/go-ipc/mmf"
	"github.com/pkg/errors"
)

type event struct {
	s      *Semaphore
	region *mmf.MemoryRegion
	name   string
	lve    *lwEvent
}

func newEvent(name string, flag int, perm os.FileMode, initial bool) (*event, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}

	region, created, err := helper.CreateWritableRegion(eventName(name), flag, perm, lweStateSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create shared state")
	}
	result := &event{
		waiter: (*uint32)(allocator.ByteSliceData(region.Data())),
		name:   name,
		region: region,
	}

	if created && initial {
		*result.waiter = 1
	}
	return result, nil
}
