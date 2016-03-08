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
	// DefaultMqMaxSize is the default queue size on linux
	DefaultMqMaxSize = 8
	// DefaultMqMaxMessageSize is the maximum queue size on linux
	DefaultMqMaxMessageSize = 8192
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
	var id int
	var err error
	if id, err = mq_open(name, sysflags, uint32(perm), attrs); err != nil {
		return nil, err
	}
	return &LinuxMessageQueue{id: id, name: name, notifySocketFd: -1}, nil
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
	return &LinuxMessageQueue{id: id, name: name, notifySocketFd: -1}, nil
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

// SendPriority sends a message with a default (0) priority.
// It blocks if the queue is full.
func (mq *LinuxMessageQueue) Send(object interface{}) error {
	return mq.SendTimeoutPriority(object, 0, time.Duration(-1))
}

// ReceiveTimeoutPriority receives a message, returning its priority.
// It blocks if the queue is empty, waiting for a message unless timeout is passed.
func (mq *LinuxMessageQueue) ReceiveTimeoutPriority(object interface{}, timeout time.Duration) (int, error) {
	if !allocator.IsReferenceType(object) {
		return 0, fmt.Errorf("expected a slice, or a pointer")
	}
	data, err := allocator.ObjectData(object)
	if err != nil {
		return 0, err
	}
	var prio int
	return prio, mq_timedreceive(mq.ID(), data, &prio, timeoutToTimeSpec(timeout))
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
	_, err := mq.ReceiveTimeoutPriority(object, time.Duration(-1))
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
	return unix.Close(mq.ID())
}

// GetAttrs returns attributes of the queue
func (mq *LinuxMessageQueue) GetAttrs() (*MqAttr, error) {
	attrs := new(MqAttr)
	if err := mq_getsetattr(mq.ID(), nil, attrs); err != nil {
		return nil, err
	}
	return attrs, nil
}

// SetBlocking sets whether the operations on the queue block.
func (mq *LinuxMessageQueue) SetBlocking(block bool) error {
	attrs := new(MqAttr)
	if !block {
		attrs.Flags |= unix.O_NONBLOCK
	}
	return mq_getsetattr(mq.ID(), attrs, nil)
}

// Destroy closes the queue and removes it permanently
func (mq *LinuxMessageQueue) Destroy() error {
	mq.Close()
	return DestroyMessageQueue(mq.name)
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

// DestroyMessageQueue removes the queue permanently
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
