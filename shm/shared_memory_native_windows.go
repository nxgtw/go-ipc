// Copyright 2016 Aleksandr Demakin. All rights reserved.

package shm

import (
	"golang.org/x/sys/windows"
)

var (
	_ SharedMemoryObject = (*WindowsNativeMemoryObject)(nil)
)

// WindowsNativeMemoryObject represents a standart windows shm implementation backed by a paging file.
// Can be used to map shared memory regions into the process' address space.
// It does not follow the usual memory object semantics, the following finctions
// are added to satisfy iSharedMemoryObject interface and do not actually do anything:
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

func (obj *WindowsNativeMemoryObject) Size() int64 {
	return 0
}

func (obj *WindowsNativeMemoryObject) Truncate(size int64) error {
	return nil
}

func (obj *WindowsNativeMemoryObject) Close() error {
	return nil
}

func (obj *WindowsNativeMemoryObject) Destroy() error {
	return nil
}
