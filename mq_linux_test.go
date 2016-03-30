// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build linux

package ipc

import (
	"os"
	//	"runtime"
	"testing"
	"time"

	"bitbucket.org/avd/go-ipc/internal/test"
	"github.com/stretchr/testify/assert"
)

func linuxMqCtor(name string, perm os.FileMode) (Messenger, error) {
	return CreateLinuxMessageQueue(name, perm, 1, DefaultLinuxMqMaxMessageSize)
}

func linuxMqOpener(name string, flags int) (Messenger, error) {
	return OpenLinuxMessageQueue(name, flags)
}

func linuxMqDtor(name string) error {
	return DestroyLinuxMessageQueue(name)
}

func TestCreateLinuxMq(t *testing.T) {
	testCreateMq(t, linuxMqCtor, linuxMqDtor)
}

func TestCreateLinuxMqExcl(t *testing.T) {
	testCreateMqExcl(t, linuxMqCtor, linuxMqDtor)
}

func TestCreateLinuxMqInvalidPerm(t *testing.T) {
	testCreateMqInvalidPerm(t, linuxMqCtor, linuxMqDtor)
}

func TestOpenLinuxMq(t *testing.T) {
	testOpenMq(t, linuxMqCtor, linuxMqOpener, linuxMqDtor)
}

func TestLinuxMqSendIntSameProcess(t *testing.T) {
	testMqSendIntSameProcess(t, linuxMqCtor, linuxMqOpener, linuxMqDtor)
}

func TestLinuxMqSendStructSameProcess(t *testing.T) {
	testMqSendStructSameProcess(t, linuxMqCtor, linuxMqOpener, linuxMqDtor)
}

func TestLinuxMqSendMessageLessThenBuffer(t *testing.T) {
	testMqSendMessageLessThenBuffer(t, linuxMqCtor, linuxMqOpener, linuxMqDtor)
}

func TestLinuxMqSendNonBlock(t *testing.T) {
	testMqSendNonBlock(t, linuxMqCtor, linuxMqDtor)
}

func TestLinuxMqSendToAnotherProcess(t *testing.T) {
	testMqSendToAnotherProcess(t, linuxMqCtor, linuxMqDtor, "linux")
}

func TestLinuxMqReceiveFromAnotherProcess(t *testing.T) {
	testMqReceiveFromAnotherProcess(t, linuxMqCtor, linuxMqDtor, "linux")
}

func TestLinuxMqSendTimeout(t *testing.T) {
	testMqSendTimeout(t, linuxMqCtor, linuxMqDtor)
}

func TestLinuxMqReceiveTimeout(t *testing.T) {
	testMqReceiveTimeout(t, linuxMqCtor, linuxMqDtor)
}

// linux-mq-specific tests

func TestLinuxMqGetAttrs(t *testing.T) {
	if !assert.NoError(t, DestroyLinuxMessageQueue(testMqName)) {
		return
	}
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, 5, 121)
	if !assert.NoError(t, err) {
		return
	}
	defer mq.Destroy()
	assert.NoError(t, mq.Send(make([]byte, 1)))
	attrs, err := mq.GetAttrs()
	assert.NoError(t, err)
	assert.Equal(t, 5, attrs.Maxmsg)
	assert.Equal(t, 121, attrs.Msgsize)
	assert.Equal(t, 1, attrs.Curmsgs)
}

func TestLinuxMqNotifyOnce(t *testing.T) {
	if !assert.NoError(t, DestroyLinuxMessageQueue(testMqName)) {
		return
	}
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, 5, 121)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		assert.NoError(t, mq.Destroy())
	}()
	ch := make(chan int)
	assert.NoError(t, mq.Notify(ch))
	go func() {
		mq.Send(make([]byte, 1))
	}()
	assert.Equal(t, mq.ID(), <-ch)
}

func TestLinuxMqNotifyTwice(t *testing.T) {
	if !assert.NoError(t, DestroyLinuxMessageQueue(testMqName)) {
		return
	}
	mq, err := CreateLinuxMessageQueue(testMqName, 0666, 5, 121)
	if !assert.NoError(t, err) {
		return
	}
	defer mq.Destroy()
	ch := make(chan int)
	assert.NoError(t, mq.Notify(ch))
	assert.Error(t, mq.Notify(ch))
	assert.NoError(t, mq.NotifyCancel())
	assert.NoError(t, mq.Notify(ch))
}

func TestLinuxMqNotifyAnotherProcess(t *testing.T) {
	if !assert.NoError(t, DestroyLinuxMessageQueue(testMqName)) {
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
	args := argsForMqNotifyWaitCommand(testMqName, 2000, "linux", "")
	resultChan := ipc_testing.RunTestAppAsync(args, nil)
	endChan := make(chan struct{})
	go func() {
		// as the app needs some time for startup,
		// we can't just send 1 message, because if the app calls notify()
		// after the message is sent, notify() won't work
		// this is to ensure, that the test app will start and receive the notification.
		// it guaranteed has 300ms between send() and receive()
		for {
			assert.NoError(t, mq.SendTimeoutPriority(data, 0, time.Millisecond*1000))
			<-time.After(time.Millisecond * 300)
			assert.NoError(t, mq.Receive(data))
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
