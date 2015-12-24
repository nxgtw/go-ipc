// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	cSIGEV_SIGNAL      = 0
	cSIGEV_NONE        = 1
	cSIGEV_THREAD      = 2
	cNOTIFY_COOKIE_LEN = 32
)

func initMqNotifications(ch chan<- int) (int, error) {
	notifySocketFd, err := syscall.Socket(syscall.AF_NETLINK,
		syscall.SOCK_RAW|syscall.SOCK_CLOEXEC,
		syscall.NETLINK_ROUTE)
	if err != nil {
		return -1, err
	}
	go func() {
		var data [cNOTIFY_COOKIE_LEN]byte
		for {
			n, _, err := syscall.Recvfrom(notifySocketFd, data[:], syscall.MSG_NOSIGNAL|syscall.MSG_WAITALL)
			if n == cNOTIFY_COOKIE_LEN && err == nil {
				ndata := (*notify_data)(unsafe.Pointer(byteSliceAddress(data[:])))
				ch <- ndata.mq_id
			} else {
				return
			}
		}
	}()
	return notifySocketFd, nil
}

// syscalls

type notify_data struct {
	mq_id   int
	padding [cNOTIFY_COOKIE_LEN - unsafe.Sizeof(int(0))]byte
}

type sigval struct { /* Data passed with notification */
	sigval_ptr uintptr /* A pointer-sized value to match the union size in syscall */
}

type sigevent struct {
	sigev_value             sigval
	sigev_signo             int32
	sigev_notify            int32
	sigev_notify_function   uintptr
	sigev_notify_attributes uintptr
	padding                 [8]int32 // 8 is the maximum padding size
}

func mq_open(name string, flags int, mode uint32, attrs *MqAttr) (int, error) {
	nameBytes, err := syscall.BytePtrFromString(name)
	if err != nil {
		return -1, err
	}
	bytes := unsafe.Pointer(nameBytes)
	attrsP := unsafe.Pointer(attrs)
	id, _, err := syscall.Syscall6(unix.SYS_MQ_OPEN,
		uintptr(bytes),
		uintptr(flags),
		uintptr(mode),
		uintptr(attrsP),
		0,
		0)
	use(bytes)
	use(attrsP)
	if err != syscall.Errno(0) {
		return -1, err
	}
	return int(id), nil
}

func mq_timedsend(id int, data []byte, prio int, timeout *unix.Timespec) error {
	rawData := byteSliceAddress(data)
	_, _, err := syscall.Syscall6(unix.SYS_MQ_TIMEDSEND,
		uintptr(id),
		uintptr(rawData),
		uintptr(len(data)),
		uintptr(prio),
		uintptr(unsafe.Pointer(timeout)),
		uintptr(0))
	use(unsafe.Pointer(rawData))
	use(unsafe.Pointer(timeout))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}

func mq_timedreceive(id int, data []byte, prio *int, timeout *unix.Timespec) error {
	rawData := byteSliceAddress(data)
	_, _, err := syscall.Syscall6(unix.SYS_MQ_TIMEDRECEIVE,
		uintptr(id),
		uintptr(rawData),
		uintptr(len(data)),
		uintptr(unsafe.Pointer(prio)),
		uintptr(unsafe.Pointer(timeout)),
		uintptr(0))
	use(unsafe.Pointer(rawData))
	use(unsafe.Pointer(timeout))
	use(unsafe.Pointer(prio))
	use(unsafe.Pointer(timeout))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}

func mq_notify(id int, event *sigevent) error {
	_, _, err := syscall.Syscall(unix.SYS_MQ_NOTIFY, uintptr(id), uintptr(unsafe.Pointer(event)), uintptr(0))
	use(unsafe.Pointer(event))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}

func mq_getsetattr(id int, attrs, oldAttrs *MqAttr) error {
	_, _, err := syscall.Syscall(unix.SYS_MQ_GETSETATTR,
		uintptr(id),
		uintptr(unsafe.Pointer(attrs)),
		uintptr(unsafe.Pointer(oldAttrs)))
	use(unsafe.Pointer(attrs))
	use(unsafe.Pointer(oldAttrs))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}

func mq_unlink(name string) error {
	nameBytes, err := syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}
	bytes := unsafe.Pointer(nameBytes)
	_, _, err = syscall.Syscall(unix.SYS_MQ_UNLINK, uintptr(bytes), uintptr(0), uintptr(0))
	use(bytes)
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
