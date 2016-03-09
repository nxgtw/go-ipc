// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"os"
	"testing"
)

func sysVMqCtor(name string, perm os.FileMode) (Messenger, error) {
	return CreateSystemVMessageQueue(name, perm)
}

func sysVMqOpener(name string, flags int) (Messenger, error) {
	return OpenSystemVMessageQueue(name, flags)
}

func sysvMqDtor(name string) error {
	return DestroySystemVMessageQueue(name)
}

func TestCreateSysVMq(t *testing.T) {
	testCreateMq(t, sysVMqCtor, sysvMqDtor)
}

func TestCreateSysVMqExcl(t *testing.T) {
	testCreateMqExcl(t, sysVMqCtor, sysvMqDtor)
}

func TestCreateSysVMqInvalidPerm(t *testing.T) {
	testCreateMqInvalidPerm(t, sysVMqCtor, sysvMqDtor)
}

func TestOpenSysVMq(t *testing.T) {
	testOpenMq(t, sysVMqCtor, sysVMqOpener, sysvMqDtor)
}
