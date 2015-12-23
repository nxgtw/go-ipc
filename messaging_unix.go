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
	DefaultMqMaxSize        = 8
	DefaultMqMaxMessageSize = 8192
)

var (
	MqNotifySignal = syscall.SIGUSR1
)

type MessageQueue struct {
	id   int
	name string
}

type MqAttr struct {
	Flags   int /* Flags: 0 or O_NONBLOCK */
	Maxmsg  int /* Max. # of messages on queue */
	Msgsize int /* Max. message size (bytes) */
	Curmsgs int /* # of messages currently in queue */
}

func CreateMessageQueue(name string, exclusive bool, perm os.FileMode, maxQueueSize, maxMsgSize int) (*MessageQueue, error) {
	sysflags := unix.O_CREAT | unix.O_RDWR
	if exclusive {
		sysflags |= unix.O_EXCL
	}
	attrs := &MqAttr{Maxmsg: maxQueueSize, Msgsize: maxMsgSize}
	if id, err := mq_open(name, sysflags, uint32(perm), attrs); err != nil {
		return nil, err
	} else {
		return &MessageQueue{id: int(id), name: name}, nil
	}
}

func OpenMessageQueue(name string, flags int) (*MessageQueue, error) {
	sysflags, err := mqFlagsToOsFlags(flags)
	if err != nil {
		return nil, err
	}
	if id, err := mq_open(name, sysflags, uint32(0), nil); err != nil {
		return nil, err
	} else {
		return &MessageQueue{id: int(id), name: name}, nil
	}
}

func (mq *MessageQueue) SendTimeout(object interface{}, prio int, timeout time.Duration) error {
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
	return mq_timedsend(mq.Id(), data, prio, timeoutToTimeSpec(timeout))
}

func (mq *MessageQueue) Send(object interface{}, prio int) error {
	return mq.SendTimeout(object, prio, time.Duration(-1))
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
	data := byteSliceFromUintptr(addr, objSize, objSize)
	return mq_timedreceive(mq.Id(), data, prio, timeoutToTimeSpec(timeout))
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

func (mq *MessageQueue) GetAttrs() (*MqAttr, error) {
	attrs := new(MqAttr)
	if err := mq_getsetattr(mq.Id(), nil, attrs); err != nil {
		return nil, err
	}
	return attrs, nil
}

func (mq *MessageQueue) SetNonBlock(nonBlock bool) error {
	attrs := new(MqAttr)
	if nonBlock {
		attrs.Flags |= unix.O_NONBLOCK
	}
	return mq_getsetattr(mq.Id(), attrs, nil)
}

func (mq *MessageQueue) Destroy() error {
	mq.Close()
	return DestroyMessageQueue(mq.name)
}

func (mq *MessageQueue) Notify() error {
	return nil
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
	if flags&(O_OPEN_OR_CREATE|O_CREATE_ONLY) != 0 {
		return 0, fmt.Errorf("to create message queue, use CreateMessageQueue func")
	}
	return sysflags, nil
}

func timeoutToTimeSpec(timeout time.Duration) *unix.Timespec {
	var ts *unix.Timespec
	if int64(timeout) >= 0 {
		sec, nsec := splitUnixTime(time.Now().Add(timeout).UnixNano())
		ts = &unix.Timespec{Sec: sec, Nsec: nsec}
	}
	return ts
}

// syscalls

type sigval struct { /* Data passed with notification */
	sigval_int uintptr /* A pointer-sized value to match the union size in syscall */
}

type sigevent struct {
	sigev_value             sigval
	sigev_notify            int
	sigev_signo             int
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
