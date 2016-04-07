// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build freebsd

package mq

func init() {
	// values from http://fxr.watson.org/fxr/source/kern/syscalls.master
	sysMsgCtl = 224
	sysMsgGet = 225
	sysMsgSnd = 226
	sysMsgRcv = 227
}
