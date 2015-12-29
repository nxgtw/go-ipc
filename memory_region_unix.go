// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type memoryRegionImpl struct {
	data       []byte
	size       int
	pageOffset int64
}

func newMemoryRegionImpl(obj MappableHandle, mode int, offset int64, size int) (*memoryRegionImpl, error) {
	prot, flags, err := memProtAndFlagsFromMode(mode)
	if err != nil {
		return nil, err
	}
	pageOffset := calcValidOffset(offset)
	if data, err := unix.Mmap(obj.Fd(), offset-pageOffset, size+int(pageOffset), prot, flags); err != nil {
		return nil, err
	} else {
		return &memoryRegionImpl{data: data, size: size, pageOffset: pageOffset}, nil
	}
}

func (impl *memoryRegionImpl) Close() error {
	if impl.data != nil {
		err := unix.Munmap(impl.data)
		impl.data = nil
		impl.pageOffset = 0
		impl.size = 0
		return err
	}
	return nil
}

func (impl *memoryRegionImpl) Data() []byte {
	return impl.data[impl.pageOffset:]
}

func (impl *memoryRegionImpl) Flush(async bool) error {
	flag := unix.MS_SYNC
	if async {
		flag = unix.MS_ASYNC
	}
	return msync(impl.data, flag)
}

func (impl *memoryRegionImpl) Size() int {
	return impl.size
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
	use(dataPointer)
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
