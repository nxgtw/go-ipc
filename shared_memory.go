// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"runtime"
)

const (
	SHM_OPEN_READ = 1 << iota
	SHM_OPEN_WRITE
	SHM_OPEN_CREATE
	SHM_OPEN_CREATE_IF_NOT_EXISTS
)

// MemoryObject represents an object which can be used to
// map shared memory regions into the process' address space
type MemoryObject struct {
	*memoryObjectImpl
}

// MemoryRegion is a mmapped area of a memory object
type MemoryRegion struct {
	*memoryRegionImpl
}

// MappableHandle is an object, which can return a handle,
// that can be used as a file descriptor for mmap
type MappableHandle interface {
	Fd() int
}

// Returns a new shared memory object.
// name - a name of the region. should not contain '/' and exceed 255 symbols
// mode - open mode. see SHM_OPEN* constants
// flags - a set of (probably, platform-specific) flags. see SHM_FLAG_* constants
func NewMemoryObject(name string, size int64, mode int, flags uint32) (*MemoryObject, error) {
	impl, err := newMemoryObjectImpl(name, size, mode, flags)
	if err != nil {
		return nil, err
	}
	result := &MemoryObject{impl}
	runtime.SetFinalizer(impl, func(object interface{}) {
		region := object.(*memoryObjectImpl)
		region.Close()
	})
	return result, nil
}

func NewMemoryRegion(object MappableHandle, mode int, offset int64, size int) (*MemoryRegion, error) {
	impl, err := newMemoryRegionImpl(object, mode, offset, size)
	if err != nil {
		return nil, err
	}
	result := &MemoryRegion{impl}
	runtime.SetFinalizer(impl, func(object interface{}) {
		region := object.(*memoryRegionImpl)
		region.Close()
	})
	return result, nil
}
