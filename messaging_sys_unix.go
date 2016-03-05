// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	typeDataSize = int(unsafe.Sizeof(int(0)))
)

func msgget(k key, flags int) (int, error) {
	id, _, err := syscall.Syscall(unix.SYS_MSGGET, uintptr(k), uintptr(flags), 0)
	if err != syscall.Errno(0) {
		return 0, os.NewSyscallError("MSGGET", err)
	}
	return int(id), nil
}

func msgsnd(id int, typ int, data []byte, flags int) error {
	messageLen := typeDataSize + len(data)
	message := make([]byte, messageLen)
	*(*int)(unsafe.Pointer(&data[0])) = id
	copy(message[:typeDataSize], data)
	rawData := unsafe.Pointer(&message[0])
	_, _, err := syscall.Syscall6(unix.SYS_MSGSND,
		uintptr(id),
		uintptr(rawData),
		uintptr(len(data)),
		uintptr(flags),
		0,
		0)
	use(rawData)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("MSGSND", err)
	}
	return nil
}

func msgrcv(id int, data []byte, typ int, flags int) error {
	messageLen := typeDataSize + len(data)
	message := make([]byte, messageLen)
	rawData := unsafe.Pointer(&message[0])
	_, _, err := syscall.Syscall6(unix.SYS_MSGRCV,
		uintptr(id),
		uintptr(rawData),
		uintptr(len(data)),
		uintptr(typ),
		uintptr(flags),
		0)
	use(rawData)
	copy(data, message[typeDataSize:])
	if err != syscall.Errno(0) {
		return os.NewSyscallError("MSGRCV", err)
	}
	return nil
}
