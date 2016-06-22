// Copyright 2016 Aleksandr Demakin. All rights reserved.

//+build linux_mq

package mq

import "os"

func createMQ(name string, flag int, perm os.FileMode) (Messenger, error) {
	mq, err := CreateLinuxMessageQueue(name, flag, perm, DefaultLinuxMqMaxSize, DefaultLinuxMqMessageSize)
	if err != nil {
		return nil, err
	}
	return mq, nil
}

func openMQ(name string, flags int) (Messenger, error) {
	mq, err := OpenLinuxMessageQueue(name, flags|os.O_RDWR)
	if err != nil {
		return nil, err
	}
	return mq, nil
}

func destroyMq(name string) error {
	return DestroyLinuxMessageQueue(name)
}
