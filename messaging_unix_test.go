// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMQ(t *testing.T) {
	mq, err := NewMessageQueue(MSGQUEUE_NEW, O_OPEN_OR_CREATE, 0666)
	if assert.NoError(t, err) {
		assert.NoError(t, DestroyMessageQueue(mq.Id()))
	}
}

func TestCreateMQExcl(t *testing.T) {
	mq, err := NewMessageQueue(MSGQUEUE_NEW, O_OPEN_OR_CREATE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	_, err = NewMessageQueue(mq.Id(), O_CREATE_ONLY, 0666)
	assert.Error(t, err)
}

func TestCreateMQOpenOnly(t *testing.T) {
	mq, err := NewMessageQueue(MSGQUEUE_NEW, O_OPEN_OR_CREATE, 0666)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NoError(t, DestroyMessageQueue(mq.Id())) {
		return
	}
	_, err = NewMessageQueue(mq.Id(), O_OPEN_ONLY, 0666)
	assert.Error(t, err)
}

func TestCreateMQWrongFlags(t *testing.T) {
	_, err := NewMessageQueue(MSGQUEUE_NEW, O_OPEN_ONLY, 0666)
	assert.Error(t, err)
}
