// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mq

import (
	"os"
	"time"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	ipc_sync "bitbucket.org/avd/go-ipc/sync"

	"github.com/pkg/errors"
)

const (
	// DefaultFastMqMaxSize is the default fast mq queue size.
	DefaultFastMqMaxSize = 8
	// DefaultFastMqMessageSize is the fast mq message size.
	DefaultFastMqMessageSize = 8192
)

// this is to ensure, that FastMq satisfies queue interfaces.
var (
	_ Messenger         = (*FastMq)(nil)
	_ PriorityMessenger = (*FastMq)(nil)
)

var (
	mqFullError  = newTemporaryError(errors.New("the queue is full"))
	mqEmptyError = newTemporaryError(errors.New("the queue is empty"))
)

// FastMq is a priority message queue based on shared memory.
// It does not support blocking send/receieve yet, and will panic,
// if opened without O_NONBLOCK.
// Currently it is the only implementation for windows.
// It'll become default implementation for all platforms,
// when bloking mode support is added.
type FastMq struct {
	name   string
	region *mmf.MemoryRegion
	flag   int
	locker ipc_sync.IPCLocker
	impl   *sharedHeap
}

func openFastMq(name string, flag int, perm os.FileMode, maxQueueSize, maxMsgSize int) (*FastMq, error) {
	var result *FastMq
	if !checkMqPerm(perm) {
		return nil, errors.New("invalid mq permissions")
	}
	openFlags := common.FlagsForOpen(flag)
	// calc size for all messages and metadata.
	size, err := calcSharedHeapSize(maxQueueSize, maxMsgSize)
	if err != nil {
		return nil, errors.Wrap(err, "mq size check failed")
	}

	// create a shared memory object for the queue.
	obj, created, err := shm.NewMemoryObjectSize(name, openFlags, perm, int64(size))
	if err != nil {
		return nil, errors.Wrap(err, "fast mq: failed to open/create shm object")
	}
	result = &FastMq{flag: flag, name: name}
	defer func() {
		fastMqCleanup(result, obj, created, err)
	}()

	// mmap memory object.
	result.region, err = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, size)
	if err != nil {
		return nil, errors.Wrap(err, "fast mq: failed to create new shm region")
	}

	// cleanup previous mutex instances. it could be useful in a case,
	// when previous mutex owner crashed, and the mutex is in incosistient state.
	if created {
		if err = ipc_sync.DestroyMutex(fastMqLockerName(name)); err != nil {
			return nil, errors.Wrap(err, "fast mq: failed to access a locker")
		}
	}
	result.locker, err = ipc_sync.NewMutex(fastMqLockerName(name), openFlags, perm)
	if err != nil {
		return nil, errors.Wrap(err, "fast mq: failed to create a locker")
	}

	// impl is an object placed into in mmaped area.
	rawData := allocator.ByteSliceData(result.region.Data())
	if created {
		result.impl = newSharedHeap(rawData, maxQueueSize, maxMsgSize)
	} else {
		result.impl = openSharedHeap(rawData)
	}
	return result, err
}

// CreateFastMq creates new FastMq.
//	name - mq name. implementation will create a shm object with this name.
//	flag - flag is a combination of os.O_EXCL, and O_NONBLOCK.
//	perm - object's permission bits.
//	maxQueueSize - queue capacity.
//	maxMsgSize - maximum message size.
func CreateFastMq(name string, flag int, perm os.FileMode, maxQueueSize, maxMsgSize int) (*FastMq, error) {
	return openFastMq(name, flag|os.O_CREATE, perm, maxQueueSize, maxMsgSize)
}

// OpenFastMq opens an existing message queue. It returns an error, if it does not exist.
//	name - unique mq name.
//	flag - 0 or O_NONBLOCK.
func OpenFastMq(name string, flag int) (*FastMq, error) {
	maxQueueSize, maxMsgSize, err := FastMqAttrs(name)
	if err != nil {
		return nil, err
	}
	return openFastMq(name, flag&O_NONBLOCK, 0666, maxQueueSize, maxMsgSize)
}

// FastMqAttrs returns capacity and max message size of the existing mq.
func FastMqAttrs(name string) (int, int, error) {
	obj, err := shm.NewMemoryObject(name, os.O_RDONLY, 0666)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to open shm object")
	}
	defer obj.Close()
	minSize := minHeapSize()
	if int(obj.Size()) < minSize {
		return 0, 0, errors.New("shm object is too small")
	}
	region, err := mmf.NewMemoryRegion(obj, mmf.MEM_READ_ONLY, 0, minSize)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to create new shm region")
	}
	defer region.Close()
	heap := openSharedHeap(allocator.ByteSliceData(region.Data()))
	return heap.maxSize(), heap.maxMsgSize(), nil
}

