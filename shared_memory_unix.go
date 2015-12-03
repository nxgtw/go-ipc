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
		err := os.Remove(path)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
}

// glibc/sysdeps/posix/shm_open.c
func shmOpen(path string, mode int, perm os.FileMode) (file *os.File, err error) {
	var unixMode int
	unixMode, err = modeToUnixMode(mode)
	if err != nil {
		return nil, err
	}
	switch {
	case mode&(O_OPEN_ONLY|O_CREATE_ONLY) != 0:
		file, err = os.OpenFile(path, unixMode, perm)
	case mode&O_OPEN_OR_CREATE != 0:
		amode, _ := accessModeToUnixMode(mode)
		for {
			if file, err = os.OpenFile(path, amode|unix.O_CREAT|unix.O_EXCL, perm); !os.IsExist(err) {
				break
			} else {
				if file, err = os.OpenFile(path, amode, perm); !os.IsNotExist(err) {
					break
				}
			}
		}
	}
	return
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

func createModeToUnixMode(mode int) (int, error) {
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

func accessModeToUnixMode(mode int) (umode int, err error) {
	if mode&O_NONBLOCK != 0 {
		umode |= unix.O_NONBLOCK
	}
	if mode&O_READ_ONLY != 0 {
		if mode&(O_WRITE_ONLY|O_READWRITE) != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return umode | os.O_RDONLY, nil
	}
	if mode&O_WRITE_ONLY != 0 {
		if mode&O_READWRITE != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return umode | os.O_WRONLY, nil
	}
	if mode&O_READWRITE != 0 {
		return umode | os.O_RDWR, nil
	}
	return 0, fmt.Errorf("no access mode flags")
}

func modeToUnixMode(mode int) (int, error) {
	if createMode, err := createModeToUnixMode(mode); err == nil {
		if accessMode, err := accessModeToUnixMode(mode); err == nil {
			return createMode | accessMode, nil
		} else {
			return 0, err
		}
	} else {
		return 0, err
	}
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
