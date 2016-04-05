// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux,amd64 darwin

package sync

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/unix"
)

func semget(k common.Key, nsems, semflg int) (int, error) {
	id, _, err := unix.Syscall(unix.SYS_SEMGET, uintptr(k), uintptr(nsems), uintptr(semflg))
	if err != syscall.Errno(0) {
		return 0, err
	}
	return int(id), nil
}

func semctl(id, num, cmd int) error {
	_, _, err := syscall.Syscall(unix.SYS_SEMCTL, uintptr(id), uintptr(num), uintptr(cmd))
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
	_, _, err := syscall.Syscall(unix.SYS_SEMOP, uintptr(id), uintptr(pOps), uintptr(len(ops)))
	allocator.Use(pOps)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMOP", err)
	}
	return nil
}

/*
220     AUE_SEMCTL      COMPAT7|NOSTD { int __semctl(int semid, int semnum, \
                                     int cmd, union semun_old *arg); }
221     AUE_SEMGET      NOSTD   { int semget(key_t key, int nsems, \
                                     int semflg); }
 222     AUE_SEMOP       NOSTD   { int semop(int semid, struct sembuf *sops, \
                                      size_t nsops); }
*
