// Copyright 2015 Aleksandr Demakin. All rights reserved.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	ipc "bitbucket.org/avd/go-ipc"
)

var (
	objName = flag.String("object", "", "shared memory object name")
)

const usage = `test program for shared memory.
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
	if obj, err := ipc.NewMemoryObject(*objName, ipc.SHM_CREATE, 0666); err != nil {
		return err
	} else {
		return obj.Truncate(int64(size))
	}
}

func destroy() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("destroy: must not provide any arguments")
	}
	return ipc.DestroyMemoryObject(*objName)
}

func read() error {
	if flag.NArg() != 3 {
		return fmt.Errorf("read: must provide exactly three arguments")
	}
	offset, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	length, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		return err
	}
	object, err := ipc.NewMemoryObject(*objName, ipc.SHM_READ, 0666)
	if err != nil {
		return err
	}
	region, err := ipc.NewMemoryRegion(object, ipc.SHM_READ, int64(offset), length)
	if err != nil {
		return err
	}
	if len(region.Data()) > 0 {
		for _, value := range region.Data() {
			if value < 16 {
				fmt.Print("0")
			}
			fmt.Printf("%X", value)
		}
		fmt.Println()
	}
	return nil
}

func test() error {
	if flag.NArg() < 3 {
		return fmt.Errorf("test: must provide exactly three arguments")
	}
	offset, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	object, err := ipc.NewMemoryObject(*objName, ipc.SHM_READ, 0666)
	if err != nil {
		return err
	}
	data, err := parseBytes(flag.Arg(2))
	if err != nil {
		return err
	}
	region, err := ipc.NewMemoryRegion(object, ipc.SHM_READ, int64(offset), len(data))
	if err != nil {
		return err
	}
	for i, value := range region.Data() {
		if value != data[i] {
			return fmt.Errorf("invalid value at %d. expected '%d', got '%d'", i, value, data[i])
		}
	}
	return nil
}

func write() error {
	if flag.NArg() < 3 {
		return fmt.Errorf("test: must provide exactly three arguments")
	}
	offset, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	object, err := ipc.NewMemoryObject(*objName, ipc.SHM_RW, 0666)
	if err != nil {
		return err
	}
	data, err := parseBytes(flag.Arg(2))
	if err != nil {
		return err
	}
	region, err := ipc.NewMemoryRegion(object, ipc.SHM_RW, int64(offset), len(data))
	if err != nil {
		return err
	}
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

func parseBytes(input string) ([]byte, error) {
	if len(input)%2 != 0 {
		return nil, fmt.Errorf("invalid byte array len")
	}
	var err error
	var b byte
	buff := bytes.NewBuffer(nil)
	for err == nil {
		if len(input) < 2 {
			err = io.EOF
		} else {
			if _, err = fmt.Sscanf(input[:2], "%X", &b); err == nil {
				buff.WriteByte(b)
				if len(input) >= 2 {
					input = input[2:]
				}
			}
		}
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buff.Bytes(), nil
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
