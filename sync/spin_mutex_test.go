// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

const testSpinMutexName = "spin-test"

func TestSpinMutexOpenMode(t *testing.T) {
	if !assert.NoError(t, DestroySpinMutex(testSpinMutexName)) {
		return
	}
	mut, err := NewSpinMutex(testSpinMutexName, O_READWRITE, 0666)
	assert.Error(t, err)
	mut, err = NewSpinMutex(testSpinMutexName, O_CREATE_ONLY|O_READ_ONLY, 0666)
	assert.Error(t, err)
	mut, err = NewSpinMutex(testSpinMutexName, O_OPEN_OR_CREATE|O_WRITE_ONLY, 0666)
	assert.Error(t, err)
	mut, err = NewSpinMutex(testSpinMutexName, O_OPEN_ONLY|O_WRITE_ONLY, 0666)
	assert.Error(t, err)
	mut, err = NewSpinMutex(testSpinMutexName, O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func(m *SpinMutex) {
		assert.NoError(t, m.Destroy())
	}(mut)
	mut, err = NewSpinMutex(testSpinMutexName, O_OPEN_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	assert.NoError(t, mut.Finish())
}

func TestSpinMutexOpenMode2(t *testing.T) {
	if !assert.NoError(t, DestroySpinMutex(testSpinMutexName)) {
		return
	}
	mut, err := NewSpinMutex(testSpinMutexName, O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func(m *SpinMutex) {
		assert.NoError(t, m.Destroy())
	}(mut)
	mut, err = NewSpinMutex(testSpinMutexName, O_OPEN_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	assert.NoError(t, mut.Finish())
	mut, err = NewSpinMutex(testSpinMutexName, O_CREATE_ONLY, 0666)
	if !assert.Error(t, err) {
		return
	}
}

func TestSpinMutexOpenMode3(t *testing.T) {
	if !assert.NoError(t, DestroySpinMutex(testSpinMutexName)) {
		return
	}
	mut, err := NewSpinMutex(testSpinMutexName, O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func(m *SpinMutex) {
		assert.NoError(t, m.Destroy())
	}(mut)
	mut, err = NewSpinMutex(testSpinMutexName, O_OPEN_OR_CREATE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	assert.NoError(t, mut.Finish())
}

func TestSpinMutexOpenMode4(t *testing.T) {
	if !assert.NoError(t, DestroySpinMutex(testSpinMutexName)) {
		return
	}
	mut, err := NewSpinMutex(testSpinMutexName, O_OPEN_OR_CREATE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	assert.NoError(t, mut.Destroy())
}

func TestSpinMutexMemory(t *testing.T) {
	if !assert.NoError(t, DestroySpinMutex(testSpinMutexName)) {
		return
	}
	mut, err := NewSpinMutex(testSpinMutexName, O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mut.Destroy()
	region, err := createMemoryRegionSimple(O_OPEN_OR_CREATE|O_READWRITE, MEM_READWRITE, 128, 0)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		region.Close()
		DestroyMemoryObject(defaultObjectName)
	}()
	data := region.Data()
	for i := range data { // fill the data with correct values
		data[i] = byte(i)
	}
	args := argsForSyncTestCommand(testSpinMutexName, "spin", 128, defaultObjectName, 512, data, "")
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
	result := runTestApp(args, nil)
	atomic.StoreInt32(&flag, 0)
	wg.Wait()
	if !assert.NoError(t, result.err) {
		t.Logf("test app error. the output is: %s", result.output)
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
	mut, err := NewSpinMutex(testSpinMutexName, O_CREATE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mut.Destroy()
	region, err := createMemoryRegionSimple(O_OPEN_OR_CREATE|O_READWRITE, MEM_READWRITE, 8, 0)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		region.Close()
		DestroyMemoryObject(defaultObjectName)
	}()
	data := region.Data()
	ptr := (*int64)(unsafe.Pointer(&(data[0])))
	args := argsForSyncInc64Command(testSpinMutexName, "spin", remoteJobs, defaultObjectName, iterations)
	var wg sync.WaitGroup
	flag := int32(1)
	wg.Add(jobs)
	resultChan := runTestAppAsync(args, nil)
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
	if !assert.NoError(t, result.err) {
		t.Logf("test app error. the output is: %s", result.output)
	}
	assert.Equal(t, remoteIncs+localIncs, *ptr)
}
