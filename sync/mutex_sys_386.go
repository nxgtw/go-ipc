// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris
// +build 386

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
	cSEMOP  = 1
	cSEMGET = 2
	cSEMCTL = 3
)

// semun is a union used in semctl syscall. is not not actually used, so its size
// matches the size in kernel. we need one only global readonly pointer to it.
type semun struct {
	unused uintptr
}

var (
	semun_inst = unsafe.Pointer(&semun{})
)

func semget(k common.Key, nsems, semflg int) (int, error) {
	id, _, err := unix.Syscall6(unix.SYS_IPC, cSEMGET, uintptr(k), uintptr(nsems), uintptr(semflg), 0, 0)
	if err != syscall.Errno(0) {
		return 0, err
	}
	return int(id), nil
}

func semctl(id, num, cmd int) error {
	_, _, err := syscall.Syscall6(unix.SYS_IPC, cSEMCTL, uintptr(id), uintptr(num), uintptr(cmd), uintptr(semun_inst), 0)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMCTL", err)
	}
	return nil
}

func semop(id int, ops []sembuf) error {
	if len(ops) == 0 {
		return nil
	}
	pOps := unsafe.Pointer(&ops[0])
	_, _, err := syscall.Syscall6(unix.SYS_IPC, cSEMOP, uintptr(id), uintptr(len(ops)), 0, uintptr(pOps), 0)
	allocator.Use(pOps)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMOP", err)
	}
	return nil
}
