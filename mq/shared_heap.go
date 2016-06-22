// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mq

import (
	"container/heap"
	"sync/atomic"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"github.com/pkg/errors"
)

const (
	sharedMqHdrSize = unsafe.Sizeof(sharedMqHdr{})
	indexEntrySize  = unsafe.Sizeof(indexEntry{})
	msgSlotSize     = unsafe.Sizeof(msgSlot{})
)

type sharedHeap struct {
	header   *sharedMqHdr
	index    []indexEntry
	slotsPtr unsafe.Pointer
}

type indexEntry struct {
	msgLen  int32
	msgPrio int32
	slotIdx int32
}

type msgSlot struct {
	entryIdx       int32
	dummyDataArray [0]byte
}

func newSharedHeap(raw []byte, maxQueueSize, maxMsgSize int, needInit bool) *sharedHeap {
	header := (*sharedMqHdr)(allocator.ByteSliceData(raw))
	if needInit { // initialize shared data.
		header.maxQueueSize = int32(maxQueueSize)
		header.maxMsgSize = int32(maxMsgSize)
		header.size = 0
	}
	rawIndexSlice := allocator.RawSliceFromUnsafePointer(
		unsafe.Pointer(&header.dummyDataArray),
		int(header.maxQueueSize),
		int(header.maxQueueSize))
	index := *(*[]indexEntry)(rawIndexSlice)
	slotsPtr := uintptr(unsafe.Pointer(&header.dummyDataArray)) + uintptr(maxQueueSize)*indexEntrySize
	return &sharedHeap{
		header:   header,
		index:    index,
		slotsPtr: unsafe.Pointer(slotsPtr),
	}
}

func (mq *sharedHeap) full() bool {
	return atomic.LoadInt32(&mq.header.size) == mq.header.maxQueueSize
}

func (mq *sharedHeap) empty() bool {
	return atomic.LoadInt32(&mq.header.size) == 0
}

func (mq *sharedHeap) maxMsgSize() int {
	return int(mq.header.maxMsgSize)
}

func (mq *sharedHeap) maxSize() int {
	return int(mq.header.maxQueueSize)
}

func (mq *sharedHeap) top() message {
	topEntry := mq.index[0]
	currentSlot := mq.slotAt(topEntry.slotIdx)
	return message{prio: int(topEntry.msgPrio), data: slotData(currentSlot, topEntry.msgLen)}
}

func (mq *sharedHeap) push(msg message) {
	heap.Push(mq, msg)
}

func (mq *sharedHeap) pop(data []byte) int {
	topEntry := mq.index[0]
	currentSlot := mq.slotAt(topEntry.slotIdx)
	copy(data, slotData(currentSlot, topEntry.msgLen))
	prio := topEntry.msgPrio
	heap.Pop(mq)
	return int(prio)
}

// sort.Interface

func (mq *sharedHeap) Len() int {
	return int(mq.header.size)
}

func (mq *sharedHeap) Less(i, j int) bool {
	// inverse less logic. as we want max-heap.
	return mq.index[i].msgPrio > mq.index[j].msgPrio
}

func (mq *sharedHeap) Swap(i, j int) {
	if i == j {
		return
	}
	lSlot, rSlot := mq.slotAt(int32(i)), mq.slotAt(int32(j))
	lSlot.entryIdx, rSlot.entryIdx = int32(j), int32(i)
	mq.index[i], mq.index[j] = mq.index[j], mq.index[i]
}

// heap.Interface

func (mq *sharedHeap) Push(x interface{}) {
	msg := x.(message)
	last := &mq.index[mq.header.size]
	last.msgPrio = int32(msg.prio)
	last.msgLen = int32(len(msg.data))
	last.slotIdx = mq.header.size
	slot := mq.slotAt(mq.header.size)
	slot.entryIdx = mq.header.size
	data := slotData(slot, last.msgLen)
	copy(data, msg.data)
	atomic.AddInt32(&mq.header.size, 1)
}

func (mq *sharedHeap) Pop() interface{} {
	atomic.AddInt32(&mq.header.size, -1)
	lastIndexEntry := mq.index[mq.header.size]
	if lastIndexEntry.slotIdx != mq.header.size {
		currentSlot := mq.slotAt(lastIndexEntry.slotIdx)
		lastSlot := mq.slotAt(mq.header.size)
		lastSlotIndexEntry := mq.index[lastSlot.entryIdx]
		copy(slotData(currentSlot, lastSlotIndexEntry.msgLen), slotData(lastSlot, lastSlotIndexEntry.msgLen))
		lastSlotIndexEntry.slotIdx = lastIndexEntry.slotIdx
	}
	return nil // return value is not used by sharedHeap.pop.
}

func (mq *sharedHeap) slotAt(idx int32) *msgSlot {
	raw := uintptr(mq.slotsPtr)
	raw += uintptr(idx * (mq.header.maxMsgSize + int32(msgSlotSize)))
	return (*msgSlot)(unsafe.Pointer(raw))
}

func slotData(slot *msgSlot, size int32) []byte {
	return allocator.ByteSliceFromUnsafePointer(unsafe.Pointer(&slot.dummyDataArray), int(size), int(size))
}

func calcSharedHeapSize(maxQueueSize, maxMsgSize int) (int, error) {
	if maxQueueSize == 0 || maxMsgSize == 0 {
		return 0, errors.New("queue size cannot be zero")
	}
	return int(sharedMqHdrSize) + // mq header
		maxQueueSize*int(indexEntrySize) + // mq index
		maxQueueSize*(maxMsgSize+int(msgSlotSize)), nil // mq messages size
}
