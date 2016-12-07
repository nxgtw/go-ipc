// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mq

import (
	"os"
	"testing"
)

func fastMqCtor(name string, flag int, perm os.FileMode) (Messenger, error) {
	return CreateFastMq(name, flag, perm, 1, DefaultFastMqMessageSize)
}

func fastMqOpener(name string, flags int) (Messenger, error) {
	return OpenFastMq(name, flags)
}

func fastMqDtor(name string) error {
	return DestroyFastMq(name)
}

func fastMqCtorPrio(name string, flag int, perm os.FileMode, maxQueueSize, maxMsgSize int) (PriorityMessenger, error) {
	return CreateFastMq(name, flag, perm, maxQueueSize, maxMsgSize)
}

func fastMqOpenerPrio(name string, flags int) (PriorityMessenger, error) {
	return OpenFastMq(name, flags)
}

func TestCreateFastMq(t *testing.T) {
	testCreateMq(t, fastMqCtor, fastMqDtor)
}

func TestCreateFastMqExcl(t *testing.T) {
	testCreateMqExcl(t, fastMqCtor, fastMqDtor)
}

func TestCreateFastMqInvalidPerm(t *testing.T) {
	testCreateMqInvalidPerm(t, fastMqCtor, fastMqDtor)
}

func TestOpenFastMq(t *testing.T) {
	testOpenMq(t, fastMqCtor, fastMqOpener, fastMqDtor)
}

func TestFastMqSendIntSameProcess(t *testing.T) {
	testMqSendIntSameProcess(t, fastMqCtor, fastMqOpener, fastMqDtor)
}

func TestFastMqSendNonBlock(t *testing.T) {
	testMqSendNonBlock(t, fastMqCtor, fastMqDtor)
}

func TestFastMqSendToAnotherProcess(t *testing.T) {
	testMqSendToAnotherProcess(t, fastMqCtor, fastMqDtor, "fast")
}

func TestFastMqPrio1(t *testing.T) {
	testPrioMq1(t, fastMqCtorPrio, fastMqOpenerPrio, fastMqDtor)
}

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

func BenchmarkFastMqNonBlock(b *testing.B) {
	params := &prioBenchmarkParams{readers: 4, writers: 4, mqSize: 8, msgSize: 1024, flag: O_NONBLOCK}
	benchmarkPrioMq1(b, fastMqCtorPrio, fastMqOpenerPrio, fastMqDtor, params)
}

func BenchmarkFastMqBlock(b *testing.B) {
	params := &prioBenchmarkParams{readers: 4, writers: 4, mqSize: 8, msgSize: 1024, flag: 0}
	benchmarkPrioMq1(b, fastMqCtorPrio, fastMqOpenerPrio, fastMqDtor, params)
}
