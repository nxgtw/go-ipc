// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"fmt"
	"os"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	DefaultMqMaxLen = 8
)

type MessageQueue struct {
	id   int
	name string
}

func NewMessageQueue(name string, flags int, perm os.FileMode, maxMsgSize int) (*MessageQueue, error) {
	sysflags, err := mqFlagsToOsFlags(flags)
	if err != nil {
		return nil, err
	}
	attrs := &mq_attr{mq_maxmsg: DefaultMqMaxLen, mq_msgsize: maxMsgSize}
	id, err := mq_open(name, sysflags, uint32(perm), attrs)
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
	// TODO(avd) - optimization: do not alloc a new object if we're sending a byte slice
	if err := alloc(data, object); err != nil {
		return err
	}
	return mq_timedsend(mq.Id(), data, prio, nil)
}

func (mq *MessageQueue) ReceiveTimeout(object interface{}, prio *int, timeout time.Duration) error {
	value := reflect.ValueOf(object)
	kind := value.Kind()
	var objSize int
	if kind == reflect.Ptr {
		valueElem := value.Elem()
		if err := checkType(valueElem.Type(), 0); err != nil {
			return err
		}
		objSize = objectSize(valueElem)
	} else if kind == reflect.Slice {
		if err := checkType(value.Type(), 0); err != nil {
			return err
		}
		objSize = objectSize(value)
	} else {
		return fmt.Errorf("the object must be a pointer or a slice")
	}
	addr := value.Pointer()
	data := byteSliceFromUntptr(addr, objSize, objSize)
	var ts *unix.Timespec
	if int64(timeout) >= 0 {
		sec, nsec := splitUnixTime(time.Now().Add(timeout).UnixNano())
		ts = &unix.Timespec{Sec: sec, Nsec: nsec}
	}
	return mq_timedreceive(mq.Id(), data, prio, ts)
}

func (mq *MessageQueue) Receive(object interface{}, prio *int) error {
	return mq.ReceiveTimeout(object, prio, time.Duration(-1))
}

func (mq *MessageQueue) Id() int {
	return mq.id
}

func (mq *MessageQueue) Close() error {
	return unix.Close(mq.Id())
}

func (mq *MessageQueue) Destroy() error {
	mq.Close()
	return DestroyMessageQueue(mq.name)
}

func DestroyMessageQueue(name string) error {
	return mq_unlink(name)
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

// syscalls

type mq_attr struct {
	mq_flags   int /* Flags: 0 or O_NONBLOCK */
	mq_maxmsg  int /* Max. # of messages on queue */
	mq_msgsize int /* Max. message size (bytes) */
	mq_curmsgs int /* # of messages currently in queue */
}

func mq_open(name string, flags int, mode uint32, attrs *mq_attr) (int, error) {
	nameBytes, err := syscall.BytePtrFromString(name)
	if err != nil {
		return -1, err
	}
	bytes := unsafe.Pointer(nameBytes)
	attrsP := unsafe.Pointer(attrs)
	defer use(bytes)
	defer use(attrsP)
	id, _, err := syscall.Syscall6(unix.SYS_MQ_OPEN,
		uintptr(bytes),
		uintptr(flags),
		uintptr(mode),
		uintptr(attrsP),
		0,
		0)
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
