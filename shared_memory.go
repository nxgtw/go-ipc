// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"os"
	"runtime"
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
// name - a name of the object. should not contain '/' and exceed 255 symbols
// size - object size
// mode - open mode. see O_* constants
// perm - file's mode and permission bits.
func NewMemoryObject(name string, mode int, perm os.FileMode) (*MemoryObject, error) {
	impl, err := newMemoryObjectImpl(name, mode, perm)
	if err != nil {
		return nil, err
	}
	result := &MemoryObject{impl}
	runtime.SetFinalizer(impl, func(object interface{}) {
		memObject := object.(*memoryObjectImpl)
		memObject.Close()
	})
	return result, nil
}

// Returns a new shared memory region.
// object - an object containing a descriptor of the file, which can be mmaped
// size - object size
// mode - open mode. see O_* constants
// offset - offset in bytes from the beginning of the mmaped file
// size - region size
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
