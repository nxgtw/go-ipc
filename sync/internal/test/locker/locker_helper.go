// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build !linux,!freebsd

package main

import (
	"fmt"
	"sync"

	ipc_sync "bitbucket.org/avd/go-ipc/sync"
)

func createLocker(typ, name string, mode int) (locker sync.Locker, err error) {
	switch typ {
	case "m":
		locker, err = ipc_sync.NewMutex(name, mode, 0666)
	case "spin":
		locker, err = ipc_sync.NewSpinMutex(name, mode, 0666)
	case "rw":
		locker, err = ipc_sync.NewRWMutex(name, mode, 0666)
	default:
		err = fmt.Errorf("unknown object type %q", typ)
	}
	return
}

func destroyLocker(typ, name string) error {
	switch typ {
	case "m":
		return ipc_sync.DestroyMutex(name)
	case "spin":
		return ipc_sync.DestroySpinMutex(name)
	case "rw":
		return ipc_sync.DestroyRWMutex(name)
	default:
		return fmt.Errorf("unknown object type %q", typ)
	}
}
