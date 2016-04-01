// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package common

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

const (
	IpcCreate = 00001000 /* create if key is nonexistent */
	IpcExcl   = 00002000 /* fail if key exists */
	IpcNoWait = 00004000 /* return error on wait */

	IpcRmid = 0 /* remove resource */
	IpcSet  = 1 /* set ipc_perm options */
	IpcStat = 2 /* get ipc_perm options */
	IpcInfo = 3 /* see ipcs */
)

type Key uint64

func KeyForName(name string) (Key, error) {
	name = TmpFilename(name)
	file, err := os.Create(name)
	if err != nil {
		return 0, errors.New("invalid name for key")
	}
	file.Close()
	k, err := Ftok(name)
	if err != nil {
		os.Remove(name)
		return 0, errors.New("invalid name for key")
	}
	return k, nil
}

func TmpFilename(name string) string {
	return os.TempDir() + "/" + name
}

func Ftok(name string) (Key, error) {
	var statfs unix.Stat_t
	if err := unix.Stat(name, &statfs); err != nil {
		return Key(0), err
	}
	return Key(statfs.Ino&0xFFFF | ((statfs.Dev & 0xFF) << 16)), nil
}

func TimeoutToTimeSpec(timeout time.Duration) *unix.Timespec {
	if timeout >= 0 {
		ts := unix.NsecToTimespec(time.Now().Add(timeout).UnixNano())
		return &ts
	}
	return nil
}
