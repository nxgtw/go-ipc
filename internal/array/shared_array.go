// Copyright 2016 Aleksandr Demakin. All rights reserved.

package array

import (
	"math"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
)

const (
	indexEntrySize = unsafe.Sizeof(indexEntry{})
)

type indexEntry struct {
	len     int32
	slotIdx int32
}

type index struct {
	entries []indexEntry
	bitmap  []uint64
	headIdx *int32
}

func bitmapSize(sz int) int {
	bitmapSize := sz / 64
	if sz%64 != 0 {
		bitmapSize++
	}
	return bitmapSize
}

func indexSize(sz int) int {
	return sz*int(indexEntrySize) + bitmapSize(sz)*8 + 4
}

func newIndex(raw unsafe.Pointer, sz int) index {
	bitmapSz := bitmapSize(sz)
	rawIndexSlice := allocator.RawSliceFromUnsafePointer(raw, sz, sz)
	raw = allocator.AvdancePointer(raw, uintptr(sz)*indexEntrySize)
	rawBitmapSlize := allocator.RawSliceFromUnsafePointer(raw, bitmapSz, bitmapSz)
	raw = allocator.AvdancePointer(raw, 8*uintptr(bitmapSz))
	return index{
		entries: *(*[]indexEntry)(rawIndexSlice),
		bitmap:  *(*[]uint64)(rawBitmapSlize),
		headIdx: (*int32)(raw),
	}
}

func lowestZeroBit(value uint64) uint8 {
	for bitIdx := uint8(0); bitIdx < 64; bitIdx++ {
		if value&1 == 0 {
			return bitIdx
		}
		value >>= 1
	}
	return math.MaxUint8
}

func (idx *index) reserveFreeSlot(at int) {
	for i, b := range idx.bitmap {
		if b != math.MaxUint64 {
			bitIdx := lowestZeroBit(b)
			idx.entries[at].slotIdx = int32(i*64 + int(bitIdx))
			idx.entries[at].len = 0
			idx.bitmap[i] |= (1 << bitIdx)
			return
		}
	}
	panic("no free slots")
}

func (idx index) freeSlot(at int) {
	slotIdx := idx.entries[at].slotIdx
	bucketIdx, bitIdx := slotIdx/64, slotIdx%64
	bucket := &idx.bitmap[bucketIdx]
	*bucket = (*bucket) & ^(1 << uint32(bitIdx))
}

// SharedArray is an array placed in the shared memory with fixed length and element size.
// It is possible to swap elements and pop them from any position. It never moves elements
// in memory, so can be used to implements an array of futexes or spin locks.
type SharedArray struct {
	data *mappedArray
	idx  index
}

// NewSharedArray initializes new shared array with size and element size.
func NewSharedArray(raw unsafe.Pointer, size, elemSize int) *SharedArray {
	data := newMappedArray(raw)
	data.init(size, elemSize)
	idx := newIndex(allocator.AvdancePointer(raw, mappedArrayHdrSize+uintptr(size*elemSize)), size)
	return &SharedArray{
		data: data,
		idx:  idx,
	}
}

// OpenSharedArray opens existing shared array.
func OpenSharedArray(raw unsafe.Pointer) *SharedArray {
	data := newMappedArray(raw)
	idx := newIndex(allocator.AvdancePointer(raw, mappedArrayHdrSize+uintptr(data.cap()*data.elemLen())), data.cap())
	return &SharedArray{
		data: data,
		idx:  idx,
	}
}

// Cap returns array's cpacity
func (arr *SharedArray) Cap() int {
	return arr.data.cap()
}

// Len returns current length.
func (arr *SharedArray) Len() int {
	return arr.data.len()
}

// ElemSize returns size of the element.
func (arr *SharedArray) ElemSize() int {
	return arr.data.elemLen()
}

// PushBack add new element to the end of the array, merging given datas.
// Returns number of bytes copied, less or equal, than the size of the element.
func (arr *SharedArray) PushBack(datas ...[]byte) int {
	curLen := arr.Len()
	if curLen >= arr.Cap() {
		panic("index out of range")
	}
	last := arr.entryAt(curLen)
	arr.idx.reserveFreeSlot(arr.logicalIdxToPhys(curLen))
	dataPtr := arr.data.atPointer(int(last.slotIdx))
	slData := allocator.ByteSliceFromUnsafePointer(dataPtr, arr.ElemSize(), arr.ElemSize())
	for _, data := range datas {
		last.len += int32(copy(slData[last.len:], data))
		if int(last.len) < len(data) {
			break
		}
	}
	arr.data.incLen()
	return int(last.len)
}

