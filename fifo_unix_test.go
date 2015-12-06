// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testFifoName = "go-fifo-test"

// tests whether we can create a fifo in the directiry chosen by the library
func TestFifoCreate(t *testing.T) {
	if !assert.NoError(t, DestroyFifo(testFifoName)) {
		return
	}
	fifo, err := NewFifo(testFifoName, O_READWRITE, 0666)
	if assert.NoError(t, err) {
		assert.NoError(t, fifo.Destroy())
	}
}

// tests whether we can create a fifo in the directiry chosen by the calling program
func TestFifoCreateAbsPath(t *testing.T) {
	if !assert.NoError(t, DestroyFifo("/tmp/go-fifo-test")) {
		return
	}
	fifo, err := NewFifo("/tmp/go-fifo-test", O_READWRITE, 0666)
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
	defer close(appKillChan)
	ch := runTestAppAsync(argsForFifoWriteCommand(testFifoName, false, testData), appKillChan)
	var err error
	success := waitForFunc(func() {
		fifo, err := NewFifo(testFifoName, O_READ_ONLY, 0666)
		if !assert.NoError(t, err) {
			return
		}
		_, err = fifo.Read(buff)
	}, time.Second)
	if !assert.True(t, success) || !assert.NoError(t, err) ||
		!assert.Equal(t, testData, buff) {
		return
	}
	if !assert.Equal(t, testData, buff) {
		return
	}
	appResult, success := waitForAppResultChan(ch, time.Second)
	if !assert.True(t, success) {
		appKillChan <- true
		return
	}
	assert.NoError(t, appResult.err)
}
