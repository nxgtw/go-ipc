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

func (mq *MessageQueue) Send(value interface{}, prio int) error {
	objSize := objectSize(reflect.ValueOf(value))
	data := make([]byte, objSize)
	if err := alloc(data, value); err != nil {
		return err
	}
	//defer use(unsafe.Pointer(&value))
	rawData := byteSliceAddress(data)
	_, _, err := syscall.Syscall6(unix.SYS_MQ_TIMEDSEND,
		uintptr(mq.Id()),
		uintptr(rawData),
		uintptr(objSize),
		uintptr(prio),
		uintptr(0),
		uintptr(0))
	if err != syscall.Errno(0) {
		print(err.Error())
		return err
	}
	return nil
}

func (mq *MessageQueue) Receive(typ int, flags int, object interface{}) error {
	value := reflect.ValueOf(object)
	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("the object must be a pointer")
	}
	elemValue := value.Elem()
	if err := checkType(elemValue.Type(), 0); err != nil {
		return err
	}
	objSize := objectSize(elemValue)
	addr := value.Pointer()
	print("to receive ", objSize, "\n")
	_, _, err := syscall.Syscall6(unix.SYS_MSGRCV,
		uintptr(mq.Id()),
		uintptr(objSize),
		uintptr(flags),
		uintptr(addr),
		uintptr(typ),
		uintptr(0))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
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
