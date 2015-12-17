// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testMqName = "go-ipc.testmq"
)

func TestCreateMq(t *testing.T) {
	_, err := NewMessageQueue(testMqName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if assert.NoError(t, err) {
		assert.NoError(t, DestroyMessageQueue(testMqName))
	}
}

func TestCreateMqExcl(t *testing.T) {
	_, err := NewMessageQueue(testMqName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	_, err = NewMessageQueue(testMqName, O_CREATE_ONLY|O_READWRITE, 0666)
	assert.Error(t, err)
}

func TestCreateMqOpenOnly(t *testing.T) {
	_, err := NewMessageQueue(testMqName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NoError(t, DestroyMessageQueue(testMqName)) {
		return
	}
	_, err = NewMessageQueue(testMqName, O_OPEN_ONLY|O_READWRITE, 0666)
	assert.Error(t, err)
}

/*
func TestMessagingSameProcess(t *testing.T) {
	mq, err := NewMessageQueue(MSGQUEUE_NEW, O_OPEN_OR_CREATE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mq.Destroy()
	var message int = 1122
	go func() {
		assert.NoError(t, mq.Send(1, 0, message))
	}()
	message = 0
	if !assert.NoError(t, mq.Receive(1, 0, &message)) {
		return
	}
	assert.Equal(t, 1122, message)
}
*/
