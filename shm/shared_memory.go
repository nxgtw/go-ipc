// Copyright 2015 Aleksandr Demakin. All rights reserved.

package shm

import (
	"os"
	"runtime"

	"bitbucket.org/avd/go-ipc"
)

// SharedMemoryObject is an interface, which must be implemented
// by ant implemetation of an object used for mapping into memory.
type SharedMemoryObject interface {
	Size() int64
	Truncate(size int64) error
	Close() error
	Destroy() error
	ipc.Mappable
}

// MemoryObject represents an object which can be used to
// map shared memory regions into the process' address space.
type MemoryObject struct {
	*memoryObject
}

// NewMemoryObject creates a new shared memory object.
//	name - a name of the object. should not contain '/' and exceed 255 symbols
//	size - object size
//	mode - open mode. see ipc.O_* constants
//	perm - file's mode and permission bits.
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

// Destroy closes the object and removes it permanently.
func (obj *MemoryObject) Destroy() error {
	return obj.memoryObject.Destroy()
}

// Name returns the name of the object as it was given to NewMemoryObject().
func (obj *MemoryObject) Name() string {
	return obj.memoryObject.Name()
}

// Close closes object's underlying file object.
// Darwin: a call to Close() causes invalid argument error,
// if the object was not truncated. So, in this case we return nil as an error.
func (obj *MemoryObject) Close() error {
	return obj.memoryObject.Close()
}

// Truncate resizes the shared memory object.
// Darwin: it is possible to truncate an object only once after it was created.
// Darwin: the size should be divisible by system page size,
// otherwise the size is set to the nearest page size devider greater, then the given size.
func (obj *MemoryObject) Truncate(size int64) error {
	return obj.memoryObject.Truncate(size)
}

// Size returns the current object size.
// Darwin: it may differ from the size passed passed to Truncate.
func (obj *MemoryObject) Size() int64 {
	return obj.memoryObject.Size()
}

// Fd returns a descriptor of the object's underlying file object.
func (obj *MemoryObject) Fd() uintptr {
	return obj.memoryObject.Fd()
}

// DestroyMemoryObject permanently removes given memory object.
func DestroyMemoryObject(name string) error {
	return destroyMemoryObject(name)
}
