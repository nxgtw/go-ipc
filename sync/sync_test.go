// Copyright 2015 Aleksandr Demakin. All rights reserved.
// ignore this for a while, as linux rw mutexes don't work,
// and windows mutexes are not ready yes.

package sync

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"strconv"

	testutil "bitbucket.org/avd/go-ipc/internal/test"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
)

const (
	lockerProgPath = "./internal/test/locker/"
	condProgPath   = "./internal/test/cond/"
	eventProgPath  = "./internal/test/event/"
	testMemObj     = "go-ipc.sync-test.region"
)

var (
	lockerProgArgs   []string
	condProgArgs     []string
	eventProgArgs    []string
	defaultMutexType = "m"
)

func locate(path string) []string {
	files, err := testutil.LocatePackageFiles(path)
	if err != nil {
		panic(err)
	}
	if len(files) == 0 {
		panic("no locker test files")
	}
	for i, name := range files {
		files[i] = path + name
	}
	if defaultMutexType == "msysv" {
		files = append([]string{`-tags="sysv_mutex_linux"`}, files...)
	}
	return files
}

func detectMutexType() {
	DestroyMutex(testLockerName)
	m, err := NewMutex(testLockerName, os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	t := reflect.ValueOf(m)
	if t.Elem().Type().Name() == "SemaMutex" {
		defaultMutexType = "msysv"
	}
	m.Close()
	DestroyMutex(testLockerName)
}

func init() {
	detectMutexType()
	lockerProgArgs = locate(lockerProgPath)
	condProgArgs = locate(condProgPath)
	eventProgArgs = locate(eventProgPath)
}

func createMemoryRegionSimple(objMode, regionMode int, size int64, offset int64) (*mmf.MemoryRegion, error) {
	object, _, err := shm.NewMemoryObjectSize(testMemObj, objMode, 0666, size)
	if err != nil {
		return nil, err
	}
	defer func() {
		errClose := object.Close()
		if errClose != nil {
			panic(errClose.Error())
		}
	}()
	region, err := mmf.NewMemoryRegion(object, regionMode, offset, int(size))
	if err != nil {
		return nil, err
	}
	return region, nil
}

// Locker test program

func argsForSyncCreateCommand(name, t string) []string {
	return append(lockerProgArgs, "-object="+name, "-type="+t, "create")
}

func argsForSyncDestroyCommand(name string) []string {
	return append(lockerProgArgs, "-object="+name, "destroy")
}

func argsForSyncInc64Command(name, t string, jobs int, shmName string, n int, logFile string) []string {
	return append(lockerProgArgs,
		"-object="+name,
		"-type="+t,
		"-jobs="+strconv.Itoa(jobs),
		"-log="+logFile,
		"inc64",
		shmName,
		strconv.Itoa(n),
	)
}

func argsForSyncTestCommand(name, t string, jobs int, shmName string, n int, data []byte, log string) []string {
	return append(lockerProgArgs,
		"-object="+name,
		"-type="+t,
		"-jobs="+strconv.Itoa(jobs),
		"-log="+log,
		"test",
		shmName,
		strconv.Itoa(n),
		testutil.BytesToString(data),
	)
}

// Cond test program

func argsForCondSignalCommand(name string) []string {
	return append(
		condProgArgs,
		"signal",
		name,
	)
}

func argsForCondBroadcastCommand(name string) []string {
	return append(condProgArgs,
		"broadcast",
		name,
	)
}

func argsForCondWaitCommand(condName, lockerName string) []string {
	return append(condProgArgs,
		"wait",
		condName,
		lockerName,
	)
}

// Event test program

func argsForEventSetCommand(name string) []string {
	return append(eventProgArgs,
		"set",
		name,
	)
}

func argsForEventWaitCommand(name string, timeoutMS int) []string {
	return append(eventProgArgs,
		"-timeout="+strconv.Itoa(timeoutMS),
		"wait",
		name,
	)
}

func startPprof() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}
