// Copyright 2015 Aleksandr Demakin. All rights reserved.

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

type MessageQueue struct {
	id             int
	name           string
	notifySocketFd int
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
		return &MessageQueue{id: int(id), name: name, notifySocketFd: -1}, nil
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
		return &MessageQueue{id: int(id), name: name, notifySocketFd: -1}, nil
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
	if mq.notifySocketFd != -1 {
		mq.NotifyCancel()
	}
	return unix.Close(mq.Id())
}

func (mq *MessageQueue) GetAttrs() (*MqAttr, error) {
	attrs := new(MqAttr)
	if err := mq_getsetattr(mq.Id(), nil, attrs); err != nil {
		return nil, err
	}
	return attrs, nil
}

func (mq *MessageQueue) SetBlocking(block bool) error {
	attrs := new(MqAttr)
	if !block {
		attrs.Flags |= unix.O_NONBLOCK
	}
	return mq_getsetattr(mq.Id(), attrs, nil)
}

func (mq *MessageQueue) Destroy() error {
	mq.Close()
	return DestroyMessageQueue(mq.name)
}

// Notifies about new messages in the queue by
// sending id of the queue to the channel.
// If there are messages in the queue, no notification will be sent
// until all of them are read.
func (mq *MessageQueue) Notify(ch chan int) error {
	if ch == nil {
		return fmt.Errorf("cannot notify on a nil-chan")
	}
	notifySocketFd, err := initMqNotifications(ch)
	if err != nil {
		return fmt.Errorf("unable to init notifications subsystem")
	}
	ndata := &notify_data{mq_id: mq.Id()}
	pndata := unsafe.Pointer(ndata)
	defer use(pndata)
	ev := &sigevent{
		sigev_notify: cSIGEV_THREAD,
		sigev_signo:  int32(notifySocketFd),
		sigev_value:  sigval{sigval_ptr: uintptr(pndata)},
	}
	if err = mq_notify(mq.Id(), ev); err != nil {
		syscall.Close(notifySocketFd)
	} else {
		mq.notifySocketFd = notifySocketFd
	}
	return err
}

func (mq *MessageQueue) NotifyCancel() error {
	if err := mq_notify(mq.Id(), nil); err == nil {
		syscall.Close(mq.notifySocketFd)
		mq.notifySocketFd = -1
		return nil
	} else {
		return err
	}
}

func DestroyMessageQueue(name string) error {
	if err := mq_unlink(name); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
	}
	return nil
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
