// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mq

import (
	"os"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	"github.com/pkg/errors"
)

type sharedMq interface {
	full() bool
	empty() bool
	maxMsgSize() int
	maxSize() int
	top() message
	push(msg message)
	pop(data []byte) int
}

type message struct {
	prio int
	data []byte
}

type sharedMqHdr struct {
	maxQueueSize   int32
	maxMsgSize     int32
	size           int32
	dummyDataArray [0]byte
}

func existingMqParams(name string) (int, int, error) {
	obj, err := shm.NewMemoryObject(name, os.O_RDWR, 0666)
	if err != nil {
		return 0, 0, errors.Wrap(err, "fast mq: failed to open shm object")
	}
	defer obj.Close()
	region, err := mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, int(sharedMqHdrSize))
	if err != nil {
		return 0, 0, errors.Wrap(err, "fast mq: failed to create new shm region")
	}
	defer region.Close()
	header := (*sharedMqHdr)(allocator.ByteSliceData(region.Data()))
	return int(header.maxQueueSize), int(header.maxMsgSize), nil
}
