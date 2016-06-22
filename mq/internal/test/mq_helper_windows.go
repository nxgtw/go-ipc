// Copyright 2016 Aleksandr Demakin. All rights reserved.

package main

import (
	"fmt"
	"os"

	"bitbucket.org/avd/go-ipc/mq"
)

func createMqWithType(name string, perm os.FileMode, typ, opt string) (mq.Messenger, error) {
	switch typ {
	case "default", "fast":
		mqSize, msgSize := mq.DefaultLinuxMqMaxSize, mq.DefaultLinuxMqMessageSize
		if first, second, err := parseTwoInts(opt); err == nil {
			mqSize, msgSize = first, second
		}
		return mq.CreateFastMq(name, 0, perm, mqSize, msgSize)
	default:
		return nil, fmt.Errorf("unknown mq type %q", typ)
	}
}

func openMqWithType(name string, flags int, typ string) (mq.Messenger, error) {
	switch typ {
	case "default", "fast":
		return mq.OpenFastMq(name, flags)
	default:
		return nil, fmt.Errorf("unknown mq type %q", typ)
	}
}

func destroyMqWithType(name, typ string) error {
	switch typ {
	case "default", "fast":
		return mq.DestroyFastMq(name)
	default:
		return fmt.Errorf("unknown mq type %q", typ)
	}
}

func notifywait(name string, timeout int, typ string) error {
	return fmt.Errorf("notifywait is not supported on current platform")
}
