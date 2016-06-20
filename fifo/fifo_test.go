// Copyright 2015 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"fmt"
	"os"
	"testing"
	"time"

	"bitbucket.org/avd/go-ipc/internal/test"

	"github.com/stretchr/testify/assert"
)

const (
	testFifoName = "go-fifo-test"
	fifoProgName = "./internal/test/main.go"
)

var (
	testData []byte
)

func init() {
	testData = make([]byte, 2048)
	for i := range testData {
		testData[i] = byte(i)
	}
}

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
	strBytes := testutil.BytesToString(data)
	return []string{fifoProgName, "-object=" + name, "-nonblock=" + boolStr(nonblock), "test", strBytes}
}

func argsForFifoWriteCommand(name string, nonblock bool, data []byte) []string {
	strBytes := testutil.BytesToString(data)
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
	fifo, err := New(testFifoName, os.O_CREATE|os.O_EXCL|os.O_RDONLY|O_NONBLOCK, 0666)
	if assert.NoError(t, err) {
		assert.NoError(t, fifo.Destroy())
		assert.Error(t, fifo.Destroy())
	}
}

// 1) write data into a fifo in the same process in blocking mode
// 2) read that data in our process in blocking mode
// 3) compare the results
func TestFifoBlockReadSameProcess(t *testing.T) {
	if !assert.NoError(t, Destroy(testFifoName)) {
		return
	}
	defer Destroy(testFifoName)
	go func() {
		fifo, err2 := New(testFifoName, os.O_CREATE|os.O_WRONLY, 0666)
		if !assert.NoError(t, err2) {
			return
		}
		n, err2 := fifo.Write(testData)
		assert.NoError(t, err2)
		assert.Equal(t, len(testData), n)
		assert.NoError(t, fifo.Close())
	}()
	buff := make([]byte, len(testData))
	success := testutil.WaitForFunc(func() {
		fifo, err2 := New(testFifoName, os.O_CREATE|os.O_RDONLY, 0666)
		if !assert.NoError(t, err2) {
			return
		}
		_, err2 = fifo.Read(buff)
		assert.NoError(t, err2)
		assert.NoError(t, fifo.Close())
	}, time.Second*2)
	if !assert.True(t, success) || !assert.Equal(t, testData, buff) {
		return
	}
	if !assert.Equal(t, testData, buff) {
		return
	}
}

// 1) write data into a fifo in a separate process in blocking mode
// 2) read that data in our process in blocking mode
// 3) compare the results
func TestFifoBlockReadAnotherProcess(t *testing.T) {
	if !assert.NoError(t, Destroy(testFifoName)) {
		return
	}
	defer Destroy(testFifoName)
	buff := make([]byte, len(testData))
	appKillChan := make(chan bool, 1)
	defer func() { appKillChan <- true }()
	ch := testutil.RunTestAppAsync(argsForFifoWriteCommand(testFifoName, false, testData), appKillChan)
	var err error
	success := testutil.WaitForFunc(func() {
		var fifo Fifo
		fifo, err = New(testFifoName, os.O_CREATE|os.O_RDONLY, 0666)
		if !assert.NoError(t, err) {
			return
		}
		_, err = fifo.Read(buff)
		assert.NoError(t, err)
		assert.NoError(t, fifo.Close())
	}, time.Second*20)
	if err != nil {
		return
	}
	if !assert.True(t, success) || !assert.Equal(t, testData, buff) {
		return
	}
	if !assert.Equal(t, testData, buff) {
		return
	}
	appResult, success := testutil.WaitForAppResultChan(ch, time.Second*2)
	if !assert.True(t, success) {
		return
	}
	assert.NoError(t, appResult.Err)
}

// 1) write data into a fifo in blocking mode
// 2) read that data in another process in blocking mode
// 3) the results are compared by another process
func TestFifoBlockWrite(t *testing.T) {
	if !assert.NoError(t, Destroy(testFifoName)) {
		return
	}
	defer Destroy(testFifoName)
	appKillChan := make(chan bool, 1)
	defer func() { appKillChan <- true }()
	ch := testutil.RunTestAppAsync(argsForFifoTestCommand(testFifoName, false, testData), appKillChan)
	fifo, err := New(testFifoName, os.O_CREATE|os.O_WRONLY, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer fifo.Close()
	n, err := fifo.Write(testData)
	if !assert.NoError(t, err) || !assert.Equal(t, len(testData), n) {
		return
	}
	appResult, success := testutil.WaitForAppResultChan(ch, time.Second*5)
	assert.True(t, success)
	if !assert.NoError(t, appResult.Err) {
		t.Log(appResult.Output)
	}
}

// 1) write data into a fifo in non-blocking mode
// 2) read that data in another process in blocking mode
// 3) the results are compared by another process
func TestFifoNonBlockWrite(t *testing.T) {
	if !assert.NoError(t, Destroy(testFifoName)) {
		return
	}
	defer Destroy(testFifoName)
	appKillChan := make(chan bool, 1)
	defer func() { appKillChan <- true }()
	ch := testutil.RunTestAppAsync(argsForFifoTestCommand(testFifoName, false, testData), appKillChan)
	// wait for app to launch and start reading from the fifo
	fifo, err := New(testFifoName, os.O_WRONLY|O_NONBLOCK, 0666)
	for n := 0; err != nil && n < 50; n++ {
		<-time.After(time.Millisecond * 200)
		fifo, err = New(testFifoName, os.O_WRONLY|O_NONBLOCK, 0666)
	}
	if !assert.NoError(t, err) {
		return
	}
	defer fifo.Close()
	if written, err := fifo.Write(testData); !assert.NoError(t, err) || !assert.Equal(t, written, len(testData)) {
		return
	}
	appResult, success := testutil.WaitForAppResultChan(ch, time.Second)
	if !assert.True(t, success) {
		return
	}
	assert.NoError(t, appResult.Err)
}

func TestFifoNonBlock1(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(Destroy(testFifoName)) {
		return
	}
	var err error
	done := testutil.WaitForFunc(func() {
		_, err = New(testFifoName, os.O_WRONLY|O_NONBLOCK, 0666)
	}, time.Second*2)
	a.True(done)
	a.Error(err)
}

func TestFifoNonBlock2(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(Destroy(testFifoName)) {
		return
	}
	var err error
	done := testutil.WaitForFunc(func() {
		_, err = New(testFifoName, os.O_WRONLY|O_NONBLOCK, 0666)
	}, time.Second*2)
	a.True(done)
	a.Error(err)
}

func TestFifoNonBlock3(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(Destroy(testFifoName)) {
		return
	}
	fifo, err := New(testFifoName, os.O_CREATE|os.O_RDONLY|O_NONBLOCK, 0666)
	if !a.NoError(err) {
		return
	}
	fifo2, err := New(testFifoName, os.O_WRONLY, 0666)
	if !a.NoError(err) {
		a.NoError(fifo.Close())
		return
	}
	a.NoError(fifo.Close())
	_, err = fifo2.Write(testData)
	a.Error(err)
	a.NoError(fifo2.Close())
}

func TestFifoNonBlock4(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(Destroy(testFifoName)) {
		return
	}
	fifo, err := New(testFifoName, os.O_CREATE|os.O_WRONLY|O_NONBLOCK, 0666)
	if !a.Error(err) {
		fifo.Destroy()
	}
}
