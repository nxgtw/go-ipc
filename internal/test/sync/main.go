// Copyright 2015 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"

	ipc "bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/test"
)

var (
	objName = flag.String("object", "", "synchronization object name")
	objType = flag.String("type", "", "synchronization object type - m | rwm")
	jobs    = flag.Int("jobs", 1, "count of simultaneous jobs")
)

const usage = `  test program for synchronization primitives.
available commands:
  create
  destroy
  inc64 shm_name n 
  test shm_name n {expected values byte array}
    performs n reads from shm_name and compares the results with the expected data
    if jobs > 1, all goroutines will execute n reads.
byte array should be passed as a continuous string of 2-symbol hex byte values like '01020A'
`

func createLocker(mode int, readonly bool) (locker sync.Locker, err error) {
	if *objType == "m" {
		err = fmt.Errorf("unimplemented")
	} else if *objType == "rwm" {
		if rwm, errRwm := ipc.NewRwMutex(*objName, mode, 0666); errRwm == nil {
			if readonly {
				locker = rwm.RLocker()
			} else {
				locker = rwm
			}
		} else {
			err = errRwm
		}
	} else {
		err = fmt.Errorf("unknown object type %q", *objType)
	}
	return
}

func create() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("create: must provide exactly one argument")
	}
	_, err := createLocker(ipc.O_CREATE_ONLY, false)
	return err
}

func destroy() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("destroy: must provide exactly one argument")
	}
	var err error
	if *objType == "m" {
		err = fmt.Errorf("unimplemented")
	} else if *objType == "rwm" {
		err = ipc.DestroyRwMutex(*objName)
	} else {
		return fmt.Errorf("unknown object type %q", *objType)
	}
	return err
}

func inc64() error {
	return nil
}

func test() error {
	if flag.NArg() != 4 {
		return fmt.Errorf("test: must provide exactly four arguments")
	}
	memObject, err := ipc.NewMemoryObject(flag.Arg(1), ipc.O_OPEN_ONLY|ipc.O_READ_ONLY, 0666)
	if err != nil {
		return err
	}
	defer memObject.Close()
	n, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		return err
	}
	expected, err := ipc_test.StringToBytes(flag.Arg(3))
	if err != nil {
		return err
	}
	region, err := ipc.NewMemoryRegion(memObject, ipc.SHM_READ_ONLY, 0, len(expected))
	defer region.Close()
	if err != nil {
		return err
	}
	locker, err := createLocker(ipc.O_OPEN_ONLY, true)
	if err != nil {
		return err
	}
	return performTest(expected, region, locker, n)
}

func performTest(expected []byte, actual *ipc.MemoryRegion, locker sync.Locker, n int) error {
	var result error
	ch := make(chan error, *jobs)
	for nJob := 0; nJob < *jobs; nJob++ {
		go func() {
			for i := 0; i < n; i++ {
				if err := testData(expected, actual.Data(), locker); err != nil {
					ch <- err
					return
				}
			}
			ch <- nil
		}()
	}
	for nJob := 0; nJob < *jobs; nJob++ {
		err := <-ch
		if result == nil && err != nil {
			result = err
		}
	}
	return result
}

func testData(expected, actual []byte, locker sync.Locker) error {
	locker.Lock()
	defer locker.Unlock()
	for i, value := range expected {
		if value != actual[i] {
			return fmt.Errorf("invalid value at %d. expected '%d', got '%d'", i, value, actual[i])
		}
	}
	return nil
}

func runCommand() error {
	command := flag.Arg(0)
	if *jobs <= 0 {
		return fmt.Errorf("invalid jobs number")
	}
	switch command {
	case "create":
		return create()
	case "destroy":
		return destroy()
	case "inc64":
		return inc64()
	case "test":
		return test()
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
