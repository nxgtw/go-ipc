// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"strconv"

	"bitbucket.org/avd/go-ipc/internal/test"
)

const (
	shmProgName  = "./internal/test/shared_memory/main.go"
	fifoProgName = "./internal/test/fifo/main.go"
	mqProgName   = "./internal/test/mq/main.go"
)

// Shared memory test program

func argsForShmCreateCommand(name string, size int64) []string {
	return []string{shmProgName, "-object=" + name, "create", fmt.Sprintf("%d", size)}
}

func argsForShmDestroyCommand(name string) []string {
	return []string{shmProgName, "-object=" + name, "destroy"}
}

func argsForShmReadCommand(name string, offset int64, lenght int) []string {
	return []string{shmProgName, "-object=" + name, "read", fmt.Sprintf("%d", offset), fmt.Sprintf("%d", lenght)}
}

func argsForShmTestCommand(name string, offset int64, data []byte) []string {
	strBytes := ipc_test.BytesToString(data)
	return []string{shmProgName, "-object=" + name, "test", fmt.Sprintf("%d", offset), strBytes}
}

func argsForShmWriteCommand(name string, offset int64, data []byte) []string {
	strBytes := ipc_test.BytesToString(data)
	return []string{shmProgName, "-object=" + name, "write", fmt.Sprintf("%d", offset), strBytes}
}

// FIFO memory test program

func argsForFifoCreateCommand(name string) []string {
	return []string{fifoProgName, "-object=" + name, "create"}
}

func argsForFifoDestroyCommand(name string) []string {
	return []string{fifoProgName, "-object=" + name, "destroy"}
}

func argsForFifoReadCommand(name string, nonblock bool, lenght int) []string {
	return []string{fifoProgName, "-object=" + name, "-nonblock=" + boolStr(nonblock), "read", fmt.Sprintf("%d", lenght)}
}

func argsForFifoTestCommand(name string, nonblock bool, data []byte) []string {
	strBytes := ipc_test.BytesToString(data)
	return []string{fifoProgName, "-object=" + name, "-nonblock=" + boolStr(nonblock), "test", strBytes}
}

func argsForFifoWriteCommand(name string, nonblock bool, data []byte) []string {
	strBytes := ipc_test.BytesToString(data)
	return []string{fifoProgName, "-object=" + name, "-nonblock=" + boolStr(nonblock), "write", strBytes}
}

// Mq test program

func argsForMqCreateCommand(name string, mqMaxSize, msgMazSize int) []string {
	return []string{mqProgName, "-object=" + name, "create", strconv.Itoa(mqMaxSize), strconv.Itoa(msgMazSize)}
}

func argsForMqDestroyCommand(name string) []string {
	return []string{mqProgName, "-object=" + name, "destroy"}
}

func argsForMqSendCommand(name string, timeout, prio int, data []byte) []string {
	return []string{
		mqProgName,
		"-object=" + name,
		"-prio=" + strconv.Itoa(prio),
		"-timeout=" + strconv.Itoa(timeout),
		"send",
		ipc_test.BytesToString(data),
	}
}

func argsForMqTestCommand(name string, timeout, prio int, data []byte) []string {
	return []string{
		mqProgName,
		"-object=" + name,
		"-prio=" + strconv.Itoa(prio),
		"-timeout=" + strconv.Itoa(timeout),
		"test",
		ipc_test.BytesToString(data),
	}
}

func argsForMqNotifyWaitCommand(name string, timeout int) []string {
	return []string{
		mqProgName,
		"-object=" + name,
		"-timeout=" + strconv.Itoa(timeout),
		"notifywait",
	}
}

func boolStr(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
