// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd

package shm

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
)

func doDestroyMemoryObject(path string) error {
	err := shm_unlink(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
	}
	return err
}

func shmName(name string) (string, error) {
	const maxNameLen = 30
	// workaround from http://www.opensource.apple.com/source/Libc/Libc-320/sys/shm_open.c
	if runtime.GOOS == "darwin" {
		newName := fmt.Sprintf("%s\t%d", name, syscall.Geteuid())
		if len(newName) <= maxNameLen {
			name = newName
		}
	}
	return "/" + name, nil
}

func shmOpen(path string, flag int, perm os.FileMode) (*os.File, error) {
	flag |= syscall.O_CLOEXEC
	fd, err := shm_open(path, flag, int(perm))
	if err != nil {
		return nil, err
	}
	return os.NewFile(fd, path), nil
}

// syscalls

func shm_open(name string, flags, mode int) (uintptr, error) {
	nameBytes, err := syscall.BytePtrFromString(name)
	if err != nil {
		return 0, err
	}
	bytes := unsafe.Pointer(nameBytes)
	fd, _, err := syscall.Syscall(syscall.SYS_SHM_OPEN, uintptr(bytes), uintptr(flags), uintptr(mode))
	allocator.Use(bytes)
	if err != syscall.Errno(0) {
		return 0, err
	}
	return fd, nil
}

func shm_unlink(name string) error {
	nameBytes, err := syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}
	bytes := unsafe.Pointer(nameBytes)
	_, _, err = syscall.Syscall(syscall.SYS_SHM_UNLINK, uintptr(bytes), uintptr(0), uintptr(0))
	allocator.Use(bytes)
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
