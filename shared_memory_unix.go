// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	maxNameLen     = 255
	defaultShmPath = "/dev/shm/"
)

var (
	shmPathOnce sync.Once
	shmPath     string
)

type memoryObjectImpl struct {
	file *os.File
}

type memoryRegionImpl struct {
	data []byte
	size int
}

func newMemoryObjectImpl(name string, size int64, mode int, flags uint32) (impl *memoryObjectImpl, err error) {
	mode, err = modeToUnixMode(mode)
	if err != nil {
		return nil, err
	}
	flags, err = flagsToUnixFlags(flags)
	if err != nil {
		return nil, err
	}
	var path string
	if path, err = shmName(name); err != nil {
		return nil, err
	}
	var file *os.File
	file, err = shmOpen(path, mode)
	if err != nil {
		return
	} else {
		defer func() {
			if err != nil && file != nil {
				file.Close()
			}
		}()
	}
	if err = file.Truncate(size); err != nil {
		return
	}
	impl = &memoryObjectImpl{file: file}
	return
}

func (impl *memoryObjectImpl) Destroy() error {
	if err := impl.Close(); err == nil {
		return os.Remove(impl.file.Name())
	} else {
		return err
	}
}

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

func newMemoryRegionImpl(obj MappableHandle, mode int, offset int64, size int) (*memoryRegionImpl, error) {
	prot := shmProtFromMode(mode)
	if data, err := unix.Mmap(obj.Fd(), offset, size, prot, unix.MAP_SHARED); err != nil {
		return nil, err
	} else {
		return &memoryRegionImpl{data: data, size: size}, nil
	}
}

func (impl *memoryRegionImpl) Close() error {
	return unix.Munmap(impl.data)
}

func (impl *memoryRegionImpl) Data() []byte {
	return impl.data
}

func (impl *memoryRegionImpl) Flush() error {
	return msync(impl.data, unix.MS_SYNC)
}

func (impl *memoryRegionImpl) Size() int {
	return impl.size
}

func DestroyMemoryObject(name string) error {
	if path, err := shmName(name); err != nil {
		return err
	} else {
		return os.Remove(path)
	}
}

// glibc/sysdeps/posix/shm_open.c
func shmOpen(path string, mode int) (file *os.File, err error) {
	return os.OpenFile(path, mode, 0666) // TODO (avd) - 0777?
}

// glibc/sysdeps/posix/shm-directory.h
func shmName(name string) (string, error) {
	name = strings.TrimLeft(name, "/")
	nameLen := len(name)
	if nameLen == 0 || nameLen >= maxNameLen || strings.Contains(name, "/") {
		return "", errors.New("invalid shm name")
	}
	if dir, err := shmDirectory(); err != nil {
		return "", err
	} else {
		return dir + name, nil
	}
}

func shmDirectory() (string, error) {
	shmPathOnce.Do(locateShmFs)
	if len(shmPath) == 0 {
		return shmPath, errors.New("error locating the shared memory path")
	}
	return shmPath, nil
}

func flagsToUnixFlags(flags uint32) (uint32, error) {
	return flags, nil
}

func modeToUnixMode(mode int) (int, error) {
	var umode int
	if mode&SHM_OPEN_CREATE != 0 {
		umode |= (os.O_CREATE | os.O_RDWR | os.O_TRUNC)
	}
	if mode&SHM_OPEN_CREATE_IF_NOT_EXISTS != 0 {
		umode |= (os.O_EXCL | os.O_CREATE)
	}
	if mode&SHM_OPEN_READ != 0 {
		if mode&SHM_OPEN_WRITE != 0 {
			umode |= os.O_RDWR
		} else {
			umode |= os.O_RDONLY
		}
	} else if mode&SHM_OPEN_WRITE != 0 {
		umode |= os.O_WRONLY
	}
	return umode, nil
}

func shmProtFromMode(mode int) int {
	prot := unix.PROT_NONE
	if mode&SHM_OPEN_READ != 0 {
		prot |= unix.PROT_READ
	}
	if mode&SHM_OPEN_WRITE != 0 {
		prot |= unix.PROT_WRITE
	}
	return prot
}

func msync(data []byte, flags int) error {
	_, _, err := unix.Syscall(unix.SYS_MSYNC, uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), uintptr(flags))
	return err
}
