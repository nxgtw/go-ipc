// Copyright 2015 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	ipc "bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/test"
)

var (
	objName  = flag.String("object", "", "shared memory object name")
	nonBlock = flag.Bool("nonblock", false, "set O_NONBLOCK flag")
)

const usage = `  test program for fifo.
available commands:
  create
    this operation never blocks
  destroy
  read len
  test {expected values byte array}
  write {values byte array}
byte array should be passed as a continuous string of 2-symbol hex byte values like '01020A'
`

func create() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("create: must provide exactly one argument")
	}
	mode := ipc.O_READWRITE
	if fifo, err := ipc.NewFifo(*objName, mode, 0666); err != nil {
		return err
	} else {
		fifo.Close()
	}
	return nil
}

func destroy() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("destroy: must not provide any arguments")
	}
	return ipc.DestroyFifo(*objName)
}

func read() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("read: must provide exactly two arguments")
	}
	length, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	mode := ipc.O_READ_ONLY
	if *nonBlock {
		mode |= ipc.O_FIFO_NONBLOCK
	}
	fifo, err := ipc.NewFifo(*objName, mode, 0666)
	if err != nil {
		return err
	}
	defer fifo.Close()
	buffer := make([]byte, length)
	var n int
	if n, err = fifo.Read(buffer); err == nil {
		if n == length {
			if length > 0 {
				fmt.Println(ipc_test.BytesToString(buffer))
			}
		} else {
			err = fmt.Errorf("wanted %d bytes, but got %d", length, n)
		}
	}
	return err
}

func test() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("test: must provide exactly two arguments")
	}
	mode := ipc.O_READ_ONLY
	if *nonBlock {
		mode |= ipc.O_FIFO_NONBLOCK
	}
	fifo, err := ipc.NewFifo(*objName, mode, 0666)
	if err != nil {
		return err
	}
	defer fifo.Close()
	data, err := ipc_test.StringToBytes(flag.Arg(1))
	if err != nil {
		return err
	}
	buffer := make([]byte, len(data))
	if _, err = fifo.Read(buffer); err == nil {
		for i, value := range data {
			if value != buffer[i] {
				return fmt.Errorf("invalid value at %d. expected '%d', got '%d'", i, value, buffer[i])
			}
		}
	}
	return err
}

func write() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("test: must provide exactly two arguments")
	}
	mode := ipc.O_WRITE_ONLY
	if *nonBlock {
		mode |= ipc.O_FIFO_NONBLOCK
	}
	fifo, err := ipc.NewFifo(*objName, mode, 0666)
	if err != nil {
		return err
	}
	defer fifo.Close()
	data, err := ipc_test.StringToBytes(flag.Arg(1))
	if err != nil {
		return err
	}
	_, err = fifo.Write(data)
	return err
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
