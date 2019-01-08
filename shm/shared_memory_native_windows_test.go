// Copyright 2016 Aleksandr Demakin. All rights reserved.

package shm

import (
	"os"
	"testing"

	"github.com/nxgtw/go-ipc/internal/test"
	"bitbucket.org/avd/go-ipc/mmf"

	"github.com/stretchr/testify/assert"
)

func TestCreateWindowsMemoryObject(t *testing.T) {
	a := assert.New(t)
	obj, err := NewWindowsNativeMemoryObject(defaultObjectName, os.O_RDWR, 1024)
	if !a.Error(err) {
		obj.Close()
		return
	}
	closer := func(obj *WindowsNativeMemoryObject) {
		a.NoError(obj.Close())
	}
	obj, err = NewWindowsNativeMemoryObject(defaultObjectName, os.O_CREATE|os.O_EXCL|os.O_RDWR, 1024)
	if !a.NoError(err) {
		return
	}
	defer closer(obj)
	obj, err = NewWindowsNativeMemoryObject(defaultObjectName, os.O_CREATE|os.O_EXCL|os.O_RDWR, 1024)
	if !a.Error(err) {
		obj.Close()
		return
	}
	obj, err = NewWindowsNativeMemoryObject(defaultObjectName, os.O_RDWR, 1024)
	if !a.NoError(err) {
		return
	}
	defer closer(obj)
	obj, err = NewWindowsNativeMemoryObject(defaultObjectName, os.O_CREATE|os.O_RDWR, 1024)
	if !a.NoError(err) {
		return
	}
	defer closer(obj)
}

func TestWriteWindowsMemoryRegionSameProcess(t *testing.T) {
	a := assert.New(t)
	object, err := NewWindowsNativeMemoryObject(defaultObjectName, os.O_CREATE|os.O_RDWR, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(object.Close())
	}()
	region, err := mmf.NewMemoryRegion(object, mmf.MEM_READWRITE, 0, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(region.Close())
	}()
	copied := copy(region.Data(), shmTestData)
	assert.Equal(t, copied, len(shmTestData))
	assert.NoError(t, region.Flush(false))
	region2, err := mmf.NewMemoryRegion(object, mmf.MEM_READ_ONLY, 0, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, region.Data(), region2.Data())
	assert.NoError(t, region2.Close())
}

func TestWriteWindowsMemoryAnotherProcess(t *testing.T) {
	a := assert.New(t)
	object, err := NewWindowsNativeMemoryObject(defaultObjectName, os.O_CREATE|os.O_RDWR, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(object.Close())
	}()
	region, err := mmf.NewMemoryRegion(object, mmf.MEM_READWRITE, 128, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(region.Close())
	}()
	copy(region.Data(), shmTestData)
	a.NoError(region.Flush(false))
	result := testutil.RunTestApp(argsForShmTestCommand(defaultObjectName, "wnm", 128, shmTestData), nil)
	if !a.NoError(result.Err) {
		t.Log(result.Output)
	}
}

func TestReadWindowsMemoryAnotherProcess(t *testing.T) {
	a := assert.New(t)
	object, err := NewWindowsNativeMemoryObject(defaultObjectName, os.O_CREATE|os.O_RDWR, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(object.Close())
	}()
	region, err := mmf.NewMemoryRegion(object, mmf.MEM_READWRITE, 0, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(region.Close())
	}()
	result := testutil.RunTestApp(argsForShmWriteCommand(defaultObjectName, "wnm", 0, shmTestData), nil)
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
	object, err := NewWindowsNativeMemoryObject(defaultObjectName, os.O_CREATE|os.O_RDWR, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(object.Close())
	}()
	region, err := mmf.NewMemoryRegion(object, mmf.MEM_READWRITE, 0, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(region.Close())
	}()
	region2, err := mmf.NewMemoryRegion(object, mmf.MEM_READWRITE, 0, len(shmTestData))
	if !a.NoError(err) {
		return
	}
	a.NoError(region2.Close())
}
