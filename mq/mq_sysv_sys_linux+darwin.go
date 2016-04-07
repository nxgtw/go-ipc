// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux,amd64 darwin

package mq

import "golang.org/x/sys/unix"

func init() {
	sysMsgCtl = unix.SYS_MSGCTL
	sysMsgGet = unix.SYS_MSGGET
	sysMsgRcv = unix.SYS_MSGRCV
	sysMsgSnd = unix.SYS_MSGSND
}
