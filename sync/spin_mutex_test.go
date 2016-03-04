// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/test"

	"github.com/stretchr/testify/assert"
)

const testSpinMutexName = "spin-test"

func spinCtor(name string, mode int, perm os.FileMode) (IPCLocker, error) {
	return NewSpinMutex(name, mode, perm)
}

func spinDtor(name string) error {
	return DestroySpinMutex(name)
}

func TestSpinMutexOpenMode(t *testing.T) {
	testLockerOpenMode(t, spinCtor, spinDtor)
}

func TestSpinMutexOpenMode2(t *testing.T) {
	testLockerOpenMode2(t, spinCtor, spinDtor)
}

func TestSpinMutexOpenMode3(t *testing.T) {
	testLockerOpenMode3(t, spinCtor, spinDtor)
}

func TestSpinMutexOpenMode4(t *testing.T) {
	testLockerOpenMode4(t, spinCtor, spinDtor)
}

func TestSpinMutexLock(t *testing.T) {
	testLockerLock(t, spinCtor, spinDtor)
}

func TestSpinMutexMemory(t *testing.T) {
	if !assert.NoError(t, DestroySpinMutex(testSpinMutexName)) {
		return
	}
	mut, err := NewSpinMutex(testSpinMutexName, ipc.O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mut.Destroy()
	region, err := createMemoryRegionSimple(ipc.O_OPEN_OR_CREATE|ipc.O_READWRITE, ipc.MEM_READWRITE, 128, 0)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		region.Close()
		ipc.DestroyMemoryObject(testMemObj)
	}()
	data := region.Data()
	for i := range data { // fill the data with correct values
		data[i] = byte(i)
	}
	args := argsForSyncTestCommand(testSpinMutexName, "spin", 128, testMemObj, 512, data, "")
	var wg sync.WaitGroup
	var flag int32 = 1
	const jobs = 4
	wg.Add(jobs)
	for i := 0; i < jobs; i++ {
		go func() {
			for atomic.LoadInt32(&flag) != 0 {
				mut.Lock()
				// corrupt the data and then restore it.
				// as the entire operation is under mutex protection,
				// no one should see these changes.
				for i := range data {
					data[i] = byte(0)
				}
				for i := range data {
					data[i] = byte(i)
				}
				mut.Unlock()
			}
			wg.Done()
		}()
	}
	result := ipc_test.RunTestApp(args, nil)
	atomic.StoreInt32(&flag, 0)
	wg.Wait()
	if !assert.NoError(t, result.Err) {
		t.Logf("test app error. the output is: %s", result.Output)
	}
}

func TestSpinMutexValueInc(t *testing.T) {
	const (
		iterations = 50000
		jobs       = 4
		remoteJobs = 4
		remoteIncs = int64(iterations * remoteJobs)
	)
	if !assert.NoError(t, DestroySpinMutex(testSpinMutexName)) {
		return
	}
	mut, err := NewSpinMutex(testSpinMutexName, ipc.O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mut.Destroy()
	region, err := createMemoryRegionSimple(ipc.O_OPEN_OR_CREATE|ipc.O_READWRITE, ipc.MEM_READWRITE, 8, 0)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		region.Close()
		ipc.DestroyMemoryObject(testMemObj)
	}()
	data := region.Data()
	ptr := (*int64)(unsafe.Pointer(&(data[0])))
	args := argsForSyncInc64Command(testSpinMutexName, "spin", remoteJobs, testMemObj, iterations)
	var wg sync.WaitGroup
	flag := int32(1)
	wg.Add(jobs)
	resultChan := ipc_test.RunTestAppAsync(args, nil)
	localIncs := int64(0)
	for i := 0; i < jobs; i++ {
		go func() {
			for atomic.LoadInt32(&flag) == 1 {
				mut.Lock()
				*ptr++
				localIncs++
				mut.Unlock()
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
	assert.Equal(t, remoteIncs+localIncs, *ptr)
}
