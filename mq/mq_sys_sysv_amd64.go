// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris
// +build amd64

package mq

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/unix"
)

func msgget(k common.Key, flags int) (int, error) {
	id, _, err := syscall.Syscall(unix.SYS_MSGGET, uintptr(k), uintptr(flags), 0)
	if err != syscall.Errno(0) {
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
	_, _, err := syscall.Syscall6(unix.SYS_MSGSND,
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

func msgrcv(id int, data []byte, typ int, flags int) error {
	messageLen := typeDataSize + len(data)
	message := make([]byte, messageLen)
	rawData := allocator.ByteSliceData(message)
	_, _, err := syscall.Syscall6(unix.SYS_MSGRCV,
		uintptr(id),
		uintptr(rawData),
		uintptr(len(data)),
		uintptr(typ),
		uintptr(flags),
		0)
	allocator.Use(rawData)
	copy(data, message[typeDataSize:])
	if err != syscall.Errno(0) {
		return os.NewSyscallError("MSGRCV", err)
	}
	return nil
}

func msgctl(id, cmd int, buf *msqidDs) error {
	pBuf := unsafe.Pointer(buf)
	_, _, err := syscall.Syscall(unix.SYS_MSGCTL, uintptr(id), uintptr(cmd), uintptr(pBuf))
	allocator.Use(pBuf)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("MSGCTL", err)
	}
	return nil
}
