// Copyright 2016 Aleksandr Demakin. All rights reserved.

package array

import (
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/nxgtw/go-ipc/internal/allocator"

	"github.com/stretchr/testify/assert"
)

func TestSharedArray(t *testing.T) {
	a := assert.New(t)
	sl := make([]byte, CalcSharedArraySize(10, 8))
	arr := NewSharedArray(allocator.ByteSliceData(sl), 10, 8)
	a.Equal(arr.Len(), 0)
	a.Panics(func() {
		arr.PopFront()
	})
	data := make([]byte, 1)
	for i := 0; i < 10; i++ {
		data[0] = byte(i)
		arr.PushBack(data)
	}
	a.Equal(arr.Len(), 10)
	a.Panics(func() {
		arr.PushBack()
	})
	for i := 0; i < 10; i++ {
		data[0] = byte(i)
		a.Equal(data, arr.At(i))
	}
	a.Equal(10, arr.Len())
	a.Panics(func() {
		arr.PushBack(data)
	})
	a.NotPanics(func() {
		copy(data, arr.At(0))
		arr.PopFront()
	})
	a.Equal(arr.Len(), 9)
	a.Equal([]byte{0}, data)
	for i := 0; i < 9; i++ {
		data[0] = byte(i + 1)
		a.Equal(data, arr.At(i))
	}
	arr.Swap(0, 8)
	a.Equal(arr.Len(), 9)
	a.Equal([]byte{9}, arr.At(0))
	a.Equal([]byte{1}, arr.At(8))
	l := arr.Len()
	for i := 0; i < l; i++ {
		a.NotPanics(func() {
			a.Equal(arr.Len(), 9-i)
			arr.PopFront()
		})
	}
	a.Equal(0, arr.Len())
	a.NotPanics(func() {
		data[0] = 13
		arr.PushBack(data)
	})
	a.NotPanics(func() {
		data[0] = 255
		arr.PushBack(data)
	})
	a.NotPanics(func() {
		arr.Swap(0, 1)
	})
	a.Equal(arr.Len(), 2)
	a.Equal([]byte{255}, arr.At(0))
	a.Equal([]byte{13}, arr.At(1))
}

type sorter struct {
	a *SharedArray
}

func (s sorter) Len() int {
	return s.a.Len()
}

func (s sorter) Less(i, j int) bool {
	l, r := s.a.At(i), s.a.At(j)
	return l[0] < r[0]
}

func (s sorter) Swap(i, j int) {
	s.a.Swap(i, j)
}

func TestSharedArray2(t *testing.T) {
	data := [...]int{8, 4, 7, 1, 0, 15, 2, 4, 10}
	a := assert.New(t)
	sl := make([]byte, CalcSharedArraySize(len(data), 8))
	arr := NewSharedArray(allocator.ByteSliceData(sl), len(data), 8)
	for i := 0; i < len(data); i++ {
		arr.PushBack([]byte{byte(data[i])})
	}
	for outer := 1; outer < len(data); outer++ {
		sort.Ints(data[outer:])
		arr.PopFront()
		sort.Sort(&sorter{a: arr})
		for i, b := range data[outer:] {
			a.Equal(byte(b), arr.At(i)[0])
		}
	}
}

func TestSharedArray3(t *testing.T) {
	data := [...]int{8, 4, 7, 1, 0, 15, 2, 4, 10}
	a := assert.New(t)
	sl := make([]byte, CalcSharedArraySize(len(data), 8))
	arr := NewSharedArray(allocator.ByteSliceData(sl), len(data), 8)
	for i := 0; i < len(data); i++ {
		arr.PushBack([]byte{byte(data[i])})
	}
	d := data[:]
	rand.Seed(time.Now().Unix())
	for i := 0; i < len(data); i++ {
		idx := rand.Intn(len(d))
		popped := arr.At(idx)
		a.Equal(d[idx], int(popped[0]))
		arr.RemoveAt(idx)
		d = append(d[:idx], d[idx+1:]...)
	}
	a.Equal(0, arr.Len())
}