// DestroyFastMq permanently removes a FastMq.
func DestroyFastMq(name string) error {
	errMutex := ipc_sync.DestroyMutex(name)
	if errObject := shm.DestroyMemoryObject(name); errObject != nil {
		return errors.Wrap(errObject, "failed to destroy memory object")
	}
	if errMutex != nil {
		return errors.Wrap(errMutex, "failed to destroy ipc locker")
	}
	return nil
}

// Send sends a message. It blocks if the queue is full.
func (mq *FastMq) Send(data []byte) error {
	return mq.SendPriority(data, 2)
}

// SendPriority sends a message with the given priority. It blocks if the queue is full.
func (mq *FastMq) SendPriority(data []byte, prio int) error {
	return mq.SendPriorityTimeout(data, prio, -1)
}

// SendPriorityTimeout sends a message with the given priority. It blocks if the queue is full,
// waiting for not longer, then the timeout.
func (mq *FastMq) SendPriorityTimeout(data []byte, prio int, timeout time.Duration) error {
	if len(data) > mq.impl.maxMsgSize() {
		return errors.New("the message is too big")
	}
	// optimization: do lock the locker if the queue is full.
	if mq.impl.Len() == mq.impl.maxSize() && mq.flag&O_NONBLOCK != 0 {
		return mqFullError
	}
	mq.locker.Lock()
	// defer is not used due to performance reasons.

	if mq.impl.Len() == mq.impl.maxSize() {
		mq.locker.Unlock()
		if mq.flag&O_NONBLOCK != 0 {
			return mqFullError
		}
		if timeout >= 0 {
		} else {
		}
		panic("blocking send is not implemented yet")
	}

	mq.impl.pushMessage(message{data: data, prio: int32(prio)})
	mq.locker.Unlock()
	return nil
}

// Receive receives a message. It blocks if the queue is empty.
func (mq *FastMq) Receive(data []byte) error {
	_, err := mq.ReceivePriorityTimeout(data, -1)
	return err
}

// ReceivePriority receives a message and returns its priority. It blocks if the queue is empty.
func (mq *FastMq) ReceivePriority(data []byte) (int, error) {
	return mq.ReceivePriorityTimeout(data, -1)
}

// ReceivePriorityTimeout receives a message and returns its priority. It blocks if the queue is empty,
// waiting for not longer, then the timeout.
func (mq *FastMq) ReceivePriorityTimeout(data []byte, timeout time.Duration) (int, error) {

	// optimization: do lock the locker if the queue is empty.
	if mq.impl.Len() == 0 && mq.flag&O_NONBLOCK != 0 {
		return 0, mqEmptyError
	}

	mq.locker.Lock()
	// defer is not used due to performance reasons.

	if mq.impl.Len() == 0 {
		mq.locker.Unlock()
		if mq.flag&O_NONBLOCK != 0 {
			return 0, mqEmptyError
		}
		if timeout >= 0 {
		} else {
		}
		panic("blocking receive is not implemented yet")
	}
	prio, err := mq.impl.popMessage(data)
	mq.locker.Unlock()
	return prio, err
}

// Cap returns the size of the mq buffer.
func (mq *FastMq) Cap() (int, error) {
	return mq.impl.maxSize(), nil
}

// SetBlocking sets whether the send/receive operations on the queue block.
// This applies to the current instance only.
func (mq *FastMq) SetBlocking(block bool) error {
	if block {
		mq.flag &= ^O_NONBLOCK
	} else {
		mq.flag |= O_NONBLOCK
	}
	return nil
}

// Close closes a FastMq instance.
func (mq *FastMq) Close() error {
	errLocker := mq.locker.Close()
	if errRegion := mq.region.Close(); errRegion != nil {
		return errors.Wrap(errRegion, "failed to close memory region")
	}
	if errLocker != nil {
		return errors.Wrap(errLocker, "failed to close ipc locker")
	}
	return nil
}

// Destroy permanently removes a FastMq instance.
func (mq *FastMq) Destroy() error {
	errClose := mq.Close()
	if errDestroy := DestroyFastMq(mq.name); errDestroy != nil {
		return errors.Wrap(errDestroy, "failed to destroy fastmq")
	}
	if errClose != nil {
		return errors.Wrap(errClose, "failed to close fastmq")
	}
	return nil
}

func fastMqLockerName(mqName string) string {
	return mqName + ".locker"
}

func fastMqCleanup(mq *FastMq, obj shm.SharedMemoryObject, created bool, err error) {
	obj.Close()
	if err == nil {
		return
	}
	if mq.region != nil {
		mq.region.Close()
	}
	if mq.locker != nil {
		mq.locker.Close()
		if created {
			if d, ok := mq.locker.(ipc.Destroyer); ok {
				d.Destroy()
			}
		}
	}
	if created {
		obj.Destroy()
	}
}
