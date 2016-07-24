// Copyright 2016 Aleksandr Demakin. All rights reserved.

package array

import (
	"sync/atomic"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
)

const (
	mappedArrayHdrSize = unsafe.Sizeof(mappedArray{})
)

type mappedArray struct {
	capacity       int32
	elemSize       int32
	size           int32
	dummyDataArray [0]byte
}

func newMappedArray(pointer unsafe.Pointer) *mappedArray {
	return (*mappedArray)(pointer)
}

func (arr *mappedArray) init(capacity, elemSize int) {
	arr.capacity = int32(capacity)
	arr.elemSize = int32(elemSize)
	arr.size = 0
}

func (arr *mappedArray) elemLen() int {
	return int(arr.elemSize)
}

func (arr *mappedArray) cap() int {
	return int(arr.capacity)
}

func (arr *mappedArray) len() int {
	return int(atomic.LoadInt32(&arr.size))
}

func (arr *mappedArray) incLen() {
	atomic.AddInt32(&arr.size, 1)
}

func (arr *mappedArray) decLen() {
	atomic.AddInt32(&arr.size, -1)
}

func (arr *mappedArray) atPointer(idx int) unsafe.Pointer {
	slotsPtr := uintptr(unsafe.Pointer(&arr.dummyDataArray))
	slotsPtr += uintptr(idx * int(arr.elemSize))
	return unsafe.Pointer(slotsPtr)
}

func (arr *mappedArray) at(idx int) []byte {
	slotsPtr := arr.atPointer(idx)
	return allocator.ByteSliceFromUnsafePointer(slotsPtr, int(arr.elemSize), int(arr.elemSize))
}
