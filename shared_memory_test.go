// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	defaultObjectName = "go-ipc-test"
)

func TestCreateMemoryRegion(t *testing.T) {
	obj, err := NewMemoryObject(defaultObjectName, 1024, SHM_OPEN_CREATE, 0)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.NoError(t, obj.Destroy())
}

func TestDestroyMemoryObject(t *testing.T) {
	obj, err := NewMemoryObject(defaultObjectName, 1024, SHM_OPEN_CREATE, 0)
	assert.NoError(t, err)
	if assert.NotNil(t, obj) {
		if !assert.NoError(t, obj.Destroy()) {
			return
		}
		_, err = NewMemoryObject(defaultObjectName, 1024, SHM_OPEN_READ, 0)
		assert.Error(t, err)
	}
}

func TestDestroyMemoryObject2(t *testing.T) {
	_, err := NewMemoryObject(defaultObjectName, 1024, SHM_OPEN_CREATE, 0)
	if assert.NoError(t, err) {
		assert.NoError(t, DestroyMemoryObject(defaultObjectName))
	}
}

func TestCreateMemoryRegionExclusive(t *testing.T) {
	obj, err := NewMemoryObject(defaultObjectName, 1024, SHM_OPEN_CREATE, 0)
	if !assert.NoError(t, err) {
		return
	}
	_, err = NewMemoryObject(defaultObjectName, 1024, SHM_OPEN_CREATE_IF_NOT_EXISTS, 0)
	assert.Error(t, err)
	obj.Destroy()
}

func TestMemoryObjectSize(t *testing.T) {
	obj, err := NewMemoryObject(defaultObjectName, 1024, SHM_OPEN_CREATE, 0)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(1024), obj.Size())
		obj.Destroy()
	}
}

func TestMemoryObjectName(t *testing.T) {
	obj, err := NewMemoryObject(defaultObjectName, 1024, SHM_OPEN_CREATE, 0)
	if assert.NoError(t, err) {
		assert.Equal(t, defaultObjectName, obj.Name())
		obj.Destroy()
	}
}

func TestMemoryObjectCloseOnGc(t *testing.T) {
	object, err := NewMemoryObject(defaultObjectName, 1024, SHM_OPEN_CREATE, 0)
	if !assert.NoError(t, err) {
		return
	}
	defer DestroyMemoryObject(defaultObjectName)
	file := object.file
	object = nil
	// this is to assure, that the finalized was called and that the
	// corresponding file was closed. this test can theoretically fail, and
	// we use several attempts, as it is not guaranteed that the object is garbage-collected
	// after a call to GC()
	for i := 0; i < 5; i++ {
		runtime.GC()
		if int(-1) == int(file.Fd()) {
			return
		}
		time.Sleep(time.Millisecond * 20)
	}
	assert.Fail(t, "the memory object was not finalized during the gc cycle")
}
