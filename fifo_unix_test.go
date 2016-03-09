// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"testing"
	"time"

	ipc_test "bitbucket.org/avd/go-ipc/internal/test"

	"github.com/stretchr/testify/assert"
)

const testFifoName = "go-fifo-test"

// tests whether we can create a fifo in the directiry chosen by the library
func TestFifoCreate(t *testing.T) {
	if !assert.NoError(t, DestroyFifo(testFifoName)) {
		return
	}
	fifo, err := NewFifo(testFifoName, O_CREATE_ONLY|O_READ_ONLY|O_NONBLOCK, 0666)
	if assert.NoError(t, err) {
		assert.NoError(t, fifo.Destroy())
	}
}

// tests whether we can create a fifo in the directiry chosen by the calling program
func TestFifoCreateAbsPath(t *testing.T) {
	if !assert.NoError(t, DestroyFifo("/tmp/go-fifo-test")) {
		return
	}
	fifo, err := NewFifo("/tmp/go-fifo-test", O_CREATE_ONLY|O_READ_ONLY|O_NONBLOCK, 0666)
	if assert.NoError(t, err) {
		assert.NoError(t, fifo.Destroy())
	}
}

// 1) write data into a fifo in a separate process in blocking mode
// 2) read that data in our process in blocking mode
// 3) compare the results
func TestFifoBlockRead(t *testing.T) {
	testData := []byte{0, 128, 255}
	if !assert.NoError(t, DestroyFifo(testFifoName)) {
		return
	}
	defer DestroyFifo(testFifoName)
	buff := make([]byte, len(testData))
	appKillChan := make(chan bool, 1)
	defer func() { appKillChan <- true }()
	ch := ipc_test.RunTestAppAsync(argsForFifoWriteCommand(testFifoName, false, testData), appKillChan)
	var err error
	success := ipc_test.WaitForFunc(func() {
		fifo, err := NewFifo(testFifoName, O_OPEN_OR_CREATE|O_READ_ONLY, 0666)
		if !assert.NoError(t, err) {
			return
		}
		_, err = fifo.Read(buff)
	}, time.Second*2)
	if !assert.True(t, success) || !assert.NoError(t, err) ||
		!assert.Equal(t, testData, buff) {
		return
	}
	if !assert.Equal(t, testData, buff) {
		return
	}
	appResult, success := ipc_test.WaitForAppResultChan(ch, time.Second)
	if !assert.True(t, success) {
		return
	}
	assert.NoError(t, appResult.Err)
}

// 1) write data into a fifo in blocking mode
// 2) read that data in another process in blocking mode
// 3) the results are compared by another process
func TestFifoBlockWrite(t *testing.T) {
	testData := []byte{0, 128, 255}
	if !assert.NoError(t, DestroyFifo(testFifoName)) {
		return
	}
	defer DestroyFifo(testFifoName)
	appKillChan := make(chan bool, 1)
	defer func() { appKillChan <- true }()
	ch := ipc_test.RunTestAppAsync(argsForFifoTestCommand(testFifoName, false, testData), appKillChan)
	fifo, err := NewFifo(testFifoName, O_OPEN_OR_CREATE|O_WRITE_ONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	_, err = fifo.Write(testData)
	if !assert.NoError(t, err) {
		return
	}
	appResult, success := ipc_test.WaitForAppResultChan(ch, time.Second)
	if !assert.True(t, success) {
		return
	}
	assert.NoError(t, appResult.Err)
}

// 1) write data into a fifo in non-blocking mode
// 2) read that data in another process in blocking mode
// 3) the results are compared by another process
func TestFifoNonBlockWrite(t *testing.T) {
	testData := []byte{0, 128, 255}
	if !assert.NoError(t, DestroyFifo(testFifoName)) {
		return
	}
	defer DestroyFifo(testFifoName)
	appKillChan := make(chan bool, 1)
	defer func() { appKillChan <- true }()
	ch := ipc_test.RunTestAppAsync(argsForFifoTestCommand(testFifoName, false, testData), appKillChan)
	// wait for app to launch and start reading from the fifo
	fifo, err := NewFifo(testFifoName, O_OPEN_ONLY|O_WRITE_ONLY|O_NONBLOCK, 0666)
	for n := 0; err != nil && n < 10; n++ {
		<-time.After(time.Millisecond * 200)
		fifo, err = NewFifo(testFifoName, O_OPEN_ONLY|O_WRITE_ONLY|O_NONBLOCK, 0666)
	}
	if !assert.NoError(t, err) {
		return
	}
	if written, err := fifo.Write(testData); !assert.NoError(t, err) || !assert.Equal(t, written, len(testData)) {
		return
	}
	appResult, success := ipc_test.WaitForAppResultChan(ch, time.Second)
	if !assert.True(t, success) {
		return
	}
	assert.NoError(t, appResult.Err)
}
