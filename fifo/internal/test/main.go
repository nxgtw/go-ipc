// Copyright 2015 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"bitbucket.org/avd/go-ipc/fifo"
	"bitbucket.org/avd/go-ipc/internal/test"
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
		return fmt.Errorf("create: must provide exactly one arguments")
	}
	mode := os.O_CREATE | fifo.O_NONBLOCK | os.O_RDONLY
	obj, err := fifo.New(*objName, mode, 0666)
	if err != nil {
		return err
	}
	obj.Close()
	return nil
}

func destroy() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("destroy: must not provide any arguments")
	}
	return fifo.Destroy(*objName)
}

func read() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("read: must provide exactly one arguments")
	}
	length, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	mode := os.O_CREATE | os.O_RDONLY
	if *nonBlock {
		mode |= fifo.O_NONBLOCK
	}
	obj, err := fifo.New(*objName, mode, 0666)
	if err != nil {
		return err
	}
	defer obj.Close()
	buffer := make([]byte, length)
	var n int
	if n, err = obj.Read(buffer); err == nil {
		if n == length {
			if length > 0 {
				fmt.Println(ipc_testing.BytesToString(buffer))
			}
		} else {
			err = fmt.Errorf("wanted %d bytes, but got %d", length, n)
		}
	}
	return err
}

func test() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("test: must provide exactly one arguments")
	}
	mode := os.O_CREATE | os.O_RDONLY
	if *nonBlock {
		mode |= fifo.O_NONBLOCK
	}
	var obj fifo.Fifo
	var err error
	completed := ipc_testing.WaitForFunc(func() {
		obj, err = fifo.New(*objName, mode, 0666)
	}, time.Second*2)
	if !completed {
		return fmt.Errorf("fifo.New took too long to finish")
	}
	if err != nil {
		return err
	}
	defer obj.Close()
	data, err := ipc_testing.StringToBytes(flag.Arg(1))
	if err != nil {
		return err
	}
	completed = ipc_testing.WaitForFunc(func() {
		buffer := make([]byte, len(data))
		if _, err = obj.Read(buffer); err == nil {
			for i, value := range data {
				if value != buffer[i] {
					err = fmt.Errorf("invalid value at %d. expected '%d', got '%d'", i, value, buffer[i])
					return
				}
			}
		}
	}, time.Second*2)
	if !completed {
		return fmt.Errorf("read took too long to finish")
	}
	return err
}

func write() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("test: must provide exactly one arguments")
	}
	mode := os.O_CREATE | os.O_WRONLY
	if *nonBlock {
		mode |= fifo.O_NONBLOCK
	}
	obj, err := fifo.New(*objName, mode, 0666)
	if err != nil {
		return err
	}
	defer obj.Close()
	data, err := ipc_testing.StringToBytes(flag.Arg(1))
	if err != nil {
		return err
	}
	var written int
	if written, err = obj.Write(data); err == nil && written != len(data) {
		err = fmt.Errorf("must write %d bytes, but wrote %d", len(data), written)
	}
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
