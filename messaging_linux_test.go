// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build linux

package ipc

import (
	"testing"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/test"
	"github.com/stretchr/testify/assert"
)

const (
	testMqName = "go-ipc.testmq"
)

func TestCreateMq(t *testing.T) {
	assert.NoError(t, DestroyMessageQueue(testMqName))
	_, err := CreateLinuxMessageQueue(testMqName, 0666, DefaultMqMaxSize, DefaultMqMaxMessageSize)
	if assert.NoError(t, err) {
		assert.NoError(t, DestroyMessageQueue(testMqName))
	}
}

func TestCreateMqExcl(t *testing.T) {
	assert.NoError(t, DestroyMessageQueue(testMqName))
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, DefaultMqMaxSize, DefaultMqMaxMessageSize)
	if !assert.NoError(t, err) {
		return
	}
	_, err = CreateLinuxMessageQueue(testMqName, 0666, DefaultMqMaxSize, DefaultMqMaxMessageSize)
	assert.Error(t, err)
	assert.NoError(t, mq.Destroy())
}

func TestCreateMqOpenOnly(t *testing.T) {
	assert.NoError(t, DestroyMessageQueue(testMqName))
	_, err := CreateLinuxMessageQueue(testMqName, 0666, DefaultMqMaxSize, DefaultMqMaxMessageSize)
	assert.NoError(t, err)
	assert.NoError(t, DestroyMessageQueue(testMqName))
	_, err = OpenLinuxMessageQueue(testMqName, O_READ_ONLY)
	assert.Error(t, err)
}

func TestMqSendInvalidType(t *testing.T) {
	assert.NoError(t, DestroyMessageQueue(testMqName))
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, DefaultMqMaxSize, int(unsafe.Sizeof(int(0))))
	if !assert.NoError(t, err) {
		return
	}
	defer mq.Destroy()
	assert.Error(t, mq.Send("string", 0))
	structWithString := struct{ a string }{"string"}
	assert.Error(t, mq.Send(structWithString, 0))
	var slslByte [][]byte
	assert.Error(t, mq.Send(slslByte, 0))
}

func TestMqSendIntSameProcess(t *testing.T) {
	var message = 1122
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, DefaultMqMaxSize, int(unsafe.Sizeof(int(0))))
	if !assert.NoError(t, err) {
		return
	}
	defer mq.Destroy()
	go func() {
		assert.NoError(t, mq.Send(message, 1))
	}()
	var received int
	var prio int
	mqr, err := OpenLinuxMessageQueue(testMqName, O_READ_ONLY)
	assert.NoError(t, err)
	assert.NoError(t, mqr.Receive(&received, &prio))
}

func TestMqSendSliceSameProcess(t *testing.T) {
	type testStruct struct {
		arr [16]int
		c   complex128
		s   struct {
			a, b byte
		}
		f float64
	}
	message := testStruct{c: complex(2, -3), f: 11.22, s: struct{ a, b byte }{127, 255}}
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, DefaultMqMaxSize, int(unsafe.Sizeof(message)))
	if !assert.NoError(t, err) {
		return
	}
	go func() {
		assert.NoError(t, mq.Send(message, 1))
	}()
	received := &testStruct{}
	mqr, err := OpenLinuxMessageQueue(testMqName, O_READ_ONLY)
	if !assert.NoError(t, err) {
		return
	}
	defer mqr.Destroy()
	assert.NoError(t, mqr.ReceiveTimeout(received, nil, 300*time.Millisecond))
	assert.Equal(t, message, *received)
}

func TestMqGetAttrs(t *testing.T) {
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, 5, 121)
	assert.NoError(t, err)
	defer mq.Destroy()
	assert.NoError(t, mq.Send(0, 0))
	attrs, err := mq.GetAttrs()
	assert.NoError(t, err)
	assert.Equal(t, 5, attrs.Maxmsg)
	assert.Equal(t, 121, attrs.Msgsize)
	assert.Equal(t, 1, attrs.Curmsgs)
}

func TestMqSetNonBlock(t *testing.T) {
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, 1, 8)
	assert.NoError(t, err)
	defer mq.Destroy()
	assert.NoError(t, mq.Send(0, 0))
	assert.NoError(t, mq.SetBlocking(false))
	assert.Error(t, mq.Send(0, 0))
}

func TestMqNotify(t *testing.T) {
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, 5, 121)
	assert.NoError(t, err)
	defer mq.Destroy()
	ch := make(chan int)
	assert.NoError(t, mq.Notify(ch))
	go func() {
		mq.Send(0, 0)
	}()
	assert.Equal(t, mq.ID(), <-ch)
}

func TestMqNotifyTwice(t *testing.T) {
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, 5, 121)
	assert.NoError(t, err)
	defer mq.Destroy()
	ch := make(chan int)
	assert.NoError(t, mq.Notify(ch))
	assert.Error(t, mq.Notify(ch))
	assert.NoError(t, mq.NotifyCancel())
	assert.NoError(t, mq.Notify(ch))
}

func TestMqSendToAnotherProcess(t *testing.T) {
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, 5, 16)
	assert.NoError(t, err)
	defer mq.Destroy()
	data := make([]byte, 16)
	for i := range data {
		data[i] = byte(i)
	}
	args := argsForMqTestCommand(testMqName, 1000, 1, data)
	go func() {
		assert.NoError(t, mq.SendTimeout(data, 1, time.Millisecond*2000))
	}()
	result := ipc_test.RunTestApp(args, nil)
	if !assert.NoError(t, result.Err) {
		t.Logf("program output is %s", result.Output)
	}
}

func TestMqReceiveFromAnotherProcess(t *testing.T) {
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, 5, 16)
	assert.NoError(t, err)
	defer mq.Destroy()
	data := make([]byte, 16)
	for i := range data {
		data[i] = byte(i)
	}
	args := argsForMqSendCommand(testMqName, 2000, 3, data)
	result := ipc_test.RunTestApp(args, nil)
	if !assert.NoError(t, result.Err) {
		t.Logf("program output is %s", result.Output)
	}
	received := make([]byte, 16)
	var prio int
	assert.NoError(t, mq.ReceiveTimeout(received, &prio, time.Millisecond*2000))
	assert.Equal(t, prio, 3)
	assert.Equal(t, data, received)
}

func TestMqNotifyAnotherProcess(t *testing.T) {
	if !assert.NoError(t, DestroyMessageQueue(testMqName)) {
		return
	}
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, 4, 16)
	if !assert.NoError(t, err) {
		return
	}
	defer mq.Destroy()
	data := make([]byte, 16)
	for i := range data {
		data[i] = byte(i)
	}
	args := argsForMqNotifyWaitCommand(testMqName, 2000)
	resultChan := ipc_test.RunTestAppAsync(args, nil)
	endChan := make(chan struct{})
	go func() {
		// as the app needs some time for startup,
		// we can't just send 1 message, because if the app calls notify()
		// after the message is sent, notify() won't work
		// this is to ensure, that the test app will start and receive the notification.
		// it guaranteed has 300ms between send() and receive()
		for {
			assert.NoError(t, mq.SendTimeout(data, 0, time.Millisecond*1000))
			<-time.After(time.Millisecond * 300)
			assert.NoError(t, mq.Receive(data, nil))
			select {
			case <-endChan:
				return
			default:
			}
		}
	}()
	result := <-resultChan
	endChan <- struct{}{}
	if !assert.NoError(t, result.Err) {
		t.Logf("program output is %q", result.Output)
	}
}
