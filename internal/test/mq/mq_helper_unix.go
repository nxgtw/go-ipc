// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd netbsd openbsd solaris

package main

import (
	"fmt"
	"os"

	"bitbucket.org/avd/go-ipc"
)

func createMqWithType(name string, perm os.FileMode, typ, opt string) (ipc.Messenger, error) {
	switch typ {
	case "default":
		return ipc.CreateMQ(name, perm)
	case "sysv":
		return ipc.CreateSystemVMessageQueue(name, perm)
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
	default:
		return fmt.Errorf("unknown mq type %q", typ)
	}
}

func notifywait(name string, timeout int, typ string) error {
	return fmt.Errorf("notifywait is not supported on current platform")
}
