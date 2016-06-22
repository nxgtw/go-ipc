// Copyright 2016 Aleksandr Demakin. All rights reserved.

//+build windows

package mq

import "os"

func createMQ(name string, flag int, perm os.FileMode) (Messenger, error) {
	mq, err := CreateFastMq(name, flag, perm, DefaultFastMqMaxSize, DefaultFastMqMessageSize)
	if err != nil {
		return nil, err
	}
	return mq, nil
}

func openMQ(name string, flags int) (Messenger, error) {
	mq, err := OpenFastMq(name, flags)
	if err != nil {
		return nil, err
	}
	return mq, nil
}

func destroyMq(name string) error {
	return DestroyFastMq(name)
}
