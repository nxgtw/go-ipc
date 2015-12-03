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
	obj, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.NoError(t, obj.Destroy())
}

func TestOpenMemoryRegionReadonly(t *testing.T) {
	obj, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer obj.Destroy()
	defer obj.Close()
	obj2, err := NewMemoryObject(defaultObjectName, O_OPEN_ONLY|O_READ_ONLY, 0)
	if !assert.NoError(t, err) {
		return
	}
	defer obj2.Close()
}

func TestDestroyMemoryObject(t *testing.T) {
	obj, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	assert.NoError(t, err)
	if assert.NotNil(t, obj) {
		if !assert.NoError(t, obj.Destroy()) {
			return
		}
		_, err = NewMemoryObject(defaultObjectName, O_OPEN_ONLY|O_READ_ONLY, 0666)
		assert.Error(t, err)
	}
}

func TestDestroyMemoryObject2(t *testing.T) {
	_, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if assert.NoError(t, err) {
		assert.NoError(t, DestroyMemoryObject(defaultObjectName))
	}
}

func TestCreateMemoryRegionExclusive(t *testing.T) {
	obj, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	_, err = NewMemoryObject(defaultObjectName, O_CREATE_ONLY|O_READWRITE, 0666)
	assert.Error(t, err)
	obj.Destroy()
}

func TestMemoryObjectSize(t *testing.T) {
	obj, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if assert.NoError(t, err) {
		if assert.NoError(t, obj.Truncate(1024)) {
			assert.Equal(t, int64(1024), obj.Size())
			obj.Destroy()
		}
	}
}

func TestMemoryObjectName(t *testing.T) {
	obj, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if assert.NoError(t, err) {
		assert.Equal(t, defaultObjectName, obj.Name())
		obj.Destroy()
	}
}

func TestMemoryObjectCloseOnGc(t *testing.T) {
	object, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
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

func TestWriteMemoryRegionSameProcess(t *testing.T) {
	testdata := []byte{1, 2, 3, 4, 128, 255}
	object, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer object.Destroy()
	if !assert.NoError(t, object.Truncate(1024)) {
		return
	}
	region, err := NewMemoryRegion(object, SHM_READWRITE, 128, len(testdata))
	if !assert.NoError(t, err) {
		return
	}
	defer region.Close()
	copy(region.Data(), testdata)
	assert.NoError(t, region.Flush(false))
	region2, err := NewMemoryRegion(object, SHM_READ_ONLY, 128, len(testdata))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, testdata, region2.Data())
}

func TestWriteMemoryAnotherProcess(t *testing.T) {
	testdata := []byte{1, 2, 3, 4, 128, 255}
	object, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer object.Destroy()
	if !assert.NoError(t, object.Truncate(1024)) {
		return
	}
	region, err := NewMemoryRegion(object, SHM_READWRITE, 128, len(testdata))
	if !assert.NoError(t, err) {
		return
	}
	defer region.Close()
	copy(region.Data(), testdata)
	assert.NoError(t, region.Flush(false))
	output, err := runTestShmProg(argsForShmTestCommand(defaultObjectName, 128, testdata))
	assert.NoError(t, err, output)
}

func TestReadMemoryAnotherProcess(t *testing.T) {
	testdata := []byte{1, 2, 3, 4, 128, 255}
	object, err := NewMemoryObject(defaultObjectName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer object.Destroy()
	if !assert.NoError(t, object.Truncate(1024)) {
		return
	}
	_, err = runTestShmProg(argsForShmWriteCommand(defaultObjectName, 0, testdata))
	if !assert.NoError(t, err) {
		return
	}
	region, err := NewMemoryRegion(object, SHM_READ_ONLY, 0, len(testdata))
	if !assert.NoError(t, err) {
		return
	}
	defer region.Close()
	assert.Equal(t, testdata, region.Data())
}
