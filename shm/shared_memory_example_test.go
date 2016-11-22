// Copyright 2016 Aleksandr Demakin. All rights reserved.

package shm

import (
	"os"

	"bitbucket.org/avd/go-ipc/mmf"
)

func ExampleMemoryObject() {
	// cleanup previous objects
	DestroyMemoryObject("obj")
	// create new object and resize it.
	obj, err := NewMemoryObject("obj", os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		panic("new")
	}
	if err := obj.Truncate(1024); err != nil {
		panic("truncate")
	}
	// create two regions for reading and writing.
	rwRegion, err := mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, 1024)
	if err != nil {
		panic("new region")
	}
	defer rwRegion.Close()
	roRegion, err := mmf.NewMemoryRegion(obj, mmf.MEM_READ_ONLY, 0, 1024)
	if err != nil {
		panic("new region")
	}
	defer roRegion.Close()
	// copy some data to the first region and read it via the second one.
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	copy(rwRegion.Data(), data)
	for i, b := range data {
		if b != roRegion.Data()[i] {
			panic("bad data")
		}
	}
}

func ExampleMemoryObject_readWriter() {
	// this example shows how to use memory obects with memory region readers/writers.

	// cleanup previous objects
	DestroyMemoryObject("obj")
	// create new object and resize it.
	obj, err := NewMemoryObject("obj", os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		panic("new")
	}
	if err := obj.Truncate(1024); err != nil {
		panic("truncate")
	}
	// create two regions for reading and writing.
	rwRegion, err := mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, 1024)
	if err != nil {
		panic("new region")
	}
	defer rwRegion.Close()
	roRegion, err := mmf.NewMemoryRegion(obj, mmf.MEM_READ_ONLY, 0, 1024)
	if err != nil {
		panic("new region")
	}
	defer roRegion.Close()
	// for each region we create a reader and a writer, which is a better solution, than
	// using region.Data() bytes directly.
	writer := mmf.NewMemoryRegionWriter(rwRegion)
	reader := mmf.NewMemoryRegionReader(roRegion)
	// write data at the specified offset
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	written, err := writer.WriteAt(data, 128)
	if err != nil || written != len(data) {
		panic("write")
	}
	// read data at the same offset via another region.
	actual := make([]byte, len(data))
	read, err := reader.ReadAt(actual, 128)
	if err != nil || read != len(data) {
		panic("read")
	}
	for i, b := range data {
		if b != actual[i] {
			panic("wrong data")
		}
	}
}
