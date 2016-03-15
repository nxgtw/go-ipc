// Copyright 2016 Aleksandr Demakin. All rights reserved.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/avd/go-ipc"
)

func createMqWithType(name string, perm os.FileMode, typ, opt string) (ipc.Messenger, error) {
	switch typ {
	case "default":
		return ipc.CreateMQ(name, perm)
	case "sysv":
		return ipc.CreateSystemVMessageQueue(name, perm)
	case "linux":
		mqSize, msgSize := ipc.DefaultLinuxMqMaxSize, ipc.DefaultLinuxMqMaxMessageSize
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
		return ipc.CreateLinuxMessageQueue(name, perm, mqSize, msgSize)
	default:
		return nil, fmt.Errorf("unknown mq type %q", typ)
	}
}

func openMqWithType(name string, flags int, typ string) (ipc.Messenger, error) {
	switch typ {
	case "default":
		return ipc.OpenMQ(name, flags)
	case "sysv":
		return ipc.OpenSystemVMessageQueue(name, flags)
	case "linux":
		return ipc.OpenLinuxMessageQueue(name, flags)
	default:
		return nil, fmt.Errorf("unknown mq type %q", typ)
	}
}

func destroyMqWithType(name, typ string) error {
	switch typ {
	case "default":
		return ipc.DestroyMQ(name)
	case "sysv":
		return ipc.DestroySystemVMessageQueue(name)
	case "linux":
		return ipc.DestroyLinuxMessageQueue(name)
	default:
		return fmt.Errorf("unknown mq type %q", typ)
	}
}

func notifywait(name string, timeout int, typ string) error {
	if typ != "linux" {
		return fmt.Errorf("notifywait is supported for 'linux' mq, not '%s'", typ)
	}
	mq, err := ipc.OpenLinuxMessageQueue(name, ipc.O_READWRITE)
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
