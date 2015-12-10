// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testRwMutexName = "go-rwm-test"

func TestLockRwMutex(t *testing.T) {
	mut, err := NewRwMutex(testRwMutexName, O_OPEN_OR_CREATE, 0666)
	if !assert.NoError(t, err) || !assert.NotNil(t, mut) {
		return
	}
	defer mut.Destroy()
	var wg sync.WaitGroup
	sharedValue := 0
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			mut.Lock()
			for i := 0; i < 1000; i++ {
				sharedValue++
			}
			mut.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, 30000, sharedValue)
}

func TestRwMutexOpenMode(t *testing.T) {
	mut, err := NewRwMutex(testRwMutexName, O_READWRITE, 0666)
	assert.Error(t, err)
	mut, err = NewRwMutex(testRwMutexName, O_CREATE_ONLY|O_READ_ONLY, 0666)
	assert.Error(t, err)
	mut, err = NewRwMutex(testRwMutexName, O_OPEN_OR_CREATE|O_WRITE_ONLY, 0666)
	assert.Error(t, err)
	mut, err = NewRwMutex(testRwMutexName, O_OPEN_ONLY|O_WRITE_ONLY, 0666)
	assert.Error(t, err)
	mut, err = NewRwMutex(testRwMutexName, O_OPEN_OR_CREATE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mut.Destroy()
	mut, err = NewRwMutex(testRwMutexName, O_OPEN_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
}

func TestRwMutexOpenMode2(t *testing.T) {
	mut, err := NewRwMutex(testRwMutexName, O_OPEN_OR_CREATE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mut.Destroy()
	mut, err = NewRwMutex(testRwMutexName, O_OPEN_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	mut, err = NewRwMutex(testRwMutexName, O_CREATE_ONLY, 0666)
	if !assert.Error(t, err) {
		return
	}
}

func TestRwMutexOpenMode3(t *testing.T) {
	if !assert.NoError(t, DestroyRwMutex(testRwMutexName)) {
		return
	}
	_, err := NewRwMutex(testRwMutexName, O_CREATE_ONLY, 0666)
	if !assert.Error(t, err) {
		return
	}
}
