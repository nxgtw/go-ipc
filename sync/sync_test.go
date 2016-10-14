// Copyright 2015 Aleksandr Demakin. All rights reserved.
// ignore this for a while, as linux rw mutexes don't work,
// and windows mutexes are not ready yes.

package sync

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
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
	lockerProgFiles []string
	condProgFiles   []string
	eventProgFiles  []string
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
	return files
}

func init() {
	lockerProgFiles = locate(lockerProgPath)
	condProgFiles = locate(condProgPath)
	eventProgFiles = locate(eventProgPath)
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
	return append(lockerProgFiles, "-object="+name, "-type="+t, "create")
}

func argsForSyncDestroyCommand(name string) []string {
	return append(lockerProgFiles, "-object="+name, "destroy")
}

func argsForSyncInc64Command(name, t string, jobs int, shmName string, n int, logFile string) []string {
	return append(lockerProgFiles,
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
	return append(lockerProgFiles,
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
	return append(condProgFiles,
		"signal",
		name,
	)
}

func argsForCondBroadcastCommand(name string) []string {
	return append(condProgFiles,
		"broadcast",
		name,
	)
}

func argsForCondWaitCommand(condName, lockerName string) []string {
	return append(condProgFiles,
		"wait",
		condName,
		lockerName,
	)
}

// Event test program

func argsForEventSetCommand(name string) []string {
	return append(eventProgFiles,
		"set",
		name,
	)
}

func argsForEventWaitCommand(name string, timeoutMS int) []string {
	return append(eventProgFiles,
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
