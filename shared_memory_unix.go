// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type memoryObjectImpl struct {
	file *os.File
}

type memoryRegionImpl struct {
	data       []byte
	size       int
	pageOffset int64
}

func newMemoryObjectImpl(name string, mode int, perm os.FileMode) (impl *memoryObjectImpl, err error) {
	var path string
	if path, err = shmName(name); err != nil {
		return nil, err
	}
	var file *os.File
	file, err = shmOpen(path, mode, perm)
	if err != nil {
		return
	}
	impl = &memoryObjectImpl{file: file}
	return
}

func (impl *memoryObjectImpl) Destroy() error {
	if err := impl.Close(); err == nil {
		return destroyMemoryObject(impl.file.Name())
	} else {
		return err
	}
}

// returns the name of the object as it was given to NewMemoryObject()
func (impl *memoryObjectImpl) Name() string {
	return filepath.Base(impl.file.Name())
}

func (impl *memoryObjectImpl) Close() error {
	return impl.file.Close()
}

func (impl *memoryObjectImpl) Truncate(size int64) error {
	return impl.file.Truncate(size)
}

func (impl *memoryObjectImpl) Size() int64 {
	if fileInfo, err := impl.file.Stat(); err != nil {
		return 0
	} else {
		return fileInfo.Size()
	}
}

func (impl *memoryObjectImpl) Fd() int {
	return int(impl.file.Fd())
}

func DestroyMemoryObject(name string) error {
	if path, err := shmName(name); err != nil {
		return err
	} else {
		return destroyMemoryObject(path)
	}
}

func newMemoryRegionImpl(obj MappableHandle, mode int, offset int64, size int) (*memoryRegionImpl, error) {
	prot, flags, err := shmProtAndFlagsFromMode(mode)
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

func shmProtAndFlagsFromMode(mode int) (prot, flags int, err error) {
	switch mode {
	case SHM_READ_ONLY:
		prot = unix.PROT_READ
		flags = unix.MAP_SHARED
	case SHM_READ_PRIVATE:
		prot = unix.PROT_READ
		flags = unix.MAP_PRIVATE
	case SHM_READWRITE:
		prot = unix.PROT_READ | unix.PROT_WRITE
		flags = unix.MAP_SHARED
	case SHM_COPY_ON_WRITE:
		prot = unix.PROT_READ | unix.PROT_WRITE
		flags = unix.MAP_PRIVATE
	default:
		err = fmt.Errorf("invalid shm region flags")
	}
	return
}

func shmCreateModeToOsMode(mode int) (int, error) {
	if mode&O_OPEN_OR_CREATE != 0 {
		if mode&(O_CREATE_ONLY|O_OPEN_ONLY) != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return os.O_CREATE | os.O_TRUNC | os.O_RDWR, nil
	}
	if mode&O_CREATE_ONLY != 0 {
		if mode&O_OPEN_ONLY != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return os.O_CREATE | os.O_EXCL | os.O_RDWR, nil
	}
	if mode&O_OPEN_ONLY != 0 {
		return 0, nil
	}
	return 0, fmt.Errorf("no create mode flags")
}

func shmModeToOsMode(mode int) (int, error) {
	if createMode, err := shmCreateModeToOsMode(mode); err == nil {
		if accessMode, err := accessModeToOsMode(mode); err == nil {
			return createMode | accessMode, nil
		} else {
			return 0, err
		}
	} else {
		return 0, err
	}
}

// syscalls
func msync(data []byte, flags int) error {
	_, _, err := unix.Syscall(unix.SYS_MSYNC, uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), uintptr(flags))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