// At returns data at the position i.
func (arr *SharedArray) At(i int) []byte {
	if i < 0 || i >= arr.Len() {
		panic("index out of range")
	}
	entry := arr.entryAt(i)
	return arr.data.at((int(entry.slotIdx)))[:int(entry.len):int(entry.len)]
}

// AtPointer returns pointer to the data at the position i.
func (arr *SharedArray) AtPointer(i int) unsafe.Pointer {
	if i < 0 || i >= arr.Len() {
		panic("index out of range")
	}
	entry := arr.entryAt(i)
	return arr.data.atPointer(int(entry.slotIdx))
}

// PopFront removes the first element of the array, writing its data to 'data'.
func (arr *SharedArray) PopFront(data []byte) {
	curLen := arr.Len()
	if curLen == 0 {
		panic("index out of range")
	}
	if data != nil {
		toCopy := arr.At(0)
		copy(data, toCopy)
	}
	arr.idx.freeSlot(int(*arr.idx.headIdx))
	arr.forwardHead()
	arr.data.decLen()
}

func (arr *SharedArray) forwardHead() {
	if arr.Len() == 1 {
		*arr.idx.headIdx = 0
	} else {
		*arr.idx.headIdx = (*arr.idx.headIdx + 1) % int32(arr.Cap())
	}
}

// PopBack removes the last element of the array, writing its data to 'data'.
func (arr *SharedArray) PopBack(data []byte) {
	curLen := arr.Len()
	if curLen == 0 {
		panic("index out of range")
	}
	if data != nil {
		toCopy := arr.At(curLen - 1)
		copy(data, toCopy)
	}
	arr.idx.freeSlot(arr.logicalIdxToPhys(curLen - 1))
	if curLen == 1 {
		*arr.idx.headIdx = 0
	}
	arr.data.decLen()
}

// PopAt removes i'th element of the array, writing its data to 'data'.
func (arr *SharedArray) PopAt(idx int, data []byte) {
	curLen := arr.Len()
	if idx < 0 || idx >= curLen {
		panic("index out of range")
	}
	if data != nil {
		toCopy := arr.At(idx)
		copy(data, toCopy)
	}
	arr.idx.freeSlot(arr.logicalIdxToPhys(idx))
	if idx <= curLen/2 {
		for i := idx; i > 0; i-- {
			arr.idx.entries[arr.logicalIdxToPhys(i)] = arr.idx.entries[arr.logicalIdxToPhys(i-1)]
		}
		arr.forwardHead()
	} else {
		for i := idx; i < curLen-1; i++ {
			arr.idx.entries[arr.logicalIdxToPhys(i)] = arr.idx.entries[arr.logicalIdxToPhys(i+1)]
		}
		if curLen == 1 {
			*arr.idx.headIdx = 0
		}
	}
	arr.data.decLen()
}

// Swap swaps two elements of the array.
func (arr *SharedArray) Swap(i, j int) {
	l := arr.Len()
	if i >= l || j >= l {
		panic("index out of range")
	}
	if i == j {
		return
	}
	i, j = arr.logicalIdxToPhys(i), arr.logicalIdxToPhys(j)
	arr.idx.entries[i], arr.idx.entries[j] = arr.idx.entries[j], arr.idx.entries[i]
}

func (arr *SharedArray) logicalIdxToPhys(log int) int {
	return (log + int(*arr.idx.headIdx)) % arr.Cap()
}

func (arr *SharedArray) entryAt(log int) *indexEntry {
	return &arr.idx.entries[arr.logicalIdxToPhys(log)]
}

// CalcSharedArraySize returns the size, needed to place shared array in memory.
func CalcSharedArraySize(size, elemSize int) int {
	return int(mappedArrayHdrSize) + // mq header
		indexSize(size) + // mq index
		size*elemSize // mq messages size
}
