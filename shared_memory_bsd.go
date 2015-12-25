// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd netbsd openbsd

package ipc

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

func destroyMemoryObject(path string) error {
	return shm_unlink(path)
}

func shmName(name string) (string, error) {
	// workaround from http://www.opensource.apple.com/source/Libc/Libc-320/sys/shm_open.c
	if runtime.GOOS == "darwin" {
		name = fmt.Sprintf("%s\t%d", name, syscall.Geteuid())
	}
	return name, nil
}

func shmOpen(path string, mode int, perm os.FileMode) (*os.File, error) {
	osMode, err := shmModeToOsMode(mode)
	if err != nil {
		return nil, err
	}
	if fd, err := shm_open(path, osMode, int(perm)); err != nil {
		return nil, err
	} else {
		return os.NewFile(fd, path), nil
	}
}

// syscalls

func shm_open(name string, flags, mode int) (uintptr, error) {
	nameBytes, err := syscall.BytePtrFromString(name)
	if err != nil {
		return 0, err
	}
	bytes := unsafe.Pointer(nameBytes)
	fd, _, err := syscall.Syscall(syscall.SYS_SHM_OPEN, uintptr(bytes), uintptr(flags), uintptr(mode))
	use(bytes)
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
	use(bytes)
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
