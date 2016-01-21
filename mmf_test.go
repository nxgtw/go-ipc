// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMmfOpenReadonly(t *testing.T) {
	const (
		offset = 67845
	)
	file, err := os.Open("internal/test/files/test.bin")
	if !assert.NoError(t, err) {
		return
	}
	defer file.Close()
	region, err := NewMemoryRegion(file, MEM_READ_ONLY, offset, 1024)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 1024, region.Size())
	for i := 0; i < 1024; i++ {
		if !assert.Equal(t, byte(i+offset), region.Data()[i]) {
			break
		}
	}
	region.Close()
}
