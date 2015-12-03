// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"os/exec"
	"syscall"

	"bitbucket.org/avd/go-ipc/test"
)

const (
	shmProgName = "./test/shared_memory/main.go"
)

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

func runTestShmProg(args []string) (output string, err error) {
	args = append([]string{"run"}, args...)
	cmd := exec.Command("go", args...)
	var out []byte
	out, err = cmd.CombinedOutput()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				err = fmt.Errorf("%v, status code = %d", err, status)
			}
		}
	} else {
		if cmd.ProcessState.Success() {
			output = string(out)
		} else {
			err = fmt.Errorf("process has exited with an error")
		}
	}
	return
}
