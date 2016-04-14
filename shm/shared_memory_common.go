// Copyright 2015 Aleksandr Demakin. All rights reserved.

package shm

import (
	"bitbucket.org/avd/go-ipc"
)

// this is to ensure, that all implementations of shm-related structs
// satisfy the same minimal interface
var (
	_ SharedMemoryObject  = (*MemoryObject)(nil)
	_ iSharedMemoryRegion = (*ipc.MemoryRegion)(nil)
)

type SharedMemoryObject interface {
	Size() int64
	Truncate(size int64) error
	Close() error
	Destroy() error
	ipc.Mappable
}

type iSharedMemoryRegion interface {
	Data() []byte
	Size() int
	Flush(async bool) error
	Close() error
}
