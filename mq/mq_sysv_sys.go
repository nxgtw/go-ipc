// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux,amd64 darwin freebsd

package mq

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/unix"
)

var (
	sysMsgGet uintptr
	sysMsgSnd uintptr
	sysMsgRcv uintptr
	sysMsgCtl uintptr
)

func msgget(k common.Key, flags int) (int, error) {
	id, _, err := unix.Syscall(sysMsgGet, uintptr(k), uintptr(flags), 0)
	if err != syscall.Errno(0) {
		if err == unix.EEXIST || err == unix.ENOENT {
			return 0, &os.PathError{Op: "MSGGET", Path: "", Err: err}
		}
		return 0, os.NewSyscallError("MSGGET", err)
	}
	return int(id), nil
}

func msgsnd(id int, typ int, data []byte, flags int) error {
	messageLen := typeDataSize + len(data)
	message := make([]byte, messageLen)
	rawData := allocator.ByteSliceData(message)
	*(*int)(rawData) = typ
	copy(message[typeDataSize:], data)
	_, _, err := unix.Syscall6(sysMsgSnd,
		uintptr(id),
		uintptr(rawData),
		uintptr(len(data)),
		uintptr(flags),
		0,
		0)
	allocator.Use(rawData)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("MSGSND", err)
	}
	return nil
}

func msgrcv(id int, data []byte, typ int, flags int) (int, error) {
	messageLen := typeDataSize + len(data)
	message := make([]byte, messageLen)
	rawData := allocator.ByteSliceData(message)
	len, _, err := unix.Syscall6(sysMsgRcv,
		uintptr(id),
		uintptr(rawData),
		uintptr(len(data)),
		uintptr(typ),
		uintptr(flags),
		0)
	allocator.Use(rawData)
	copy(data, message[typeDataSize:])
	if err != syscall.Errno(0) {
		return 0, os.NewSyscallError("MSGRCV", err)
	}
	return int(len), nil
}

func msgctl(id, cmd int, buf *msqidDs) error {
	pBuf := unsafe.Pointer(buf)
	_, _, err := unix.Syscall(sysMsgCtl, uintptr(id), uintptr(cmd), uintptr(pBuf))
	allocator.Use(pBuf)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("MSGCTL", err)
	}
	return nil
}
