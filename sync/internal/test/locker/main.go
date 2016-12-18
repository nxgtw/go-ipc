// Copyright 2015 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/test"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	ipc_sync "bitbucket.org/avd/go-ipc/sync"
)

var (
	objName  = flag.String("object", "", "synchronization object name")
	objType  = flag.String("type", "m", "synchronization object type - m | spin")
	jobs     = flag.Int("jobs", 1, "count of simultaneous jobs")
	readlock = flag.Bool("ro", false, "use read lock where possible")
)

type readLocker interface {
	RLocker() ipc_sync.IPCLocker
}

const usage = `  test program for synchronization primitives.
available commands:
  create
  destroy
  inc64 shm_name n 
    increments an int64 value at the beginning of the shm_name region n times
  test shm_name n {expected values byte array}
    performs n reads from shm_name and compares the results with the expected data
if jobs > 1, all goroutines will execute operations reads.
byte array should be passed as a continuous string of 2-symbol hex byte values like '01020A'
`

func create() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("destroy: must not provide any arguments")
	}
	if _, err := createLocker(*objType, *objName, os.O_CREATE|os.O_EXCL); err != nil {
		return err
	}
	return nil
}

func destroy() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("destroy: must not provide any arguments")
	}
	return destroyLocker(*objType, *objName)
}

func inc64() error {
	if flag.NArg() != 3 {
		return fmt.Errorf("test: must provide exactly two arguments")
	}
	memObject, err := shm.NewMemoryObject(flag.Arg(1), os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer memObject.Close()
	n, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		return err
	}
	region, err := mmf.NewMemoryRegion(memObject, mmf.MEM_READWRITE, 0, int(unsafe.Sizeof(int64(0))))
	if err != nil {
		return err
	}
	defer region.Close()
	locker, err := createLocker(*objType, *objName, 0)
	if err != nil {
		return err
	}
	data := region.Data()
	ptr := (*int64)(allocator.ByteSliceData(data))
	if err = performInc(ptr, locker, n); err == nil {
		fmt.Println(*ptr)
	}
	return err
}

func performInc(ptr *int64, locker sync.Locker, n int) error {
	return performParallel(func(int) error {
		for i := 0; i < n; i++ {
			locker.Lock()
			*ptr++
			locker.Unlock()
		}
		return nil
	})
}

func test() error {
	if flag.NArg() != 4 {
		return fmt.Errorf("test: must provide exactly three arguments")
	}
	memObject, err := shm.NewMemoryObject(flag.Arg(1), os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer memObject.Close()
	n, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		return err
	}
	expected, err := testutil.StringToBytes(flag.Arg(3))
	if err != nil {
		return err
	}
	region, err := mmf.NewMemoryRegion(memObject, mmf.MEM_READ_ONLY, 0, len(expected))
	if err != nil {
		return err
	}
	defer region.Close()
	locker, err := createLocker(*objType, *objName, 0)
	if err != nil {
		return err
	}
	if *readlock {
		if rLocker, ok := locker.(readLocker); ok {
			locker = rLocker.RLocker()
		} else {
			return fmt.Errorf("%q mutex type does not implement RLocker")
		}
	}
	return performTest(expected, region, locker, n)
}

func performTest(expected []byte, actual *mmf.MemoryRegion, locker sync.Locker, n int) error {
	return performParallel(func(id int) error {
		for i := 0; i < n; i++ {
			if err := testData(expected, actual.Data(), locker, i); err != nil {
				return err
			}
		}
		return nil
	})
}

// TODO(avd) - add code to cancel jobs?
func performParallel(f func(int) error) error {
	var result error
	ch := make(chan error, *jobs)
	for nJob := 0; nJob < *jobs; nJob++ {
		go func(id int) {
			ch <- f(nJob)
		}(nJob)
	}
	for nJob := 0; nJob < *jobs; nJob++ {
		err := <-ch
		if result == nil && err != nil { // save the first error
			result = err
		}
	}
	return result
}

func testData(expected, actual []byte, locker sync.Locker, id int) error {
	locker.Lock()
	defer func() {
		locker.Unlock()
	}()
	for i, expectedValue := range expected {
		actualValue := actual[i]
		if expectedValue != actualValue {
			return fmt.Errorf("invalid value at %d. expected '%d', got '%d'", i, expectedValue, actualValue)
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
