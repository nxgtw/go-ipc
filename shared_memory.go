// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"os"
	"runtime"
)

const (
	SHM_CREATE = 1 << iota
	SHM_CREATE_ONLY
	SHM_OPEN_ONLY
	SHM_READ
	SHM_RW
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
// size - object size
// mode - open mode. see SHM_* constants
// perm - file's mode and permission bits.
func NewMemoryObject(name string, mode int, perm os.FileMode) (*MemoryObject, error) {
	impl, err := newMemoryObjectImpl(name, mode, perm)
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
