// Copyright 2016 Aleksandr Demakin. All rights reserved.

package shm

import (
	"testing"

	ipc_test "bitbucket.org/avd/go-ipc/internal/test"
	"bitbucket.org/avd/go-ipc/mmf"

	"github.com/stretchr/testify/assert"
)

func createWindowsMemoryRegionSimple(regionMode int, size int64, offset int64) (*mmf.MemoryRegion, error) {
	object := NewWindowsNativeMemoryObject(defaultObjectName)
	region, err := mmf.NewMemoryRegion(object, regionMode, offset, int(size))
	if err != nil {
		return nil, err
	}
	return region, nil
}

func TestWriteWindowsMemoryRegionSameProcess(t *testing.T) {
	region, err := createWindowsMemoryRegionSimple(mmf.MEM_READWRITE, int64(len(shmTestData)), 0)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		assert.NoError(t, region.Close())
	}()
	copied := copy(region.Data(), shmTestData)
	assert.Equal(t, copied, len(shmTestData))
	assert.NoError(t, region.Flush(false))
	region2, err := createWindowsMemoryRegionSimple(mmf.MEM_READ_ONLY, int64(len(shmTestData)), 0)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, region.Data(), region2.Data())
	assert.NoError(t, region2.Close())
}

func TestWriteWindowsMemoryAnotherProcess(t *testing.T) {
	a := assert.New(t)
	region, err := createWindowsMemoryRegionSimple(mmf.MEM_READWRITE, int64(len(shmTestData)), 128)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(region.Close())
	}()
	copy(region.Data(), shmTestData)
	a.NoError(region.Flush(false))
	result := ipc_test.RunTestApp(argsForShmTestCommand(defaultObjectName, "wnm", 128, shmTestData), nil)
	if !a.NoError(result.Err) {
		t.Log(result.Output)
	}
}

func TestReadWindowsMemoryAnotherProcess(t *testing.T) {
	a := assert.New(t)
	object := NewWindowsNativeMemoryObject(defaultObjectName)
	region, err := mmf.NewMemoryRegion(object, mmf.MEM_READWRITE, 0, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	defer region.Close()
	result := ipc_test.RunTestApp(argsForShmWriteCommand(defaultObjectName, "wnm", 0, shmTestData), nil)
	if !a.NoError(result.Err) {
		t.Log(result.Output)
		return
	}
	a.Equal(shmTestData, region.Data())
	if !a.NoError(result.Err) {
		t.Log(result.Output)
	}
}

func TestMemoryRegionCreation(t *testing.T) {
	a := assert.New(t)
	region, err := createWindowsMemoryRegionSimple(mmf.MEM_READ_ONLY, 4096, 0)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(region.Close())
	}()
	region2, err := createWindowsMemoryRegionSimple(mmf.MEM_READWRITE, 4096, 0)
	if !assert.NoError(t, err) {
		return
	}
	a.NoError(region2.Close())
}
