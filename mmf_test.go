// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMmfOpen(t *testing.T) {
	a := assert.New(t)
	file, err := os.Open("internal/test/files/test.bin")
	if !assert.NoError(t, err) {
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if !a.NoError(err) {
		return
	}
	_, err = NewMemoryRegion(file, MEM_READ_ONLY, 0, int(stat.Size()))
	a.NoError(err)
	_, err = NewMemoryRegion(file, MEM_READ_ONLY, 0, 0)
	a.NoError(err)
	_, err = NewMemoryRegion(file, MEM_READ_ONLY, 67746, int(stat.Size())-67746)
	a.NoError(err)
	_, err = NewMemoryRegion(file, MEM_READ_ONLY, stat.Size()-1024, 1024)
	a.NoError(err)
	_, err = NewMemoryRegion(file, MEM_READ_ONLY, stat.Size()-1024, 1025)
	a.Error(err)
}

func TestMmfOpenReadonly(t *testing.T) {
	const (
		offset = 67746
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
