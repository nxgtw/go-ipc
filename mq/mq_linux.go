// Copyright 2015 Aleksandr Demakin. All rights reserved.

package mq

import (
	"os"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

const (
	// DefaultLinuxMqMaxSize is the default linux mq queue size.
	DefaultLinuxMqMaxSize = 8
	// DefaultLinuxMqMessageSize is the linux mq message size.
	// Its max value can be set via procfs.
	DefaultLinuxMqMessageSize = 8192
)

// this is to ensure, that linux implementation of ipc mq satisfies queue interfaces.
var (
	_ Messenger         = (*LinuxMessageQueue)(nil)
	_ TimedMessenger    = (*LinuxMessageQueue)(nil)
	_ PriorityMessenger = (*LinuxMessageQueue)(nil)
)

// LinuxMessageQueue is a linux-specific ipc mechanism based on message passing.
type LinuxMessageQueue struct {
	id           int
	name         string
	cancelSocket int
	flags        int
	// The following field is needed if the size of the input buffer
	// less, then the queue message size.
	// In this case we use inputBuff to receive a message, and if the real size
	// of the message <= the input buffer size, we copy our buffer into that object.
	inputBuff []byte
}

// linuxMqAttr contains attributes of the queue.
type linuxMqAttr struct {
	Flags   int /* Flags: 0 or O_NONBLOCK */
	Maxmsg  int /* Max. # of messages on queue */
	Msgsize int /* Max. message size (bytes) */
	Curmsgs int /* # of messages currently in queue */
}

// CreateLinuxMessageQueue creates new queue with the given name and permissions.
//	name - unique mq name.
//	flag - flag is a combination of os.O_EXCL and O_NONBLOCK.
//	perm - object's permission bits.
//	maxQueueSize - queue capacity.
//	maxMsgSize - maximum message size.
func CreateLinuxMessageQueue(name string, flag int, perm os.FileMode, maxQueueSize, maxMsgSize int) (*LinuxMessageQueue, error) {
	if !checkMqPerm(perm) {
		return nil, errors.New("invalid mq permissions")
	}
	sysflags := unix.O_CREAT | unix.O_RDWR | unix.O_CLOEXEC
	if flag&os.O_EXCL != 0 {
		sysflags |= unix.O_EXCL
	}
	attrs := &linuxMqAttr{Maxmsg: maxQueueSize, Msgsize: maxMsgSize}
	id, err := mq_open(name, sysflags, uint32(perm), attrs)
	if err != nil {
		return nil, errors.Wrap(err, "mq_open failed")
	}
	return &LinuxMessageQueue{
		id:           id,
		name:         name,
		cancelSocket: -1,
		inputBuff:    make([]byte, maxMsgSize),
		flags:        flag,
	}, nil
}

// OpenLinuxMessageQueue opens an existing message queue. It returns an error, if it does not exist.
//	name - unique mq name.
//	flag - flag is a combination of (os.O_RDONLY or os.O_WRONLY or os.O_RDWR) and O_NONBLOCK.
//		O_RDONLY
//			Open the queue to receive messages only.
//		O_WRONLY
//			Open the queue to send messages only.
//		O_RDWR
//			Open the queue to both send and receive messages.
func OpenLinuxMessageQueue(name string, flag int) (*LinuxMessageQueue, error) {
	id, err := mq_open(name, common.FlagsForAccess(flag)|unix.O_CLOEXEC, uint32(0), nil)
	if err != nil {
		return nil, errors.Wrap(err, "mq_open failed")
	}
	result := &LinuxMessageQueue{
		id:           id,
		name:         name,
		cancelSocket: -1,
		flags:        flag,
	}
	attrs, err := result.getAttrs()
	if err != nil {
		result.Close()
		return nil, errors.Wrap(err, "failed to get mq attrs")
	}
	result.inputBuff = make([]byte, attrs.Msgsize)
	return result, nil
}

// SendTimeoutPriority sends a message with a given priority.
// It blocks if the queue is full, waiting for a message unless timeout is passed.
func (mq *LinuxMessageQueue) SendTimeoutPriority(data []byte, prio int, timeout time.Duration) error {
	return common.UninterruptedSyscallTimeout(func(curTimeout time.Duration) error {
		return mq_timedsend(mq.ID(), data, prio, common.AbsTimeoutToTimeSpec(curTimeout))
	}, timeout)
}

// SendPriority sends a message with a given priority.
// It blocks if the queue is full.
func (mq *LinuxMessageQueue) SendPriority(data []byte, prio int) error {
	return mq.SendTimeoutPriority(data, prio, time.Duration(-1))
}

// SendTimeout sends a message with a default (0) priority.
// It blocks if the queue is full, waiting for a message unless timeout is passed.
func (mq *LinuxMessageQueue) SendTimeout(data []byte, timeout time.Duration) error {
	return mq.SendTimeoutPriority(data, 0, timeout)
}

// Send sends a message with a default (0) priority.
// It blocks if the queue is full.
func (mq *LinuxMessageQueue) Send(data []byte) error {
	timeout := time.Duration(-1)
	if mq.flags&O_NONBLOCK != 0 {
		timeout = time.Duration(0)
	}
	return mq.SendTimeoutPriority(data, 0, timeout)
}

// ReceiveTimeoutPriority receives a message, returning its priority.
// It blocks if the queue is empty, waiting for a message unless timeout is passed.
// Returns message len and priority.
func (mq *LinuxMessageQueue) ReceiveTimeoutPriority(input []byte, timeout time.Duration) (int, int, error) {
	dataToReceive := input
	curMaxMsgSize := len(mq.inputBuff)
	if len(input) < curMaxMsgSize {
		dataToReceive = mq.inputBuff
	}
	var prio, actualMsgSize, maxMsgSize int
	err := common.UninterruptedSyscallTimeout(func(curTimeout time.Duration) error {
		var err error
		actualMsgSize, maxMsgSize, err = mq_timedreceive(
			mq.ID(),
			dataToReceive,
			&prio,
			common.AbsTimeoutToTimeSpec(curTimeout))
		return err
	}, timeout)
	if maxMsgSize != 0 && actualMsgSize != 0 {
		if curMaxMsgSize != maxMsgSize {
			mq.inputBuff = make([]byte, maxMsgSize)
		}
	}
	if err != nil {
		return 0, 0, errors.Wrap(err, "linux mq: receive failed")
	}
	if len(input) < curMaxMsgSize {
		if len(input) < actualMsgSize {
			return 0, 0, errors.Errorf("the buffer of %d bytes is too small for a %d bytes message", len(input), actualMsgSize)
		}
		copy(input, dataToReceive[:actualMsgSize])
	}
	return actualMsgSize, prio, nil
}

// ReceivePriority receives a message, returning its priority.
// It blocks if the queue is empty. Returns message len and priority.
func (mq *LinuxMessageQueue) ReceivePriority(data []byte) (int, int, error) {
	return mq.ReceiveTimeoutPriority(data, time.Duration(-1))
}

// ReceiveTimeout receives a message.
// It blocks if the queue is empty, waiting for a message unless timeout is passed.
// Returns message len.
func (mq *LinuxMessageQueue) ReceiveTimeout(data []byte, timeout time.Duration) (int, error) {
	len, _, err := mq.ReceiveTimeoutPriority(data, timeout) // ignore proirity
	return len, err
}

// Receive receives a message. It blocks if the queue is empty.
// Returns message len.
func (mq *LinuxMessageQueue) Receive(data []byte) (int, error) {
	timeout := time.Duration(-1)
	if mq.flags&O_NONBLOCK != 0 {
		timeout = time.Duration(0)
	}
	len, _, err := mq.ReceiveTimeoutPriority(data, timeout) // ignore priority
	return len, err
}

// ID returns unique id of the queue.
func (mq *LinuxMessageQueue) ID() int {
	return mq.id
}

// Close closes the queue.
func (mq *LinuxMessageQueue) Close() error {
	if mq.cancelSocket >= 0 {
		if err := mq.NotifyCancel(); err != nil {
			return errors.Wrap(err, "failed to cancel notifications")
		}
	}
	err := unix.Close(mq.ID())
	*mq = LinuxMessageQueue{cancelSocket: -1}
	return err
}

// Cap returns the size of the mq buffer.
func (mq *LinuxMessageQueue) Cap() int {
	attrs, err := mq.getAttrs()
	if err != nil {
		return 0
	}
	return attrs.Maxmsg
}

// SetBlocking sets whether the send/receive operations on the queue block.
// This applies to the current instance only.
func (mq *LinuxMessageQueue) SetBlocking(block bool) error {
	if block {
		mq.flags &= ^O_NONBLOCK
	} else {
		mq.flags |= O_NONBLOCK
	}
	return nil
}

// Destroy closes the queue and removes it permanently.
func (mq *LinuxMessageQueue) Destroy() error {
	name := mq.name
	if err := mq.Close(); err != nil {
		return errors.Wrap(err, "mq close failed")
	}
	return DestroyLinuxMessageQueue(name)
}

// Notify notifies about new messages in the queue by sending id of the queue to the channel.
// If there are messages in the queue, no notification will be sent
// unless all of them are read.
func (mq *LinuxMessageQueue) Notify(ch chan<- int) error {
	if ch == nil {
		return errors.Errorf("cannot notify on a nil-chan")
	}
	if mq.cancelSocket >= 0 {
		return errors.Errorf("notify has already been called")
	}
	notifySocket, cancelSocket, err := initLinuxMqNotifications(ch)
	if err != nil {
		return errors.Wrap(err, "unable to init notifications subsystem")
	}
	ndata := &notify_data{mq_id: mq.ID()}
	pndata := unsafe.Pointer(ndata)
	defer allocator.Use(pndata)
	ev := &sigevent{
		sigev_notify: cSIGEV_THREAD,
		sigev_signo:  int32(notifySocket),
		sigev_value:  sigval{sigval_ptr: uintptr(pndata)},
	}
	if err = mq_notify(mq.ID(), ev); err != nil {
		cancelLinuxMqNotifications(mq.cancelSocket)
		err = errors.Wrap(err, "mq_notify failed")
	} else {
		mq.cancelSocket = cancelSocket
	}
	return err
}

// NotifyCancel cancels notification subscribtion.
func (mq *LinuxMessageQueue) NotifyCancel() error {
	var err error
	if err = mq_notify(mq.ID(), nil); err == nil {
		if err = cancelLinuxMqNotifications(mq.cancelSocket); err != nil {
			err = errors.Wrap(err, "failed to cancel notifications")
		}
		mq.cancelSocket = -1
	} else {
		err = errors.Wrap(err, "mq_notify failed")
	}
	return err
}

// getAttrs returns attributes of the queue.
func (mq *LinuxMessageQueue) getAttrs() (*linuxMqAttr, error) {
	attrs := new(linuxMqAttr)
	if err := mq_getsetattr(mq.ID(), nil, attrs); err != nil {
		return nil, errors.Wrap(err, "mq_getsetattr failed")
	}
	return attrs, nil
}

// DestroyLinuxMessageQueue removes the queue permanently.
func DestroyLinuxMessageQueue(name string) error {
	err := mq_unlink(name)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err, "mq_unlink failed")
		}
	}
	return err
}

// SetLinuxMqBlocking sets whether the operations on a linux mq block.
// This will apply for all send/receive operations on any instance of the
// linux mq with the given name.
func SetLinuxMqBlocking(name string, block bool) error {
	mq, err := OpenLinuxMessageQueue(name, os.O_RDWR)
	if err != nil {
		return errors.Wrap(err, "mq open failed")
	}
	attrs := new(linuxMqAttr)
	if !block {
		attrs.Flags |= unix.O_NONBLOCK
	}
	return mq_getsetattr(mq.ID(), attrs, nil)
}
