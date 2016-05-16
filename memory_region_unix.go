// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"golang.org/x/sys/unix"
)

func init() {
	mmapOffsetMultiple = int64(os.Getpagesize())
}

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
	calculatedSize, err := fileSizeFromFd(obj)
	if err != nil {
		return nil, err
	}
	if calculatedSize > 0 && int64(size)+offset > calculatedSize {
		return nil, fmt.Errorf("invalid mapping length")
	}
	pageOffset := calcMmapOffsetFixup(offset)
	var data []byte
	if data, err = unix.Mmap(int(obj.Fd()), offset-pageOffset, size+int(pageOffset), prot, flags); err != nil {
		return nil, err
	}
	return &memoryRegion{data: data, size: size, pageOffset: pageOffset}, nil
}

func (region *memoryRegion) Close() error {
	if region.data != nil {
		err := unix.Munmap(region.data)
		region.data = nil
		region.pageOffset = 0
		region.size = 0
		return err
	}
	return nil
}

func (region *memoryRegion) Data() []byte {
	return region.data[region.pageOffset:]
}

func (region *memoryRegion) Flush(async bool) error {
	flag := unix.MS_SYNC
	if async {
		flag = unix.MS_ASYNC
	}
	return msync(region.data, flag)
}

func (region *memoryRegion) Size() int {
	return region.size
}

func memProtAndFlagsFromMode(mode int) (prot, flags int, err error) {
	switch mode {
	case MEM_READ_ONLY:
		prot = unix.PROT_READ
		flags = unix.MAP_SHARED
	case MEM_READ_PRIVATE:
		prot = unix.PROT_READ
		flags = unix.MAP_PRIVATE
	case MEM_READWRITE:
		prot = unix.PROT_READ | unix.PROT_WRITE
		flags = unix.MAP_SHARED
	case MEM_COPY_ON_WRITE:
		prot = unix.PROT_READ | unix.PROT_WRITE
		flags = unix.MAP_PRIVATE
	default:
		err = fmt.Errorf("invalid mem region flags")
	}
	return
}

// syscalls
func msync(data []byte, flags int) error {
	dataPointer := unsafe.Pointer(&data[0])
	_, _, err := unix.Syscall(unix.SYS_MSYNC, uintptr(dataPointer), uintptr(len(data)), uintptr(flags))
	allocator.Use(dataPointer)
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
