// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris
// +build amd64

package sync

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/unix"
)

const (
	cSemUndo = 0x1000
)

type sembuf struct {
	semnum uint16
	semop  int16
	semflg int16
}

func semget(k common.Key, nsems, semflg int) (int, error) {
	id, _, err := unix.Syscall(unix.SYS_SEMGET, uintptr(k), uintptr(nsems), uintptr(semflg))
	if err != syscall.Errno(0) {
		return 0, err
	}
	return int(id), nil
}

func semctl(id, num, cmd int) error {
	_, _, err := syscall.Syscall6(unix.SYS_SEMCTL, uintptr(id), uintptr(id), uintptr(num), uintptr(cmd), 0, 0)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMCTL", err)
	}
	return nil
}

func semtimedop(id int, ops []sembuf, timeout *unix.Timespec) error {
	if len(ops) == 0 {
		return nil
	}
	pOps := unsafe.Pointer(&ops[0])
	pTimeout := unsafe.Pointer(timeout)
	_, _, err := syscall.Syscall6(unix.SYS_SEMTIMEDOP, uintptr(id), uintptr(pOps), uintptr(len(ops)), uintptr(pTimeout), 0, 0)
	allocator.Use(pOps)
	allocator.Use(pTimeout)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMTIMEDOP", err)
	}
	return nil
}
