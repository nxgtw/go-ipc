// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMemoryRegion(t *testing.T) {
	region, err := NewMemoryRegion("go-ipc-test", 1024*1024, SHM_CREATE|SHM_RDWR, 0)
	assert.NoError(t, err)
	if assert.NotNil(t, region) {
		defer region.Destroy()
	}
}
