// Copyright 2015 Aleksandr Demakin. All rights reserved.

package mmf

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testFolder = "./testdata/"
	testFile   = testFolder + "test.bin"
)

func TestMmfOpen(t *testing.T) {
	a := assert.New(t)
	file, err := os.Open(testFile)
	if !assert.NoError(t, err) {
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if !a.NoError(err) {
		return
	}
	mr, err := NewMemoryRegion(file, MEM_READ_ONLY, 0, int(stat.Size()))
	if !a.NoError(err) {
		return
	}
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
	file, err := os.Open(testFile)
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
	inFile, err := os.Open(testFile)
	if !assert.NoError(t, err) {
		return
	}
	defer inFile.Close()
	stat, err := inFile.Stat()
	if !a.NoError(err) {
		return
	}
	outFile, err := os.Create(os.TempDir() + "/tmp.bin")
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(outFile.Close())
		a.NoError(os.Remove(os.TempDir() + "/tmp.bin"))
	}()
	if !a.NoError(outFile.Truncate(stat.Size())) {
		return
	}
	inRegion, err := NewMemoryRegion(inFile, MEM_READ_ONLY, 0, 0)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(inRegion.Close())
	}()
	outRegion, err := NewMemoryRegion(outFile, MEM_READWRITE, 0, 0)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(outRegion.Close())
	}()
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

func ExampleMemoryRegion() {
	// this example shows how to copy a file using mmf.

	// open source file for reading.
	inFile, err := os.Open("in.dat")
	if err != nil {
		panic("open file")
	}
	stat, err := inFile.Stat()
	if err != nil {
		panic("stat")
	}
	// open destination file for writing.
	outFile, err := os.Create("out.dat")
	if err != nil {
		panic("create file")
	}
	// then, mmap contents of both files.
	inRegion, err := NewMemoryRegion(inFile, MEM_READ_ONLY, 0, 0)
	if err != nil {
		panic("in region")
	}
	defer inRegion.Close()
	outRegion, err := NewMemoryRegion(outFile, MEM_READWRITE, 0, 0)
	if err != nil {
		panic("out region")
	}
	defer outRegion.Close()

	// copy file contents.
	rd := NewMemoryRegionReader(inRegion)
	wr := NewMemoryRegionWriter(outRegion)
	written, err := io.Copy(wr, rd)

	if err != nil || written != stat.Size() {
		panic("copy")
	}

	// ensure the data has been written.
	if err := outRegion.Flush(false); err != nil {
		panic("flush")
	}
}
