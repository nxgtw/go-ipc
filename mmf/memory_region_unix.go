// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux

package mmf

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"github.com/pkg/errors"
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

func newMemoryRegion(obj Mappable, flag int, offset int64, size int) (*memoryRegion, error) {
	prot, flags, err := memProtAndFlagsFromMode(flag)
	if err != nil {
		return nil, errors.Wrap(err, "memory region flags check failed")
	}
	if size, err = checkMmapSize(obj, size); err != nil {
		return nil, errors.Wrap(err, "size check failed")
	}
	calculatedSize, err := fileSizeFromFd(obj)
	if err != nil {
		return nil, errors.Wrap(err, "file size check failed")
	}
	// we need this check on unix, because you can actually mmap more bytes,
	// then the size of the object, which can cause unexpected problems.
	if calculatedSize > 0 && int64(size)+offset > calculatedSize {
		return nil, errors.New("invalid mapping length")
	}
	pageOffset := calcMmapOffsetFixup(offset)
	var data []byte
	if data, err = unix.Mmap(int(obj.Fd()), offset-pageOffset, size+int(pageOffset), prot, flags); err != nil {
		return nil, errors.Wrap(err, "mmap failed")
	}
	return &memoryRegion{data: data, size: size, pageOffset: pageOffset}, nil
}

func (region *memoryRegion) Close() error {
	if region.data != nil {
		err := unix.Munmap(region.data)
		region.data = nil
		region.pageOffset = 0
		region.size = 0
		return errors.Wrap(err, "munmap failed")
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
	if err := msync(region.data, flag); err != nil {
		return errors.Wrap(err, "mync failed")
	}
	return nil
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
		err = errors.Errorf("invalid memory region flags %d", mode)
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
