// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"golang.org/x/sys/unix"
)

const (
	// DefaultLinuxMqMaxSize is the default queue size on linux
	DefaultLinuxMqMaxSize = 8
	// DefaultLinuxMqMaxMessageSize is the maximum message size on linux
	DefaultLinuxMqMaxMessageSize = 8192
)

// this is to ensure, that linux implementation of ipc mq
// satisfy the minimal queue interface
var (
	_ Messenger = (*LinuxMessageQueue)(nil)
)

// LinuxMessageQueue is a linux-specific ipc mechanism based on message passing
type LinuxMessageQueue struct {
	id             int
	name           string
	notifySocketFd int
	flags          int
	// The following field are needed if the size of the object to where we want
	// to receive a message is less, then the maximum message size of the queue.
	// In this case we use inputBuff to receive a message, and if the real size
	// of the message <= the imput object size, we copy our buffer into that object.
	inputBuff []byte
}

// MqAttr contains attributes of the queue
type MqAttr struct {
	Flags   int /* Flags: 0 or O_NONBLOCK */
	Maxmsg  int /* Max. # of messages on queue */
	Msgsize int /* Max. message size (bytes) */
	Curmsgs int /* # of messages currently in queue */
}

// CreateLinuxMessageQueue creates a new named message queue.
// 'x' permission cannot be used.
func CreateLinuxMessageQueue(name string, perm os.FileMode, maxQueueSize, maxMsgSize int) (*LinuxMessageQueue, error) {
	if !checkMqPerm(perm) {
		return nil, errors.New("invalid mq permissions")
	}
	sysflags := unix.O_CREAT | unix.O_RDWR | unix.O_EXCL
	attrs := &MqAttr{Maxmsg: maxQueueSize, Msgsize: maxMsgSize}
	id, err := mq_open(name, sysflags, uint32(perm), attrs)
	if err != nil {
		return nil, err
	}
	return &LinuxMessageQueue{
		id:             id,
		name:           name,
		notifySocketFd: -1,
		inputBuff:      make([]byte, maxMsgSize),
	}, nil
}

// OpenLinuxMessageQueue opens an existing message queue.
// Returns an error, if it does not exist.
func OpenLinuxMessageQueue(name string, flags int) (*LinuxMessageQueue, error) {
	sysflags, err := mqFlagsToOsFlags(flags)
	if err != nil {
		return nil, err
	}
	var id int
	if id, err = mq_open(name, sysflags, uint32(0), nil); err != nil {
		return nil, err
	}
	result := &LinuxMessageQueue{id: id, name: name, notifySocketFd: -1, flags: flags}
	if attrs, err := result.GetAttrs(); err != nil {
		result.Close()
		return nil, err
	} else {
		result.inputBuff = make([]byte, attrs.Msgsize)
	}
	return result, nil
}

// SendTimeoutPriority sends a message with a given priority.
// It blocks if the queue is full, waiting for a message unless timeout is passed.
func (mq *LinuxMessageQueue) SendTimeoutPriority(object interface{}, prio int, timeout time.Duration) error {
	data, err := allocator.ObjectData(object)
	if err != nil {
		return err
	}
	return mq_timedsend(mq.ID(), data, prio, timeoutToTimeSpec(timeout))
}

// SendPriority sends a message with a given priority.
// It blocks if the queue is full.
func (mq *LinuxMessageQueue) SendPriority(object interface{}, prio int) error {
	return mq.SendTimeoutPriority(object, prio, time.Duration(-1))
}

// SendTimeout sends a message with a default (0) priority.
// It blocks if the queue is full, waiting for a message unless timeout is passed.
func (mq *LinuxMessageQueue) SendTimeout(object interface{}, timeout time.Duration) error {
	return mq.SendTimeoutPriority(object, 0, timeout)
}

// Send sends a message with a default (0) priority.
// It blocks if the queue is full.
func (mq *LinuxMessageQueue) Send(object interface{}) error {
	timeout := time.Duration(-1)
	if mq.flags&O_NONBLOCK != 0 {
		timeout = time.Duration(0)
	}
	err := mq.SendTimeoutPriority(object, 0, timeout)
	if mq.flags&O_NONBLOCK != 0 && err != nil {
		if sysErr, ok := err.(syscall.Errno); ok {
			if sysErr.Temporary() {
				err = nil
			}
		}
	}
	return err
}

// ReceiveTimeoutPriority receives a message, returning its priority.
// It blocks if the queue is empty, waiting for a message unless timeout is passed.
func (mq *LinuxMessageQueue) ReceiveTimeoutPriority(object interface{}, timeout time.Duration) (int, error) {
	if !allocator.IsReferenceType(object) {
		return 0, fmt.Errorf("expected a slice, or a pointer")
	}
	objData, err := allocator.ObjectData(object)
	if err != nil {
		return 0, err
	}
	var data []byte
	curMaxMsgSize := len(mq.inputBuff)
	if len(objData) < curMaxMsgSize {
		data = mq.inputBuff
	} else {
		data = objData
	}
	var prio int
	msgSize, maxMsgSize, err := mq_timedreceive(mq.ID(), data, &prio, timeoutToTimeSpec(timeout))
	if maxMsgSize != 0 {
		if curMaxMsgSize != maxMsgSize {
			mq.inputBuff = make([]byte, maxMsgSize)
		}
	}
	if err != nil {
		return 0, err
	}
	if len(objData) < curMaxMsgSize {
		copy(objData, data[:msgSize])
	}
	//fmt.Println(data)
	return prio, nil
}

