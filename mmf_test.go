// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"io"
	"io/ioutil"
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
	mr, err := NewMemoryRegion(file, MEM_READ_ONLY, 0, int(stat.Size()))
	a.NoError(err)
	a.NoError(mr.Close())
	mr, err = NewMemoryRegion(file, MEM_READ_ONLY, 0, 0)
	a.NoError(err)
	a.NoError(mr.Close())
	mr, err = NewMemoryRegion(file, MEM_READ_ONLY, 67746, int(stat.Size())-67746)
	a.NoError(err)
	a.NoError(mr.Close())
	mr, err = NewMemoryRegion(file, MEM_READ_ONLY, stat.Size()-1024, 1024)
	a.NoError(err)
	a.NoError(mr.Close())
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

func TestMmfFileCopy(t *testing.T) {
	a := assert.New(t)
	inFile, err := os.Open("internal/test/files/test.bin")
	if !assert.NoError(t, err) {
		return
	}
	defer inFile.Close()
	stat, err := inFile.Stat()
	if !a.NoError(err) {
		return
	}
	inRegion, err := NewMemoryRegion(inFile, MEM_READ_ONLY, 0, 0)
	if !a.NoError(err) {
		return
	}
	outFile, err := os.Create("internal/test/files/tmp.bin")
	if !a.NoError(err) {
		return
	}
	defer func() {
		outFile.Close()
		os.Remove("internal/test/files/tmp.bin")
	}()
	if !a.NoError(outFile.Truncate(stat.Size())) {
		return
	}
	outRegion, err := NewMemoryRegion(outFile, MEM_READWRITE, 0, 0)
	if !a.NoError(err) {
		return
	}
	rd := NewMemoryRegionReader(inRegion)
	wr := NewMemoryRegionWriter(outRegion)
	written, err := io.Copy(wr, rd)
	a.Equal(written, stat.Size())
	a.NoError(err)
	if !a.NoError(outRegion.Flush(false)) {
		return
	}
	expected, err := ioutil.ReadAll(inFile)
	if !a.NoError(err) {
		return
	}
	actual, err := ioutil.ReadAll(outFile)
	if !a.NoError(err) {
		return
	}
	a.Equal(expected, actual)
}
