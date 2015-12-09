// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"sync"
	"testing"

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

	assert.NoError(t, checkObject(i))
	assert.NoError(t, checkObject(c))
	assert.NoError(t, checkObject(arr))
	assert.NoError(t, checkObject(arr[:]))
	assert.NoError(t, checkObject(validStruct{}))
	assert.Error(t, checkObject(invalidStruct1{}))
	assert.Error(t, checkObject(invalidStruct2{}))
	assert.Error(t, checkObject(invalidStruct3{}))
	assert.Error(t, checkObject(arr2))
	assert.Error(t, checkObject(arr2[:]))
	assert.NoError(t, checkObject(sync.Mutex{}))
	assert.Error(t, checkObject(m))

	assert.Error(t, checkObject(slsl))
}
