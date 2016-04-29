// Copyright 2016 Aleksandr Demakin. All rights reserved.

package shm

import (
	"fmt"

	"golang.org/x/sys/windows"
)

var (
	_ SharedMemoryObject = (*WindowsNativeMemoryObject)(nil)
)

// WindowsNativeMemoryObject represents a standart windows shm implementation backed by a paging file.
// Can be used to map shared memory regions into the process' address space.
// It does not follow the usual memory object semantics, the following finctions
// are added to satisfy SharedMemoryObject interface and do not actually do anything:
//	Size
//	Truncate
//	Close
//	Destroy
type WindowsNativeMemoryObject struct {
	name string
}

func NewWindowsNativeMemoryObject(name string) *WindowsNativeMemoryObject {
	return &WindowsNativeMemoryObject{name: name}
}

// Name returns the name of the object as it was given to NewWindowsNativeMemoryObject().
func (obj *WindowsNativeMemoryObject) Name() string {
	return obj.name
}

// Fd returns windows.InvalidHandle so that the paging file is used for mapping.
func (obj *WindowsNativeMemoryObject) Fd() uintptr {
	return uintptr(windows.InvalidHandle)
}

// Size always returns 0, as the memory is backed by a paging file.
func (obj *WindowsNativeMemoryObject) Size() int64 {
	return 0
}

// Truncate returns an error. You can specify the size of a mapped region, not the object.
func (obj *WindowsNativeMemoryObject) Truncate(size int64) error {
	return fmt.Errorf("truncate cannot be done on windows shared memory")
}

// Close returns an error. It is not supported for windows shared memory.
func (obj *WindowsNativeMemoryObject) Close() error {
	return fmt.Errorf("close cannot be done on windows shared memory")
}

// Destroy returns an error. It is not supported for windows shared memory.
func (obj *WindowsNativeMemoryObject) Destroy() error {
	return fmt.Errorf("destroy cannot be done on windows shared memory")
}
