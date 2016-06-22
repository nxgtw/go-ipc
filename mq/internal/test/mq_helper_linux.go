// Copyright 2016 Aleksandr Demakin. All rights reserved.

package main

import (
	"fmt"
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/mq"
)

func createMqWithType(name string, perm os.FileMode, typ, opt string) (mq.Messenger, error) {
	switch typ {
	case "default":
		return mq.New(name, os.O_RDWR, perm)
	case "sysv":
		return mq.CreateSystemVMessageQueue(name, os.O_RDWR, perm)
	case "fast":
		mqSize, msgSize := mq.DefaultLinuxMqMaxSize, mq.DefaultLinuxMqMessageSize
		if first, second, err := parseTwoInts(opt); err == nil {
			mqSize, msgSize = first, second
		}
		return mq.CreateFastMq(name, 0, perm, mqSize, msgSize)
	case "linux":
		mqSize, msgSize := mq.DefaultLinuxMqMaxSize, mq.DefaultLinuxMqMessageSize
		if first, second, err := parseTwoInts(opt); err == nil {
			mqSize, msgSize = first, second
		}
		return mq.CreateLinuxMessageQueue(name, os.O_RDWR, perm, mqSize, msgSize)
	default:
		return nil, fmt.Errorf("unknown mq type %q", typ)
	}
}

func openMqWithType(name string, flags int, typ string) (mq.Messenger, error) {
	switch typ {
	case "default":
		return mq.Open(name, flags)
	case "sysv":
		return mq.OpenSystemVMessageQueue(name, flags)
	case "fast":
		return mq.OpenFastMq(name, flags)
	case "linux":
		return mq.OpenLinuxMessageQueue(name, flags)
	default:
		return nil, fmt.Errorf("unknown mq type %q", typ)
	}
}

func destroyMqWithType(name, typ string) error {
	switch typ {
	case "default":
		return mq.Destroy(name)
	case "sysv":
		return mq.DestroySystemVMessageQueue(name)
	case "fast":
		return mq.DestroyFastMq(name)
	case "linux":
		return mq.DestroyLinuxMessageQueue(name)
	default:
		return fmt.Errorf("unknown mq type %q", typ)
	}
}

func notifywait(name string, timeout int, typ string) error {
	if typ != "linux" {
		return fmt.Errorf("notifywait is supported for 'linux' mq, not '%s'", typ)
	}
	mq, err := mq.OpenLinuxMessageQueue(name, os.O_RDWR)
	if err != nil {
		return err
	}
	defer mq.Close()
	notifyChan := make(chan int, 1)
	if err = mq.Notify(notifyChan); err != nil {
		return err
	}
	var timeChan <-chan time.Time
	if timeout > 0 {
		timeChan = time.After(time.Duration(timeout) * time.Millisecond)
	}
	select {
	case id := <-notifyChan:
		if id != mq.ID() {
			return fmt.Errorf("expected mq with id %q, got with %q", mq.ID(), id)
		}
	case <-timeChan:
		return fmt.Errorf("operation timeout")
	}
	return nil
}
