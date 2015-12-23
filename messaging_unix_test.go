// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

const (
	testMqName = "go-ipc.testmq"
)

func TestCreateMq(t *testing.T) {
	_, err := CreateMessageQueue(testMqName, false, 0666, DefaultMqMaxSize, DefaultMqMaxMessageSize)
	if assert.NoError(t, err) {
		assert.NoError(t, DestroyMessageQueue(testMqName))
	}
}

func TestCreateMqExcl(t *testing.T) {
	_, err := CreateMessageQueue(testMqName, false, 0666, DefaultMqMaxSize, DefaultMqMaxMessageSize)
	if !assert.NoError(t, err) {
		return
	}
	_, err = CreateMessageQueue(testMqName, true, 0666, DefaultMqMaxSize, DefaultMqMaxMessageSize)
	assert.Error(t, err)
}

func TestCreateMqOpenOnly(t *testing.T) {
	_, err := CreateMessageQueue(testMqName, false, 0666, DefaultMqMaxSize, DefaultMqMaxMessageSize)
	assert.NoError(t, err)
	assert.NoError(t, DestroyMessageQueue(testMqName))
	_, err = OpenMessageQueue(testMqName, O_READ_ONLY)
	assert.Error(t, err)
}

func TestMqSendIntSameProcess(t *testing.T) {
	var message int = 1122
	mq, err := CreateMessageQueue(testMqName, false, 0666, DefaultMqMaxSize, int(unsafe.Sizeof(int(0))))
	if !assert.NoError(t, err) {
		return
	}
	defer mq.Destroy()
	go func() {
		assert.NoError(t, mq.Send(message, 1))
	}()
	var received int
	var prio int
	mqr, err := OpenMessageQueue(testMqName, O_READ_ONLY)
	assert.NoError(t, err)
	assert.NoError(t, mqr.Receive(&received, &prio))
}

func TestMqSendSliceSameProcess(t *testing.T) {
	mq, err := CreateMessageQueue(testMqName, false, 0666, DefaultMqMaxSize, 32)
	if !assert.NoError(t, err) {
		return
	}
	message := make([]byte, 32)
	for i, _ := range message {
		message[i] = byte(i)
	}
	go func() {
		assert.NoError(t, mq.Send(message, 1))
	}()
	received := make([]byte, 32)
	mqr, err := OpenMessageQueue(testMqName, O_READ_ONLY)
	assert.NoError(t, err)
	defer mqr.Destroy()
	assert.NoError(t, mqr.ReceiveTimeout(received, nil, 300*time.Millisecond))
	assert.Equal(t, received, message)
}

func TestMqGetAttrs(t *testing.T) {
	mq, err := CreateMessageQueue(testMqName, false, 0666, 5, 121)
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
	mq, err := CreateMessageQueue(testMqName, false, 0666, 1, 8)
	assert.NoError(t, err)
	defer mq.Destroy()
	assert.NoError(t, mq.Send(0, 0))
	assert.NoError(t, mq.SetNonBlock(true))
	assert.Error(t, mq.Send(0, 0))
}
