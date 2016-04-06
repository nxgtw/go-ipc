// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd netbsd openbsd solaris

package main

import (
	"fmt"
	"os"

	"bitbucket.org/avd/go-ipc/mq"
)

func createMqWithType(name string, perm os.FileMode, typ, opt string) (mq.Messenger, error) {
	switch typ {
	case "default":
		return mq.CreateMQ(name, perm)
	case "sysv":
		return mq.CreateSystemVMessageQueue(name, perm)
	default:
		return nil, fmt.Errorf("unknown mq type %q", typ)
	}
}

func openMqWithType(name string, flags int, typ string) (mq.Messenger, error) {
	switch typ {
	case "default":
		return mq.OpenMQ(name, flags)
	case "sysv":
		return mq.OpenSystemVMessageQueue(name, flags)
	default:
		return nil, fmt.Errorf("unknown mq type %q", typ)
	}
}

func destroyMqWithType(name, typ string) error {
	switch typ {
	case "default":
		return mq.DestroyMQ(name)
	case "sysv":
		return mq.DestroySystemVMessageQueue(name)
	default:
		return fmt.Errorf("unknown mq type %q", typ)
	}
}

func notifywait(name string, timeout int, typ string) error {
	return fmt.Errorf("notifywait is not supported on current platform")
}
