// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux,amd64 darwin

package sync

import "golang.org/x/sys/unix"

func init() {
	sysSemGet = unix.SYS_SEMGET
	sysSemCtl = unix.SYS_SEMCTL
	sysSemOp = unix.SYS_SEMOP
}
