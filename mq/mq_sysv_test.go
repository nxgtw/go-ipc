// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package mq

import (
	"os"
	"testing"
)

func sysVMqCtor(name string, flag int, perm os.FileMode) (Messenger, error) {
	return CreateSystemVMessageQueue(name, flag, perm)
}

func sysVMqOpener(name string, flags int) (Messenger, error) {
	return OpenSystemVMessageQueue(name, flags)
}

func sysVMqDtor(name string) error {
	return DestroySystemVMessageQueue(name)
}

func TestCreateSysVMq(t *testing.T) {
	testCreateMq(t, sysVMqCtor, sysVMqDtor)
}

func TestCreateSysVMqExcl(t *testing.T) {
	testCreateMqExcl(t, sysVMqCtor, sysVMqDtor)
}

func TestCreateSysVMqInvalidPerm(t *testing.T) {
	testCreateMqInvalidPerm(t, sysVMqCtor, sysVMqDtor)
}

func TestOpenSysVMq(t *testing.T) {
	testOpenMq(t, sysVMqCtor, sysVMqOpener, sysVMqDtor)
}

func TestSysVMqSendIntSameProcess(t *testing.T) {
	testMqSendIntSameProcess(t, sysVMqCtor, sysVMqOpener, sysVMqDtor)
}

func TestSysVMqSendStructSameProcess(t *testing.T) {
	testMqSendStructSameProcess(t, sysVMqCtor, sysVMqOpener, sysVMqDtor)
}

func TestSysVMqSendMessageLessThenBuffer(t *testing.T) {
	testMqSendMessageLessThenBuffer(t, sysVMqCtor, sysVMqOpener, sysVMqDtor)
}

func TestSysVMqSendNonBlock(t *testing.T) {
	testMqSendNonBlock(t, sysVMqCtor, sysVMqDtor)
}

func TestSysVMqSendToAnotherProcess(t *testing.T) {
	testMqSendToAnotherProcess(t, sysVMqCtor, sysVMqDtor, "sysv")
}

func TestSysVMqReceiveNonBlock(t *testing.T) {
	testMqReceiveNonBlock(t, sysVMqCtor, sysVMqDtor)
}

func TestSysVMqReceiveFromAnotherProcess(t *testing.T) {
	testMqReceiveFromAnotherProcess(t, sysVMqCtor, sysVMqDtor, "sysv")
}
