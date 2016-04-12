// Copyright 2016 Aleksandr Demakin. All rights reserved.

package shm

import (
	"testing"

	"bitbucket.org/avd/go-ipc"

	"github.com/stretchr/testify/assert"
)

func createWindowsMemoryRegionSimple(regionMode int, size int64, offset int64) (*ipc.MemoryRegion, error) {
	object := NewWindowsNativeMemoryObject(defaultObjectName)
	defer func() {
		err := object.Close()
		if err != nil {
			panic(err.Error())
		}
	}()
	region, err := ipc.NewMemoryRegion(object, regionMode, offset, int(size))
	if err != nil {
		return nil, err
	}
	return region, nil
}

func TestWindowsMemoryObjectName(t *testing.T) {
	a := assert.New(t)
	obj := NewWindowsNativeMemoryObject(defaultObjectName)
	a.Equal(defaultObjectName, obj.Name())
	a.NoError(obj.Destroy())
}

func TestWriteWindowsMemoryRegionSameProcess(t *testing.T) {
	region, err := createWindowsMemoryRegionSimple(ipc.MEM_READWRITE, int64(len(shmTestData)), 0)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		assert.NoError(t, region.Close())
	}()
	copy(region.Data(), shmTestData)
	assert.NoError(t, region.Flush(false))
	region2, err := createWindowsMemoryRegionSimple(ipc.MEM_READ_ONLY, int64(len(shmTestData)), 0)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, shmTestData, region2.Data())
	assert.NoError(t, region2.Close())
}

/*
func TestWriteWindowsMemoryAnotherProcess(t *testing.T) {
	region, err := createMemoryRegionSimple(ipc.O_OPEN_OR_CREATE|ipc.O_READWRITE, ipc.MEM_READWRITE, int64(len(shmTestData)), 128)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		assert.NoError(t, region.Close())
		assert.NoError(t, DestroyMemoryObject(defaultObjectName))
	}()
	copy(region.Data(), shmTestData)
	assert.NoError(t, region.Flush(false))
	result := ipc_test.RunTestApp(argsForShmTestCommand(defaultObjectName, 128, shmTestData), nil)
	assert.NoError(t, result.Err)
}

func TestReadWindowsMemoryAnotherProcess(t *testing.T) {
	object := NewWindowsNativeMemoryObject(defaultObjectName)
	result := ipc_test.RunTestApp(argsForShmWriteCommand(defaultObjectName, 0, shmTestData), nil)
	if !assert.NoError(t, result.Err) {
		t.Log(result.Output)
		return
	}
	region, err := ipc.NewMemoryRegion(object, ipc.MEM_READ_ONLY, 0, len(shmTestData))
	if !assert.NoError(t, err) {
		return
	}
	defer region.Close()
	assert.Equal(t, shmTestData, region.Data())
}*/
