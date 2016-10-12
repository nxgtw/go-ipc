// Copyright 2016 Aleksandr Demakin. All rights reserved.

//+build linux freebsd

package mq

import "testing"

func TestFastMqSendStructSameProcess(t *testing.T) {
	testMqSendStructSameProcess(t, fastMqCtor, fastMqOpener, fastMqDtor)
}

func TestFastMqSendMessageLessThenBuffer(t *testing.T) {
	testMqSendMessageLessThenBuffer(t, fastMqCtor, fastMqOpener, fastMqDtor)
}

func TestFastMqReceiveFromAnotherProcess(t *testing.T) {
	testMqReceiveFromAnotherProcess(t, fastMqCtor, fastMqDtor, "fast")
}

func TestFastMqSendTimeout(t *testing.T) {
	testMqSendTimeout(t, fastMqCtor, fastMqDtor)
}

func TestFastMqReceiveTimeout(t *testing.T) {
	testMqReceiveTimeout(t, fastMqCtor, fastMqDtor)
}
