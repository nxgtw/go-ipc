// Copyright 2015 Aleksandr Demakin. All rights reserved.

package mmf

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"github.com/pkg/errors"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/windows"
)

func init() {
	g, p := getAllocGranularity(), os.Getpagesize()
	if g >= p {
		mmapOffsetMultiple = int64(g)
	} else {
		mmapOffsetMultiple = int64(p)
	}
}

type memoryRegion struct {
	data       []byte
	size       int
	pageOffset int64
	fileHandle windows.Handle
}

func newMemoryRegion(obj Mappable, mode int, offset int64, size int) (*memoryRegion, error) {
	prot, flags, err := memProtAndFlagsFromMode(mode)
	if err != nil {
		return nil, errors.Wrap(err, "memory region flags check failed")
	}
	if size, err = checkMmapSize(obj, size); err != nil {
		return nil, errors.Wrap(err, "size check failed")
	}
	maxSizeHigh := uint32((offset + int64(size)) >> 32)
	maxSizeLow := uint32((offset + int64(size)) & 0xFFFFFFFF)
	var name *uint16
	// check for a named region for windows native shared memory via a pagefile.
	isForPaging := windows.Handle(obj.Fd()) == windows.InvalidHandle
	if isForPaging {
		if name, err = windows.UTF16PtrFromString(obj.Name()); err != nil {
			return nil, errors.Wrap(err, "invalid object name")
		}
	}

	var handle windows.Handle
	creator := func(create bool) error {
		var err error
		// We need some special handling here for windows native memory.
		// Consider the following situation:
		//	object := NewWindowsNativeMemoryObject("object")
		//	region1, _ :=  ipc.NewMemoryRegion(object, ipc.MEM_READ_ONLY, 0, size)
		//	region2, _ :=  ipc.NewMemoryRegion(object, ipc.MEM_READWRITE, 0, size)
		// 1) the first call creates a mapping file object with PAGE_READONLY protection.
		// 2) the second call fails to create mapping for the same object, as it has been reated readonly.
		// Although the object itself does not permit writing, the first call makes it readonly.
		// Thus, we create a mapping object with PAGE_READWRITE permission, and pass actual permission
		// to MapViewOfFile.
		// We leave prot as is for usual file mapping, as the caller is responsible for
		// prot for original file.
		if create || !isForPaging {
			locProt := prot
			if isForPaging {
				locProt = windows.PAGE_READWRITE
			}
			handle, err = windows.CreateFileMapping(
				windows.Handle(obj.Fd()),
				nil,
				locProt,
				maxSizeHigh,
				maxSizeLow,
				name)
		} else {
			handle, err = openFileMapping(flags, 0, obj.Name())
		}
		return err
	}

	if _, err = common.OpenOrCreate(creator, os.O_CREATE); err != nil {
		return nil, errors.Wrap(err, "create mapping file failed")
	}

	// if we use windows native shared memory, we can't close this handle right now.
	// it will be closed, when Close is called.
	mapHandle := handle
	if !isForPaging {
		defer windows.CloseHandle(handle)
		handle = windows.InvalidHandle
	}
	pageOffset := calcMmapOffsetFixup(offset)
	offset -= pageOffset
	lowOffset := uint32(offset & 0xFFFFFFFF)
	highOffset := uint32(offset >> 32)
	addr, err := windows.MapViewOfFile(mapHandle, flags, highOffset, lowOffset, uintptr(int64(size)+pageOffset))
	if err != nil {
		return nil, errors.Wrap(os.NewSyscallError("MapViewOfFile", err), "failed to mmap file view")
	}
	totalSize := size + int(pageOffset)
	return &memoryRegion{
		data:       allocator.ByteSliceFromUnsafePointer(unsafe.Pointer(addr), totalSize, totalSize),
		size:       size,
		pageOffset: pageOffset,
		fileHandle: handle,
	}, nil
}

func (region *memoryRegion) Close() error {
	runtime.SetFinalizer(region, nil)
	err := windows.UnmapViewOfFile(uintptr(allocator.ByteSliceData(region.data)))
	if err != nil {
		return errors.Wrap(err, "UnmapViewOfFile failed")
	}
	if region.fileHandle != windows.InvalidHandle {
		if err = windows.CloseHandle(region.fileHandle); err != nil {
			return errors.Wrap(err, "CloseHandle failed")
		}
	}
	return nil
}

func (region *memoryRegion) Data() []byte {
	return region.data[region.pageOffset:]
}

func (region *memoryRegion) Size() int {
	return region.size
}

func (region *memoryRegion) Flush(async bool) error {
	err := windows.FlushViewOfFile(uintptr(allocator.ByteSliceData(region.data)), uintptr(len(region.data)))
	if err != nil {
		return errors.Wrap(err, "FlushViewOfFile failed")
	}
	return nil
}

func memProtAndFlagsFromMode(mode int) (prot uint32, flags uint32, err error) {
	switch mode {
	case MEM_READ_ONLY:
		fallthrough
	case MEM_READ_PRIVATE:
		prot = windows.PAGE_READONLY
		flags = windows.FILE_MAP_READ
	case MEM_READWRITE:
		prot = windows.PAGE_READWRITE
		flags = windows.FILE_MAP_WRITE
	case MEM_COPY_ON_WRITE:
		prot = windows.PAGE_WRITECOPY
		flags = windows.FILE_MAP_COPY
	default:
		err = fmt.Errorf("invalid mem region flags %d", mode)
	}
	return
}
