// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

// this is to ensure, that all implementations of shm-related structs
// satisfy the same minimal interface
var (
	_ iSharedMemoryObject = &MemoryObject{}
	_ iSharedMemoryRegion = &MemoryRegion{}
	_ MappableHandle      = &MemoryObject{}
)

type iSharedMemoryObject interface {
	Name() string
	Size() int64
	Truncate(size int64) error
	Close() error
	Destroy() error
}

type iSharedMemoryRegion interface {
	Data() []byte
	Size() int
	Flush() error
	Close() error
}
