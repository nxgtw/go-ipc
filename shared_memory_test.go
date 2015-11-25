// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMemoryRegion(t *testing.T) {
	region, err := NewMemoryRegion("go-ipc-test", 1024, SHM_OPEN_CREATE, 0)
	assert.NoError(t, err)
	if assert.NotNil(t, region) {
		assert.NoError(t, region.Destroy())
	}
}

/*
func TestCreateMemoryRegionEcxl(t *testing.T) {
	region, err := NewMemoryRegion("go-ipc-test", 0, SHM_OPEN_CREATE|SHM_OPEN_RDWR, 0)
	assert.NoError(t, err)
	if assert.NotNil(t, region) {
		defer region.Destroy()
	}
	region2, err2 := NewMemoryRegion("go-ipc-test", 0, SHM_OPEN_CREATE_IF_NOT_EXISTS|SHM_OPEN_RDONLY, 0)
	assert.Error(t, err2)
	assert.Nil(t, region2)
}
*/
