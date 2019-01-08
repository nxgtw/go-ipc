// Copyright 2016 Aleksandr Demakin. All rights reserved.

package shm

import (
	"os"

	"github.com/nxgtw/go-ipc/internal/common"
	"github.com/nxgtw/go-ipc/internal/sys/windows"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

const (
	// O_COPY_ON_WRITE used as a flag for NewWindowsNativeMemoryObject.
	// It is passed to CreateFileMapping and OpebFileMapping calls.
	// Its value was chosen simply not to intersect with O_RDONLY and O_RDWR.
	O_COPY_ON_WRITE = windows.O_NONBLOCK
)

var (
	_ SharedMemoryObject = (*WindowsNativeMemoryObject)(nil)
)

// WindowsNativeMemoryObject represents a standart windows shm implementation backed by a paging file.
// Can be used to map shared memory regions into the process' address space.
// It does not follow the usual memory object semantics, and it is destroyed only when all its handles are closed.
// The following functions were added to satisfy SharedMemoryObject interface and return an error:
//	Truncate
//	Destroy
type WindowsNativeMemoryObject struct {
	name   string
	size   int
	handle windows.Handle
}

// NewWindowsNativeMemoryObject returns a new Windows native shared memory object.
//	name - object name.
//	flag - combination of open flags from 'os' package along with O_COPY_ON_WRITE.
//	size - mapping file size.
func NewWindowsNativeMemoryObject(name string, flag, size int) (*WindowsNativeMemoryObject, error) {
	prot, sysFlags, err := sysProtAndFlagsFromFlag(flag)
	if err != nil {
		return nil, errors.Wrap(err, "windows native shm flags check failed")
	}
	maxSizeHigh := uint32((int64(size)) >> 32)
	maxSizeLow := uint32((int64(size)) & 0xFFFFFFFF)

	var handle windows.Handle
	creator := func(create bool) error {
		if create {
			handle, err = sys.CreateFileMapping(
				windows.InvalidHandle,
				nil,
				prot,
				maxSizeHigh,
				maxSizeLow,
				name)
			if os.IsExist(err) {
				windows.CloseHandle(handle)
			}
		} else {
			handle, err = sys.OpenFileMapping(sysFlags, 0, name)
		}
		return err
	}

	if _, err = common.OpenOrCreate(creator, flag); err != nil {
		return nil, errors.Wrap(err, "create mapping file failed")
	}

	return &WindowsNativeMemoryObject{name: name, handle: handle, size: size}, nil
}

// Name returns the name of the object as it was given to NewWindowsNativeMemoryObject().
func (obj *WindowsNativeMemoryObject) Name() string {
	return obj.name
}

// Fd returns windows.InvalidHandle so that the paging file is used for mapping.
func (obj *WindowsNativeMemoryObject) Fd() uintptr {
	return uintptr(obj.handle)
}

// Size returns mapping fiel size.
func (obj *WindowsNativeMemoryObject) Size() int64 {
	return int64(obj.size)
}

// Close closes mapping file object.
func (obj *WindowsNativeMemoryObject) Close() error {
	if err := windows.CloseHandle(obj.handle); err != nil {
		return errors.Wrap(err, "close handle failed")
	}
	return nil
}

// Truncate returns an error. You can specify the size when calling NewWindowsNativeMemoryObject.
func (obj *WindowsNativeMemoryObject) Truncate(size int64) error {
	return errors.New("truncate cannot be done on windows shared memory")
}

// Destroy returns an error. It is not supported for windows shared memory.
func (obj *WindowsNativeMemoryObject) Destroy() error {
	return errors.New("destroy cannot be done on windows shared memory")
}

// IsNative returns true, indicating, that this object can be mapped without extra call to CreateFileMapping.
func (obj *WindowsNativeMemoryObject) IsNative() bool {
	return true
}

func sysProtAndFlagsFromFlag(flag int) (prot uint32, flags uint32, err error) {
	flag = common.FlagsForAccess(flag)
	switch flag {
	case os.O_RDONLY:
		prot = windows.PAGE_READONLY
		flags = windows.FILE_MAP_READ
	case os.O_RDWR:
		prot = windows.PAGE_READWRITE
		flags = windows.FILE_MAP_WRITE
	case O_COPY_ON_WRITE:
		prot = windows.PAGE_WRITECOPY
		flags = windows.FILE_MAP_COPY
	default:
		err = errors.Errorf("invalid mem region flags %d", flag)
	}
	return
}
