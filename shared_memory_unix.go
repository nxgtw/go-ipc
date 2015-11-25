// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"errors"
	"fmt"
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
	fd   int
	path string
}

func newMemoryRegionImpl(name string, size int64, mode int, flags uint32) (*memoryRegionImpl, error) {
	var err error
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
	fd, err := shmOpen(path, mode)
	if err != nil {
		print("asdasd")
		return nil, err
	}
	if err = unix.Ftruncate(fd, size); err != nil {
		return nil, err
	}
	return &memoryRegionImpl{fd: fd, path: path}, nil
}

func (impl *memoryRegionImpl) Destroy() error {
	if err := impl.Close(); err == nil {
		return unix.Unlink(impl.path)
	} else {
		return err
	}
}

func (impl *memoryRegionImpl) Close() error {
	if impl.fd != -1 {
		if err := unix.Close(impl.fd); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("shared memory object was not opened")
	}
	impl.fd = -1
	return nil
}

func (impl *memoryRegionImpl) Truncate(size int64) error {
	return unix.Ftruncate(impl.fd, size)
}

// glibc/sysdeps/posix/shm_open.c
func shmOpen(path string, mode int) (fd int, err error) {
	if fd, err = unix.Open(path, mode, 0666); err != nil { // TODO (avd) - 0777?
		return
	}
	unix.CloseOnExec(fd)
	return
}

// glibc/sysdeps/posix/shm-directory.h
func shmName(name string) (string, error) {
	name = strings.TrimLeft(name, "/")
	nameLen := len(name)
	if nameLen == 1 || nameLen >= maxNameLen || strings.Contains(name, "/") {
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
		umode |= (unix.O_CREAT | unix.O_RDWR | unix.O_TRUNC)
	}
	if mode&SHM_OPEN_CREATE_IF_NOT_EXISTS != 0 {
		umode |= (unix.O_EXCL | unix.O_CREAT)
	}
	if mode&SHM_OPEN_TRUNC != 0 {
		umode |= unix.O_TRUNC
	}
	if mode&SHM_OPEN_RDONLY != 0 {
		umode |= unix.O_RDONLY
	}
	if mode&SHM_OPEN_RDWR != 0 {
		umode |= unix.O_RDWR
	}
	return umode, nil
}
