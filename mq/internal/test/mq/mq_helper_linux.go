// Copyright 2016 Aleksandr Demakin. All rights reserved.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/mq"
)

func createMqWithType(name string, perm os.FileMode, typ, opt string) (mq.Messenger, error) {
	switch typ {
	case "default":
		return mq.New(name, perm)
	case "sysv":
		return mq.CreateSystemVMessageQueue(name, perm)
	case "linux":
		mqSize, msgSize := mq.DefaultLinuxMqMaxSize, mq.DefaultLinuxMqMaxMessageSize
		if len(opt) > 0 {
			var err error
			parts := strings.Split(opt, ",")
			mqSize, err = strconv.Atoi(parts[0])
			if err != nil {
				return nil, err
			}
			if len(parts) > 1 {
				msgSize, err = strconv.Atoi(parts[1])
				if err != nil {
					return nil, err
				}
			}
		}
		return mq.CreateLinuxMessageQueue(name, perm, mqSize, msgSize)
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
	mq, err := mq.OpenLinuxMessageQueue(name, ipc.O_READWRITE)
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
