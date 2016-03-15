// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

// +build !linux_mq

package ipc

import "os"

func createMQ(name string, perm os.FileMode) (Messenger, error) {
	return CreateSystemVMessageQueue(name, perm)
}

func openMQ(name string, flags int) (Messenger, error) {
	return OpenSystemVMessageQueue(name, flags)
}

func destroyMq(name string) error {
	return DestroySystemVMessageQueue(name)
}
