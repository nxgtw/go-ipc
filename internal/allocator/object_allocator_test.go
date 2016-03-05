// Copyright 2015 Aleksandr Demakin. All rights reserved.

package allocator

import (
	"sync"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCheckObjectType(t *testing.T) {
	type validStruct struct {
		a, b int
		u    uintptr
		s    struct {
			arr [3]int
		}
	}
	type invalidStruct1 struct {
		a, b *int
	}
	type invalidStruct2 struct {
		a, b []int
	}
	type invalidStruct3 struct {
		s string
	}
	var i int
	var c complex128
	var arr = [3]int{}
	var arr2 = [3]string{}
	var slsl [][]int
	var m map[int]int

	assert.NoError(t, CheckObjectReferences(i))
	assert.NoError(t, CheckObjectReferences(c))
	assert.NoError(t, CheckObjectReferences(arr))
	assert.NoError(t, CheckObjectReferences(arr[:]))
	assert.NoError(t, CheckObjectReferences(validStruct{}))
	assert.NoError(t, CheckObjectReferences(sync.Mutex{}))

	assert.Error(t, CheckObjectReferences(invalidStruct1{}))
	assert.Error(t, CheckObjectReferences(invalidStruct2{}))
	assert.Error(t, CheckObjectReferences(invalidStruct3{}))
	assert.Error(t, CheckObjectReferences(arr2))
	assert.Error(t, CheckObjectReferences(arr2[:]))
	assert.Error(t, CheckObjectReferences(m))
	assert.Error(t, CheckObjectReferences(slsl))
}

func TestAllocInt(t *testing.T) {
	var i = 0x01027FFF
	data := make([]byte, unsafe.Sizeof(i))
	if !assert.NoError(t, Alloc(data, i)) {
		return
	}
	ptr := (*int)(unsafe.Pointer(&data[0]))
	assert.Equal(t, i, *ptr)
}

func TestAllocIntArray(t *testing.T) {
	i := [3]int{0x01, 0x7F, 0xFF}
	data := make([]byte, unsafe.Sizeof(i))
	if !assert.NoError(t, Alloc(data, i)) {
		return
	}
	ptr := (*[3]int)(unsafe.Pointer(&data[0]))
	assert.Equal(t, i, *ptr)
}

func TestAllocStruct(t *testing.T) {
	type internal struct {
		d complex128
		p uintptr
	}
	type s struct {
		a, b int
		ss   internal
	}
	obj := s{-1, 11, internal{complex(10, 11), uintptr(0)}}
	data := make([]byte, unsafe.Sizeof(obj))
	if !assert.NoError(t, Alloc(data, obj)) {
		return
	}
	ptr := (*s)(unsafe.Pointer(&data[0]))
	assert.Equal(t, obj, *ptr)
}

func TestAllocStructPtr(t *testing.T) {
	type internal struct {
		d complex128
		p uintptr
	}
	type s struct {
		a, b int
		ss   internal
	}
	obj := &s{-1, 11, internal{complex(10, 11), uintptr(0xDEADBEEF)}}
	data := make([]byte, unsafe.Sizeof(*obj))
	if !assert.NoError(t, Alloc(data, obj)) {
		return
	}
	ptr := (*s)(unsafe.Pointer(&data[0]))
	assert.Equal(t, obj, ptr)
}

func TestAllocMutex(t *testing.T) {
	var obj sync.Mutex
	data := make([]byte, unsafe.Sizeof(obj))
	if !assert.NoError(t, Alloc(data, obj)) {
		return
	}
	ptr := (*sync.Mutex)(unsafe.Pointer(&data[0]))
	assert.Equal(t, obj, *ptr)
}

func TestAllocSlice(t *testing.T) {
	obj := make([]int, 10)
	for i := range obj {
		obj[i] = int(i)
	}
	data := make([]byte, unsafe.Sizeof(int(0))*10)
	if !assert.NoError(t, Alloc(data, obj)) {
		return
	}
	sl := ByteSliceTointSlice(data, 10, 10)
	assert.Equal(t, obj, sl)
}

func TestAllocSliceReadAsArray(t *testing.T) {
	obj := make([]int, 10)
	for i := range obj {
		obj[i] = int(i)
	}
	data := make([]byte, unsafe.Sizeof(int(0))*10)
	if !assert.NoError(t, Alloc(data, obj)) {
		return
	}
	ptr := (*[10]int)(unsafe.Pointer(&data[0]))
	assert.Equal(t, obj, (*ptr)[:])
}

func TestAllocArrayReadAsSlice(t *testing.T) {
	i := [3]int{0x01, 0x7F, 0xFF}
	data := make([]byte, unsafe.Sizeof(i))
	if !assert.NoError(t, Alloc(data, i)) {
		return
	}
	sl := ByteSliceTointSlice(data, 3, 3)
	assert.Equal(t, i[:], sl)
}
