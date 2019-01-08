// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/nxgtw/go-ipc/internal/allocator"
	"github.com/nxgtw/go-ipc/internal/common"

	"golang.org/x/sys/unix"
)

const (
	cSEMOP      = 1
	cSEMGET     = 2
	cSEMCTL     = 3
	cSEMTIMEDOP = 4
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
		if err == unix.EEXIST || err == unix.ENOENT {
			return 0, &os.PathError{Op: "SEMGET", Path: "", Err: err}
		}
		return 0, err
	}
	return int(id), nil
}

func semctl(id, num, cmd int) error {
	_, _, err := unix.Syscall6(unix.SYS_IPC, cSEMCTL, uintptr(id), uintptr(num), uintptr(cmd), uintptr(semun_inst), 0)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMCTL", err)
	}
	return nil
}

func semop(id int, ops []sembuf) error {
	return semtimedop(id, ops, nil)
}

func semtimedop(id int, ops []sembuf, timeout *unix.Timespec) error {
	if len(ops) == 0 {
		return nil
	}
	pOps := unsafe.Pointer(&ops[0])
	pTimeout := unsafe.Pointer(timeout)
	_, _, err := unix.Syscall6(unix.SYS_IPC, cSEMTIMEDOP, uintptr(id), uintptr(len(ops)), 0, uintptr(pOps), uintptr(pTimeout))
	allocator.Use(pOps)
	allocator.Use(pTimeout)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMTIMEDOP", err)
	}
	return nil
}
