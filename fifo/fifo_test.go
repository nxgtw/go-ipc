// Copyright 2015 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"fmt"
	"testing"
	"time"

	"bitbucket.org/avd/go-ipc"
	ipc_test "bitbucket.org/avd/go-ipc/internal/test"

	"github.com/stretchr/testify/assert"
)

const (
	testFifoName = "go-fifo-test"
	fifoProgName = "./internal/test/fifo/main.go"
)

// FIFO memory test program

func argsForFifoCreateCommand(name string) []string {
	return []string{fifoProgName, "-object=" + name, "create"}
}

func argsForFifoDestroyCommand(name string) []string {
	return []string{fifoProgName, "-object=" + name, "destroy"}
}

func argsForFifoReadCommand(name string, nonblock bool, lenght int) []string {
	return []string{fifoProgName, "-object=" + name, "-nonblock=" + boolStr(nonblock), "read", fmt.Sprintf("%d", lenght)}
}

func argsForFifoTestCommand(name string, nonblock bool, data []byte) []string {
	strBytes := ipc_test.BytesToString(data)
	return []string{fifoProgName, "-object=" + name, "-nonblock=" + boolStr(nonblock), "test", strBytes}
}

func argsForFifoWriteCommand(name string, nonblock bool, data []byte) []string {
	strBytes := ipc_test.BytesToString(data)
	return []string{fifoProgName, "-object=" + name, "-nonblock=" + boolStr(nonblock), "write", strBytes}
}

func boolStr(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// tests whether we can create a fifo in the directiry chosen by the library
func TestFifoCreate(t *testing.T) {
	if !assert.NoError(t, Destroy(testFifoName)) {
		return
	}
	fifo, err := New(testFifoName, ipc.O_CREATE_ONLY|ipc.O_READ_ONLY|ipc.O_NONBLOCK, 0666)
	if assert.NoError(t, err) {
		assert.NoError(t, fifo.Destroy())
	}
}

// 1) write data into a fifo in a separate process in blocking mode
// 2) read that data in our process in blocking mode
// 3) compare the results
func TestFifoBlockRead(t *testing.T) {
	testData := []byte{0, 128, 255}
	if !assert.NoError(t, Destroy(testFifoName)) {
		return
	}
	defer Destroy(testFifoName)
	buff := make([]byte, len(testData))
	appKillChan := make(chan bool, 1)
	defer func() { appKillChan <- true }()
	ch := ipc_test.RunTestAppAsync(argsForFifoWriteCommand(testFifoName, false, testData), appKillChan)
	var err error
	success := ipc_test.WaitForFunc(func() {
		fifo, err2 := New(testFifoName, ipc.O_OPEN_OR_CREATE|ipc.O_READ_ONLY, 0666)
		if !assert.NoError(t, err2) {
			return
		}
		_, err2 = fifo.Read(buff)
		assert.NoError(t, err2)
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
	if !assert.NoError(t, Destroy(testFifoName)) {
		return
	}
	defer Destroy(testFifoName)
	appKillChan := make(chan bool, 1)
	defer func() { appKillChan <- true }()
	ch := ipc_test.RunTestAppAsync(argsForFifoTestCommand(testFifoName, false, testData), appKillChan)
	fifo, err := New(testFifoName, ipc.O_OPEN_OR_CREATE|ipc.O_WRITE_ONLY, 0666)
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
	if !assert.NoError(t, Destroy(testFifoName)) {
		return
	}
	defer Destroy(testFifoName)
	appKillChan := make(chan bool, 1)
	defer func() { appKillChan <- true }()
	ch := ipc_test.RunTestAppAsync(argsForFifoTestCommand(testFifoName, false, testData), appKillChan)
	// wait for app to launch and start reading from the fifo
	fifo, err := New(testFifoName, ipc.O_OPEN_ONLY|ipc.O_WRITE_ONLY|ipc.O_NONBLOCK, 0666)
	for n := 0; err != nil && n < 10; n++ {
		<-time.After(time.Millisecond * 200)
		fifo, err = New(testFifoName, ipc.O_OPEN_ONLY|ipc.O_WRITE_ONLY|ipc.O_NONBLOCK, 0666)
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
