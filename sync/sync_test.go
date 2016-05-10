// Copyright 2015 Aleksandr Demakin. All rights reserved.
// ignore this for a while, as linux rw mutexes don't work,
// and windows mutexes are not ready yes.

package sync

import (
	"strconv"

	ipc "bitbucket.org/avd/go-ipc"
	ipc_test "bitbucket.org/avd/go-ipc/internal/test"
	"bitbucket.org/avd/go-ipc/shm"
)

const (
	testProgPath = "./internal/test/sync/"
	testMemObj   = "go-ipc.sync-test.region"
)

var testProgFiles []string

func init() {
	var err error
	testProgFiles, err = ipc_test.LocatePackageFiles(testProgPath)
	if err != nil {
		panic(err)
	}
	if len(testProgFiles) == 0 {
		panic("no files to test mq")
	}
	for i, name := range testProgFiles {
		testProgFiles[i] = testProgPath + name
	}
}

func createMemoryRegionSimple(objMode, regionMode int, size int64, offset int64) (*ipc.MemoryRegion, error) {
	object, err := shm.NewMemoryObject(testMemObj, objMode, 0666)
	if err != nil {
		return nil, err
	}
	defer func() {
		errClose := object.Close()
		if errClose != nil {
			panic(errClose.Error())
		}
	}()
	if objMode&ipc.O_OPEN_ONLY == 0 {
		if err = object.Truncate(size + offset); err != nil {
			return nil, err
		}
	}
	region, err := ipc.NewMemoryRegion(object, regionMode, offset, int(size))
	if err != nil {
		return nil, err
	}
	return region, nil
}

// Sync test program

func argsForSyncCreateCommand(name, t string) []string {
	return append(testProgFiles, "-object="+name, "-type="+t, "create")
}

func argsForSyncDestroyCommand(name string) []string {
	return append(testProgFiles, "-object="+name, "destroy")
}

func argsForSyncInc64Command(name, t string, jobs int, shmName string, n int, logFile string) []string {
	return append(testProgFiles,
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
	return append(testProgFiles,
		"-object="+name,
		"-type="+t,
		"-jobs="+strconv.Itoa(jobs),
		"-log="+log,
		"test",
		shmName,
		strconv.Itoa(n),
		ipc_test.BytesToString(data),
	)
}
