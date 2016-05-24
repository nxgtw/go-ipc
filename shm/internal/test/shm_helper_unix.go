// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux darwin freebsd

package main

import (
	"fmt"
	"os"

	"bitbucket.org/avd/go-ipc/shm"
)

func newShmObject(name string, mode int, perm os.FileMode, typ string) (*shm.MemoryObject, error) {
	switch typ {
	case "default", "":
		return shm.NewMemoryObject(name, mode, perm)
	default:
		return nil, fmt.Errorf("unknown object type '%s'", typ)
	}
}

func destroyShmObject(name string, typ string) error {
	switch typ {
	case "default", "":
		return shm.DestroyMemoryObject(name)
	default:
		return fmt.Errorf("unknown object type '%s'", typ)
	}
}
