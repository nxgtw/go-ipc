// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd netbsd openbsd

package shm

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"
)

func destroyMemoryObject(path string) error {
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

func shmOpen(path string, mode int, perm os.FileMode) (*os.File, error) {
	var f *os.File
	amode, err := common.AccessModeToOsMode(mode)
	if err != nil {
		return nil, err
	}
	opener := func(create bool) error {
		sysMode := amode | syscall.O_CLOEXEC
		if create {
			sysMode |= (syscall.O_CREAT | syscall.O_EXCL)
		}
		fd, err := shm_open(path, sysMode, int(perm))
		if err != nil {
			return err
		}
		f = os.NewFile(fd, path)
		return nil
	}
	_, err = common.OpenOrCreate(opener, mode&(ipc.O_CREATE_ONLY|ipc.O_OPEN_OR_CREATE|ipc.O_OPEN_ONLY))
	return f, err
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
