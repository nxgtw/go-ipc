// Copyright 2015 Aleksandr Demakin. All rights reserved.

package shm

import (
	"os"
	"runtime"
)

// MemoryObject represents an object which can be used to
// map shared memory regions into the process' address space
type MemoryObject struct {
	*memoryObject
}

// NewMemoryObject creates a new shared memory object.
// name - a name of the object. should not contain '/' and exceed 255 symbols
// size - object size
// mode - open mode. see O_* constants
// perm - file's mode and permission bits.
func NewMemoryObject(name string, mode int, perm os.FileMode) (*MemoryObject, error) {
	impl, err := newMemoryObject(name, mode, perm)
	if err != nil {
		return nil, err
	}
	result := &MemoryObject{impl}
	runtime.SetFinalizer(impl, func(memObject *memoryObject) {
		memObject.Close()
	})
	return result, nil
}
