// Copyright 2016 Aleksandr Demakin. All rights reserved.

package array

import (
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
)

const (
	indexEntrySize = unsafe.Sizeof(indexEntry{})
	slotSize       = unsafe.Sizeof(slot{})
)

// SharedArray is an array placed in the shared memory with fixed length and element size.
type SharedArray struct {
	data  *mappedArray
	index []indexEntry
}

type indexEntry struct {
	len     int32
	slotIdx int32
}

type slot struct {
	entryIdx       int32
	dummyDataArray [0]byte
}

// NewSharedArray initializes new shared array with size and element size.
func NewSharedArray(raw unsafe.Pointer, size, elemSize int) *SharedArray {
	data := newMappedArray(raw)
	data.init(size, int(slotSize)+elemSize)
	rawIndexSlice := allocator.RawSliceFromUnsafePointer(
		data.atPointer(size),
		size,
		size)
	index := *(*[]indexEntry)(rawIndexSlice)
	return &SharedArray{
		data:  data,
		index: index,
	}
}

// OpenSharedArray opens existing shared array.
func OpenSharedArray(raw unsafe.Pointer) *SharedArray {
	data := newMappedArray(raw)
	rawIndexSlice := allocator.RawSliceFromUnsafePointer(
		data.atPointer(data.cap()),
		data.cap(),
		data.cap())
	index := *(*[]indexEntry)(rawIndexSlice)
	return &SharedArray{
		data:  data,
		index: index,
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
	return arr.data.elemLen() - int(slotSize)
}

// PushBack add new element to the end of the array, merging given datas.
// Returns number of bytes copied, less or equal, than the size of the element.
func (arr *SharedArray) PushBack(datas ...[]byte) int {
	curLen := int32(arr.Len())
	if int(curLen) >= arr.Cap() {
		panic("index out of range")
	}
	last := &arr.index[curLen]
	last.slotIdx = curLen
	sl := arr.slotAt(int(curLen))
	sl.entryIdx = curLen
	slData := slotData(sl, arr.ElemSize())
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
	entry := arr.index[i]
	currentSlot := arr.slotAt(int(entry.slotIdx))
	return slotData(currentSlot, int(entry.len))
}

// AtPointer returns pointer to the data at the position i.
func (arr *SharedArray) AtPointer(i int) unsafe.Pointer {
	if i < 0 || i >= arr.Len() {
		panic("index out of range")
	}
	entry := arr.index[i]
	currentSlot := arr.slotAt(int(entry.slotIdx))
	return unsafe.Pointer(&currentSlot.dummyDataArray)
}

// PopFront removes the first element of the array, writing its data to 'data'.
func (arr *SharedArray) PopFront(data []byte) {
	if arr.Len() == 0 {
		panic("index out of range")
	}
	topEntry := arr.index[0]
	currentSlot := arr.slotAt(int(topEntry.slotIdx))
	if data != nil {
		copy(data, slotData(currentSlot, int(topEntry.len)))
	}
	arr.data.decLen()
	l := arr.Len()
	lastIndexEntry := arr.index[l]
	if int(lastIndexEntry.slotIdx) != l {
		currentSlot := arr.slotAt(int(lastIndexEntry.slotIdx))
		lastSlot := arr.slotAt(l)
		lastSlotIndexEntry := arr.index[lastSlot.entryIdx]
		copy(slotData(currentSlot, int(lastSlotIndexEntry.len)), slotData(lastSlot, int(lastSlotIndexEntry.len)))
		lastSlotIndexEntry.slotIdx = lastIndexEntry.slotIdx
	}
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
	lSlot, rSlot := arr.slotAt(i), arr.slotAt(j)
	lSlot.entryIdx, rSlot.entryIdx = int32(j), int32(i)
	arr.index[i], arr.index[j] = arr.index[j], arr.index[i]
}

func (arr *SharedArray) slotAt(idx int) *slot {
	return (*slot)(arr.data.atPointer(idx))
}

// CalcSharedArraySize returns the size, needed to place shared array in memory.
func CalcSharedArraySize(size, elemSize int) int {
	return int(mappedArrayHdrSize) + // mq header
		size*int(indexEntrySize) + // mq index
		size*(elemSize+int(slotSize)) // mq messages size
}

func slotData(s *slot, size int) []byte {
	return allocator.ByteSliceFromUnsafePointer(unsafe.Pointer(&s.dummyDataArray), size, size)
}
