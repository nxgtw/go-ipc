// Copyright 2016 Aleksandr Demakin. All rights reserved.

//+build linux_mq

package ipc

import "os"

func createMQ(name string, perm os.FileMode) (Messenger, error) {
	return CreateLinuxMessageQueue(name, perm, DefaultLinuxMqMaxSize, DefaultLinuxMqMaxMessageSize)
}

func openMQ(name string, flags int) (Messenger, error) {
	return OpenLinuxMessageQueue(name, flags)
}

func destroyMq(name string) error {
	return DestroyLinuxMessageQueue(name)
}
