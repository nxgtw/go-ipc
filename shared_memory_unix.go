// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"errors"
	"os"
	"strings"
	"sync"

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

type memoryRegionImpl struct {
	file *os.File
	data []byte
}

func newMemoryRegionImpl(name string, size int64, mode int, flags uint32) (impl *memoryRegionImpl, err error) {
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
	file, err := shmOpen(path, mode)
	if err != nil {
		return nil, err
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
	if data, mmapErr := unix.Mmap(int(file.Fd()), 0, int(size), unix.PROT_NONE, unix.MAP_SHARED); err != nil {
		err = mmapErr
		return
	} else {
		impl = &memoryRegionImpl{file: file, data: data}
	}
	return
}

func (impl *memoryRegionImpl) Destroy() error {
	if err := impl.Close(); err == nil {
		return os.Remove(impl.file.Name())
	} else {
		return err
	}
}

func (impl *memoryRegionImpl) Close() error {
	if err := unix.Munmap(impl.data); err != nil {
		return err
	}
	return impl.file.Close()
}

func (impl *memoryRegionImpl) Truncate(size int64) error {
	return impl.file.Truncate(size)
}

func (impl *memoryRegionImpl) Size() int64 {
	if fileInfo, err := impl.file.Stat(); err != nil {
		return 0
	} else {
		return fileInfo.Size()
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