// ReceivePriority receives a message, returning its priority.
// It blocks if the queue is empty.
func (mq *LinuxMessageQueue) ReceivePriority(object interface{}) (int, error) {
	return mq.ReceiveTimeoutPriority(object, time.Duration(-1))
}

// ReceiveTimeout receives a message.
// It blocks if the queue is empty, waiting for a message unless timeout is passed.
func (mq *LinuxMessageQueue) ReceiveTimeout(object interface{}, timeout time.Duration) error {
	_, err := mq.ReceiveTimeoutPriority(object, timeout)
	return err
}

// Receive receives a message.
// It blocks if the queue is empty.
func (mq *LinuxMessageQueue) Receive(object interface{}) error {
	timeout := time.Duration(-1)
	if mq.flags&O_NONBLOCK != 0 {
		timeout = time.Duration(0)
	}
	_, err := mq.ReceiveTimeoutPriority(object, timeout)
	return err
}

// ID return unique id of the queue.
func (mq *LinuxMessageQueue) ID() int {
	return mq.id
}

// Close closes the queue.
func (mq *LinuxMessageQueue) Close() error {
	if mq.notifySocketFd != -1 {
		mq.NotifyCancel()
	}
	err := unix.Close(mq.ID())
	*mq = LinuxMessageQueue{notifySocketFd: -1}
	return err
}

// GetAttrs returns attributes of the queue
func (mq *LinuxMessageQueue) GetAttrs() (*MqAttr, error) {
	attrs := new(MqAttr)
	if err := mq_getsetattr(mq.ID(), nil, attrs); err != nil {
		return nil, err
	}
	return attrs, nil
}

// SetBlocking sets whether the send/receive operations on the queue block.
// This appliesa to the current instance only.
func (mq *LinuxMessageQueue) SetBlocking(block bool) error {
	if block {
		mq.flags &= ^O_NONBLOCK
	} else {
		mq.flags |= O_NONBLOCK
	}
	return nil
}

// Destroy closes the queue and removes it permanently
func (mq *LinuxMessageQueue) Destroy() error {
	name := mq.name
	mq.Close()
	return DestroyLinuxMessageQueue(name)
}

// Notify notifies about new messages in the queue by sending id of the queue to the channel.
// If there are messages in the queue, no notification will be sent
// unless all of them are read.
func (mq *LinuxMessageQueue) Notify(ch chan<- int) error {
	if ch == nil {
		return fmt.Errorf("cannot notify on a nil-chan")
	}
	notifySocketFd, err := initMqNotifications(ch)
	if err != nil {
		return fmt.Errorf("unable to init notifications subsystem")
	}
	ndata := &notify_data{mq_id: mq.ID()}
	pndata := unsafe.Pointer(ndata)
	defer use(pndata)
	ev := &sigevent{
		sigev_notify: cSIGEV_THREAD,
		sigev_signo:  int32(notifySocketFd),
		sigev_value:  sigval{sigval_ptr: uintptr(pndata)},
	}
	if err = mq_notify(mq.ID(), ev); err != nil {
		syscall.Close(notifySocketFd)
	} else {
		mq.notifySocketFd = notifySocketFd
	}
	return err
}

// NotifyCancel cancels notification subscribtion
func (mq *LinuxMessageQueue) NotifyCancel() error {
	var err error
	if err := mq_notify(mq.ID(), nil); err == nil {
		syscall.Close(mq.notifySocketFd)
		mq.notifySocketFd = -1
		return nil
	}
	return err
}

// DestroyLinuxMessageQueue removes the queue permanently
func DestroyLinuxMessageQueue(name string) error {
	err := mq_unlink(name)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
	}
	return err
}

// SetLinuxMqBlocking sets whether the operations on a linux mq block.
// This will apply for all send/receive operations on any instance of the
// linux mq with the given name.
func SetLinuxMqBlocking(name string, block bool) error {
	mq, err := OpenLinuxMessageQueue(name, O_READWRITE)
	if err != nil {
		return err
	}
	attrs := new(MqAttr)
	if !block {
		attrs.Flags |= unix.O_NONBLOCK
	}
	return mq_getsetattr(mq.ID(), attrs, nil)
}

func mqFlagsToOsFlags(flags int) (int, error) {
	// by default, assume we're opening it for readwrite
	if flags&(O_READWRITE|O_READ_ONLY|O_WRITE_ONLY) == 0 {
		flags = O_READWRITE
	}
	sysflags, err := accessModeToOsMode(flags)
	if err != nil {
		return 0, err
	}
	sysflags |= unix.O_CLOEXEC
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
