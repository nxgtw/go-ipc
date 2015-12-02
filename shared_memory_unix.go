// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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
	data       []byte
	size       int
	pageOffset int64
}

func newMemoryObjectImpl(name string, mode int, perm os.FileMode) (impl *memoryObjectImpl, err error) {
	mode, err = modeToUnixMode(mode)
	if err != nil {
		return nil, err
	}
	var path string
	if path, err = shmName(name); err != nil {
		return nil, err
	}
	var file *os.File
	file, err = shmOpen(path, mode, perm)
	if err != nil {
		return
	} else {
		defer func() {
			if err != nil && file != nil {
				file.Close()
			}
		}()
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
	pageOffset := calcValidOffset(offset)
	if data, err := unix.Mmap(obj.Fd(), offset-pageOffset, size+int(pageOffset), prot, unix.MAP_SHARED); err != nil {
		return nil, err
	} else {
		return &memoryRegionImpl{data: data, size: size, pageOffset: pageOffset}, nil
	}
}

func (impl *memoryRegionImpl) Close() error {
	return unix.Munmap(impl.data)
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

func DestroyMemoryObject(name string) error {
	if path, err := shmName(name); err != nil {
		return err
	} else {
		return os.Remove(path)
	}
}

// glibc/sysdeps/posix/shm_open.c
func shmOpen(path string, mode int, perm os.FileMode) (file *os.File, err error) {
	return os.OpenFile(path, mode, perm)
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

func modeToUnixMode(mode int) (umode int, err error) {
	if mode&SHM_CREATE != 0 {
		umode |= (os.O_CREATE | os.O_TRUNC | os.O_RDWR)
		return
	}
	if mode&SHM_CREATE_ONLY != 0 {
		umode |= (os.O_CREATE | os.O_EXCL | os.O_RDWR)
		return
	}
	if mode&SHM_READ != 0 {
		if mode&SHM_RW == 0 {
			umode |= os.O_RDONLY
		} else {
			return 0, fmt.Errorf("both SHM_READ and SHM_RW flags are set")
		}
	} else if mode&SHM_RW != 0 {
		umode |= os.O_RDWR
	}
	return
}

func shmProtFromMode(mode int) int {
	prot := unix.PROT_NONE
	if mode&SHM_READ != 0 {
		prot |= unix.PROT_READ
	}
	if mode&SHM_RW != 0 {
		prot |= unix.PROT_WRITE
	}
	return prot
}

func msync(data []byte, flags int) error {
	_, _, err := unix.Syscall(unix.SYS_MSYNC, uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), uintptr(flags))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}

func calcValidOffset(offset int64) int64 {
	pageSize := int64(os.Getpagesize())
	return (offset - (offset/pageSize)*pageSize)
}
