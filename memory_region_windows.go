// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"golang.org/x/sys/windows"
)

type memoryRegion struct {
	data       []byte
	size       int
	pageOffset int64
}

func newMemoryRegion(obj Mappable, mode int, offset int64, size int) (*memoryRegion, error) {
	prot, flags, err := memProtAndFlagsFromMode(mode)
	if err != nil {
		return nil, err
	}
	if size, err = checkMmapSize(obj, size); err != nil {
		return nil, err
	}
	maxSizeHigh := uint32((offset + int64(size)) >> 32)
	maxSizeLow := uint32((offset + int64(size)) & 0xFFFFFFFF)
	var name *uint16
	// check for a named region for windows native shared memory via a pagefile
	if windows.Handle(obj.Fd()) == windows.InvalidHandle && len(obj.Name()) > 0 {
		if name, err = windows.UTF16PtrFromString(obj.Name()); err != nil {
			return nil, err
		}
	}
	handle, err := windows.CreateFileMapping(windows.Handle(obj.Fd()), nil, prot, maxSizeHigh, maxSizeLow, name)
	if err != nil {
		return nil, os.NewSyscallError("CreateFileMapping", err)
	}
	defer windows.CloseHandle(handle)
	pageOffset := calcMmapOffsetFixup(offset)
	offset -= pageOffset
	lowOffset := uint32(offset & 0xFFFFFFFF)
	highOffset := uint32(offset >> 32)
	addr, err := windows.MapViewOfFile(handle, flags, highOffset, lowOffset, uintptr(int64(size)+pageOffset))
	if err != nil {
		return nil, os.NewSyscallError("MapViewOfFile", err)
	}
	sz := size + int(pageOffset)
	return &memoryRegion{allocator.ByteSliceFromUnsafePointer(unsafe.Pointer(addr), sz, sz), size, pageOffset}, nil
}

func (region *memoryRegion) Close() error {
	runtime.SetFinalizer(region, nil)
	return windows.UnmapViewOfFile(uintptr(unsafe.Pointer(&region.data[0])))
}

func (region *memoryRegion) Data() []byte {
	return region.data[region.pageOffset:]
}

func (region *memoryRegion) Size() int {
	return region.size
}

func (region *memoryRegion) Flush(async bool) error {
	return windows.FlushViewOfFile(uintptr(allocator.ByteSliceData(region.data)), uintptr(len(region.data)))
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
		err = fmt.Errorf("invalid mem region flags")
	}
	return
}

func mmapOffsetMultiple() int64 {
	g, p := getAllocGranularity(), os.Getpagesize()
	if g >= p {
		return int64(g)
	}
	return int64(p)
}
