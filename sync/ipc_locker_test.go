// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	ipc "bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/test"
	"bitbucket.org/avd/go-ipc/shm"

	"github.com/stretchr/testify/assert"
)

const (
	testLockerName = "go-ipc.locker"
)

type lockerCtor func(name string, mode int, perm os.FileMode) (IPCLocker, error)
type lockerDtor func(name string) error

func testLockerOpenMode(t *testing.T, ctor lockerCtor, dtor lockerDtor) {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return
		}
	}
	lk, err := ctor(testLockerName, ipc.O_READWRITE, 0666)
	a.Error(err)
	lk, err = ctor(testLockerName, ipc.O_CREATE_ONLY|ipc.O_READ_ONLY, 0666)
	a.Error(err)
	lk, err = ctor(testLockerName, ipc.O_OPEN_OR_CREATE|ipc.O_WRITE_ONLY, 0666)
	a.Error(err)
	lk, err = ctor(testLockerName, ipc.O_OPEN_ONLY|ipc.O_WRITE_ONLY, 0666)
	a.Error(err)
	lk, err = ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.NoError(err) {
		return
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	lk, err = ctor(testLockerName, ipc.O_OPEN_ONLY, 0666)
	if !a.NoError(err) {
		return
	}
	a.NoError(lk.Close())
}

func testLockerOpenMode2(t *testing.T, ctor lockerCtor, dtor lockerDtor) {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return
		}
	}
	lk, err := ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.NoError(err) {
		return
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	lk, err = ctor(testLockerName, ipc.O_OPEN_ONLY, 0666)
	if !a.NoError(err) {
		return
	}
	defer func(lk IPCLocker) {
		a.NoError(lk.Close())
	}(lk)
	_, err = ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.Error(err) {
		return
	}
}

func testLockerOpenMode3(t *testing.T, ctor lockerCtor, dtor lockerDtor) {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return
		}
	}
	lk, err := ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	lk, err = ctor(testLockerName, ipc.O_OPEN_OR_CREATE, 0666)
	if !a.NoError(err) {
		return
	}
	a.NoError(lk.Close())
}

func testLockerOpenMode4(t *testing.T, ctor lockerCtor, dtor lockerDtor) {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return
		}
	}
	lk, err := ctor(testLockerName, ipc.O_OPEN_OR_CREATE, 0666)
	if !a.NoError(err) {
		return
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	lk, err = ctor(testLockerName, ipc.O_OPEN_ONLY, 0666)
	if !a.NoError(err) {
		return
	}
	defer func(lk IPCLocker) {
		a.NoError(lk.Close())
	}(lk)
	_, err = ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.Error(err) {
		return
	}
}

func testLockerOpenMode5(t *testing.T, ctor lockerCtor, dtor lockerDtor) {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return
		}
	}
	_, err := ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.NoError(err) {
		return
	}
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return
		}
		_, err = ctor(testLockerName, ipc.O_OPEN_ONLY, 0666)
		if !a.Error(err) {
			return
		}
	}
}

func testLockerLock(t *testing.T, ctor lockerCtor, dtor lockerDtor) {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return
		}
	}
	lk, err := ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.NoError(err) || !a.NotNil(lk) {
		return
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	var wg sync.WaitGroup
	sharedValue := 0
	routines, iters := 16, 1000000
	old := runtime.GOMAXPROCS(routines)
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func() {
			lk.Lock()
			for i := 0; i < iters; i++ {
				sharedValue++
			}
			lk.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
	runtime.GOMAXPROCS(old)
	a.Equal(routines*iters, sharedValue)
}

func testLockerMemory(t *testing.T, typ string, ctor lockerCtor, dtor lockerDtor) {
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return
		}
	}
	lk, err := ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.NoError(err) {
		return
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	region, err := createMemoryRegionSimple(ipc.O_OPEN_OR_CREATE|ipc.O_READWRITE, ipc.MEM_READWRITE, 128, 0)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(region.Close())
		a.NoError(shm.DestroyMemoryObject(testMemObj))
	}()
	data := region.Data()
	for i := range data { // fill the data with correct values
		data[i] = byte(i)
	}
	args := argsForSyncTestCommand(testLockerName, typ, 4, testMemObj, 1024, data, "")
	var wg sync.WaitGroup
	var flag int32 = 1
	jobs := runtime.NumCPU() - 1
	if jobs == 0 {
		jobs = 1
	}
	wg.Add(jobs)
	for i := 0; i < jobs; i++ {
		go func() {
			for atomic.LoadInt32(&flag) != 0 {
				lk.Lock()
				// corrupt the data and then restore it.
				// as the entire operation is under mutex protection,
				// no one should see these changes.
				for i := range data {
					data[i] = byte(0)
				}
				for i := range data {
					data[i] = byte(i)
				}
				lk.Unlock()
			}
			wg.Done()
		}()
	}
	result := ipc_testing.RunTestApp(args, nil)
	atomic.StoreInt32(&flag, 0)
	wg.Wait()
	if !a.NoError(result.Err) {
		t.Logf("test app error. the output is: %s", result.Output)
	}
}

func testLockerValueInc(t *testing.T, typ string, ctor lockerCtor, dtor lockerDtor) {
	const (
		iterations = 75000
		remoteJobs = 4
		remoteIncs = int64(iterations * remoteJobs)
	)
	a := assert.New(t)
	if dtor != nil {
		if !a.NoError(dtor(testLockerName)) {
			return
		}
	}
	lk, err := ctor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.NoError(err) {
		return
	}
	defer func(lk IPCLocker) {
		if d, ok := lk.(ipc.Destroyer); ok {
			a.NoError(d.Destroy())
		} else {
			a.NoError(lk.Close())
		}
	}(lk)
	region, err := createMemoryRegionSimple(ipc.O_OPEN_OR_CREATE|ipc.O_READWRITE, ipc.MEM_READWRITE, 8, 0)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(region.Close())
		a.NoError(shm.DestroyMemoryObject(testMemObj))
	}()
	data := region.Data()
	ptr := (*int64)(allocator.ByteSliceData(data))
	args := argsForSyncInc64Command(testLockerName, typ, remoteJobs, testMemObj, iterations, "")
	var wg sync.WaitGroup
	flag := int32(1)
	jobs := runtime.NumCPU()
	if jobs == 0 {
		jobs = 1
	}
	wg.Add(jobs)
	resultChan := ipc_testing.RunTestAppAsync(args, nil)
	localIncs := int64(0)
	for i := 0; i < jobs; i++ {
		go func() {
			for atomic.LoadInt32(&flag) == 1 {
				lk.Lock()
				*ptr++
				localIncs++
				//time.Sleep(time.Millisecond)
				//println(localIncs)
				lk.Unlock()
			}
			wg.Done()
		}()
	}
	result := <-resultChan
	atomic.StoreInt32(&flag, 0)
	wg.Wait()
	if !assert.NoError(t, result.Err) {
		t.Logf("test app error. the output is: %s", result.Output)
	}
	a.Equal(remoteIncs+localIncs, *ptr)
}
