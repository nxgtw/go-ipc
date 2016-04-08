// Copyright 2015 Aleksandr Demakin. All rights reserved.

package shm

import (
	"bitbucket.org/avd/go-ipc"
)

// this is to ensure, that all implementations of shm-related structs
// satisfy the same minimal interface
var (
	_ iSharedMemoryObject = (*MemoryObject)(nil)
	_ iSharedMemoryRegion = (*ipc.MemoryRegion)(nil)
)

type iSharedMemoryObject interface {
	Name() string
	Size() int64
	Truncate(size int64) error
	Close() error
	Destroy() error
	ipc.MappableHandle
}

type iSharedMemoryRegion interface {
	Data() []byte
	Size() int
	Flush(async bool) error
	Close() error
}
