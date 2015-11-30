// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMemoryRegion(t *testing.T) {
	region, err := NewMemoryRegion("go-ipc-test", 1024, SHM_OPEN_CREATE, 0)
	assert.NoError(t, err)
	assert.NotNil(t, region)
}

func TestDestroyMemoryRegion(t *testing.T) {
	region, err := NewMemoryRegion("go-ipc-test", 1024, SHM_OPEN_CREATE, 0)
	assert.NoError(t, err)
	if assert.NotNil(t, region) {
		if !assert.NoError(t, region.Destroy()) {
			return
		}
		_, err = NewMemoryRegion("go-ipc-test", 1024, SHM_OPEN_READ, 0)
		assert.Error(t, err)
	}
}

func TestCreateMemoryRegionExclusive(t *testing.T) {
	region, err := NewMemoryRegion("go-ipc-test", 1024, SHM_OPEN_CREATE, 0)
	assert.NoError(t, err)
	if assert.NotNil(t, region) {
		assert.NoError(t, region.Destroy())
	}
	region, err = NewMemoryRegion("go-ipc-test", 1024, SHM_OPEN_CREATE_IF_NOT_EXISTS, 0)
	assert.Error(t, err)
}

func TestMemoryRegionSize(t *testing.T) {
	region, err := NewMemoryRegion("go-ipc-test", 1024, SHM_OPEN_CREATE, 0)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(1024), region.Size())
	}
}
