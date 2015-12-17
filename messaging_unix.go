// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"fmt"
	"os"
	"reflect"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type MessageQueue struct {
	id   int
	name string
}

func NewMessageQueue(name string, flags int, perm os.FileMode) (*MessageQueue, error) {
	sysflags, err := mqFlagsToOsFlags(flags)
	if err != nil {
		return nil, err
	}
	id, err := mq_open(name, sysflags, uint32(perm))
	if err != nil {
		return nil, err
	}
	return &MessageQueue{id: int(id)}, nil
}

func (mq *MessageQueue) Send(object interface{}, prio int) error {
	value := reflect.ValueOf(object)
	if err := checkType(value.Type(), 0); err != nil {
		return err
	}
	objSize := objectSize(value)
	data := make([]byte, objSize)
	print(objSize)
	if err := alloc(data, value); err != nil {
		return err
	}
	return mq_timedsend(mq.Id(), data, objSize, prio, nil)
}

func (mq *MessageQueue) Receive(object interface{}, prio *int) error {
	value := reflect.ValueOf(object)
	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("the object must be a pointer")
	}
	valueElem := value.Elem()
	if err := checkType(valueElem.Type(), 0); err != nil {
		return err
	}
	objSize := objectSize(valueElem)
	addr := value.Pointer()
	data := byteSliceFromUntptr(addr, objSize, objSize)
	return mq_timedreceive(mq.Id(), data, objSize, prio, nil)
}

func (mq *MessageQueue) Id() int {
	return mq.id
}

func (mq *MessageQueue) Destroy() error {
	return DestroyMessageQueue(mq.name)
}

func mqFlagsToOsFlags(flags int) (int, error) {
	sysflags, err := accessModeToOsMode(flags)
	if err != nil {
		return 0, err
	}
	sysflags |= unix.O_CLOEXEC
	if flags&O_NONBLOCK != 0 {
		sysflags |= unix.O_NONBLOCK
	}
	if flags&O_OPEN_OR_CREATE != 0 {
		sysflags |= unix.O_CREAT
	}
	if flags&O_CREATE_ONLY != 0 {
		sysflags |= (unix.O_CREAT | unix.O_EXCL)
	}
	return sysflags, nil
}

func DestroyMessageQueue(name string) error {
	return mq_unlink(name)
}

// syscalls
func mq_open(name string, flags int, mode uint32) (int, error) {
	nameBytes, err := syscall.BytePtrFromString(name)
	if err != nil {
		return -1, err
	}
	bytes := unsafe.Pointer(nameBytes)
	defer use(bytes)
	id, _, err := syscall.Syscall(unix.SYS_MQ_OPEN, uintptr(bytes), uintptr(flags), uintptr(mode))
	if err != syscall.Errno(0) {
		return -1, err
	}
	return int(id), nil
}

func mq_timedsend(id int, data []byte, lenght int, prio int, timeout *unix.Timespec) error {
	rawData := byteSliceAddress(data)
	defer use(unsafe.Pointer(rawData))
	defer use(unsafe.Pointer(timeout))
	_, _, err := syscall.Syscall6(unix.SYS_MQ_TIMEDSEND,
		uintptr(id),
		uintptr(rawData),
		uintptr(lenght),
		uintptr(prio),
		uintptr(unsafe.Pointer(timeout)),
		uintptr(0))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}

func mq_timedreceive(id int, data []byte, lenght int, prio *int, timeout *unix.Timespec) error {
	rawData := byteSliceAddress(data)
	defer use(unsafe.Pointer(rawData))
	defer use(unsafe.Pointer(timeout))
	defer use(unsafe.Pointer(prio))
	_, _, err := syscall.Syscall6(unix.SYS_MQ_TIMEDRECEIVE,
		uintptr(id),
		uintptr(rawData),
		uintptr(lenght),
		uintptr(unsafe.Pointer(prio)),
		uintptr(unsafe.Pointer(timeout)),
		uintptr(0))
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
	defer use(bytes)
	_, _, err = syscall.Syscall(unix.SYS_MQ_UNLINK, uintptr(bytes), uintptr(0), uintptr(0))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
