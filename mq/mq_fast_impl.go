// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mq

import (
	"unsafe"

	"github.com/nxgtw/go-ipc/internal/allocator"
)

const (
	fastMqHdrSize = int(unsafe.Sizeof(fastMqHdr{}))
)

type fastMqHdr struct {
	blockedSenders   int32
	blockedReceivers int32
}

type fastMq struct {
	header *fastMqHdr
	heap   *sharedHeap
}

func newFastMq(data []byte, maxQueueSize, maxMsgSize int, created bool) *fastMq {
	rawData := allocator.ByteSliceData(data)
	result := &fastMq{header: (*fastMqHdr)(rawData)}
	rawData = allocator.AdvancePointer(rawData, uintptr(fastMqHdrSize))
	if created {
		result.heap = newSharedHeap(rawData, maxQueueSize, maxMsgSize)
		result.header.blockedReceivers = 0
		result.header.blockedSenders = 0
	} else {
		result.heap = openSharedHeap(rawData)
	}
	return result
}

// calcFastMqSize returns number of bytes needed to store all messages and metadata.
func calcFastMqSize(maxQueueSize, maxMsgSize int) (int, error) {
	sz, err := calcSharedHeapSize(maxQueueSize, maxMsgSize)
	if err != nil {
		return 0, err
	}
	return fastMqHdrSize + sz, nil
}

func minFastMqSize() int {
	return fastMqHdrSize + minHeapSize()
}
