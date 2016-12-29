// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd

package shm

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"golang.org/x/sys/unix"
)

func doDestroyMemoryObject(path string) error {
	err := shm_unlink(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
	}
	return err
}

func shmName(name string) (string, error) {
	const maxNameLen = 30
	// workaround from http://www.opensource.apple.com/source/Libc/Libc-320/sys/shm_open.c
	if isDarwin {
		newName := fmt.Sprintf("%s\t%d", name, unix.Geteuid())
		if len(newName) < maxNameLen {
			name = newName
		}
	}
	return "/" + name, nil
}

func shmOpen(path string, flag int, perm os.FileMode) (*os.File, error) {
	flag |= unix.O_CLOEXEC
	fd, err := shm_open(path, flag, int(perm))
	if err != nil {
		return nil, err
	}
	return os.NewFile(fd, path), nil
}

// syscalls

func shm_open(name string, flags, mode int) (uintptr, error) {
	nameBytes, err := unix.BytePtrFromString(name)
	if err != nil {
		return 0, err
	}
	bytes := unsafe.Pointer(nameBytes)
	fd, _, err := unix.Syscall(unix.SYS_SHM_OPEN, uintptr(bytes), uintptr(flags), uintptr(mode))
	allocator.Use(bytes)
	if err != syscall.Errno(0) {
		if err == unix.ENOENT || err == unix.EEXIST {
			return 0, &os.PathError{Path: name, Op: "shm_open", Err: err}
		}
		return 0, err
	}
	return fd, nil
}

func shm_unlink(name string) error {
	nameBytes, err := unix.BytePtrFromString(name)
	if err != nil {
		return err
	}
	bytes := unsafe.Pointer(nameBytes)
	_, _, err = unix.Syscall(unix.SYS_SHM_UNLINK, uintptr(bytes), uintptr(0), uintptr(0))
	allocator.Use(bytes)
	if err != syscall.Errno(0) {
		if err == unix.ENOENT {
			return &os.PathError{Path: name, Op: "shm_unlink", Err: err}
		}
		return err
	}
	return nil
}
