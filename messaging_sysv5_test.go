// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"fmt"
	"os"
	"testing"
)

func sysVMqCtor(name string, perm os.FileMode) (Messenger, error) {
	return CreateSystemVMessageQueue(name, perm)
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

func TestSysVMqSendInvalidType(t *testing.T) {
	testMqSendInvalidType(t, sysVMqCtor, sysVMqDtor)
}

func TestSysVMqSendIntSameProcess(t *testing.T) {
	testMqSendIntSameProcess(t, sysVMqCtor, sysVMqOpener, sysVMqDtor)
}

func TestSysVMqSendIntSameProcess2(t *testing.T) {
	DestroySystemVMessageQueue(testMqName)
	var toSend int64 = 0x1234567812345678
	var toReceive int32
	m1, err := CreateSystemVMessageQueue(testMqName, 0666)
	m2, err := OpenSystemVMessageQueue(testMqName, 0)
	go m1.Send(toSend)
	err = m2.Receive(&toReceive)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Printf("0x%x\n", toReceive)
	}
}
