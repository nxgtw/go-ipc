// Copyright 2015 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"bitbucket.org/avd/go-ipc/internal/test"
	"bitbucket.org/avd/go-ipc/mmf"
)

var (
	objName = flag.String("object", "", "shared memory object name")
	objType = flag.String("type", "", "object type (empty for default | 'wnm' for windows native)")
)

const usage = `  test program for shared memory.
available commands:
  create {size}
  destroy
  read offset len
  test offset {expected values byte array}
  write offset {values byte array}
byte array should be passed as a continuous string of 2-symbol hex byte values like '01020A'
`

func create() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("create: must provide exactly one argument")
	}
	size, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	obj, err := newShmObject(*objName, os.O_CREATE|os.O_RDWR, 0666, *objType, size)
	if err != nil {
		return err
	}
	return obj.Close()
}

func destroy() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("destroy: must not provide any arguments")
	}
	return destroyShmObject(*objName, *objType)
}

func read() error {
	if flag.NArg() != 3 {
		return fmt.Errorf("read: must provide exactly two arguments")
	}
	offset, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	length, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		return err
	}
	object, err := newShmObject(*objName, os.O_RDONLY, 0666, *objType, length)
	if err != nil {
		return err
	}
	defer object.Close()
	region, err := mmf.NewMemoryRegion(object, mmf.MEM_READ_ONLY, int64(offset), length)
	if err != nil {
		return err
	}
	defer region.Close()
	if len(region.Data()) > 0 {
		fmt.Println(testutil.BytesToString(region.Data()))
	}
	return nil
}

func test() error {
	if flag.NArg() != 3 {
		return fmt.Errorf("test: must provide exactly two arguments")
	}
	offset, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	data, err := testutil.StringToBytes(flag.Arg(2))
	if err != nil {
		return err
	}
	object, err := newShmObject(*objName, os.O_RDONLY, 0666, *objType, len(data))
	if err != nil {
		return err
	}
	defer object.Close()
	region, err := mmf.NewMemoryRegion(object, mmf.MEM_READ_ONLY, int64(offset), len(data))
	defer region.Close()
	if err != nil {
		return err
	}
	for i, value := range region.Data() {
		if value != data[i] {
			return fmt.Errorf("invalid value at %d. expected '%d', got '%d'", i, data[i], value)
		}
	}
	return nil
}

func write() error {
	if flag.NArg() != 3 {
		return fmt.Errorf("test: must provide exactly two arguments")
	}
	offset, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	data, err := testutil.StringToBytes(flag.Arg(2))
	if err != nil {
		return err
	}
	object, err := newShmObject(*objName, os.O_CREATE|os.O_RDWR, 0666, *objType, len(data))
	if err != nil {
		return err
	}
	defer object.Close()
	region, err := mmf.NewMemoryRegion(object, mmf.MEM_READWRITE, int64(offset), len(data))
	if err != nil {
		return err
	}
	defer func() {
		region.Flush(true)
		region.Close()
	}()
	rData := region.Data()
	for i, value := range data {
		rData[i] = value
	}
	return nil
}

func runCommand() error {
	command := flag.Arg(0)
	switch command {
	case "create":
		return create()
	case "destroy":
		return destroy()
	case "read":
		return read()
	case "test":
		return test()
	case "write":
		return write()
	default:
		return fmt.Errorf("unknown command")
	}
}

func main() {
	flag.Parse()
	if len(*objName) == 0 || flag.NArg() == 0 {
		fmt.Print(usage)
		flag.Usage()
		os.Exit(1)
	}
	if err := runCommand(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
