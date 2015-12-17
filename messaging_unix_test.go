// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"testing"
	"time"

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

func TestMqSameProcess(t *testing.T) {
	mq, err := NewMessageQueue(testMqName, O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer mq.Destroy()
	var message int = 1122
	go func() {
		err := mq.Send(message, 0)
		if err != nil {
			print(err.Error())
		}
		assert.NoError(t, err)
	}()
	message = 0
	<-time.After(time.Second)
	if !assert.NoError(t, mq.Receive(&message, nil)) {
		return
	}
	assert.Equal(t, 1122, message)
}
