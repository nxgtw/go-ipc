// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
)

const (
	SHM_CMD_CREATE  = "create"
	SHM_CMD_DESTROY = "destroy"
	SHM_CMD_READ    = "read"
	SHM_CMD_TEST    = "test"
	SHM_CMD_WRITE   = "write"
)

func argsForShmCreateCommand(name string, size int64) []string {
	return []string{"-object=" + name, "create", fmt.Sprintf("%d", size)}
}

func argsForShmDestroyCommand(name string) []string {
	return []string{"-object=" + name, "destroy"}
}

func argsForShmReadCommand(name string, offset int64, lenght int) []string {
	return []string{"-object=" + name, "read", fmt.Sprintf("%d", offset), fmt.Sprintf("%d", lenght)}
}

func argsForShmTestCommand(name string, offset int64, data []byte) []string {
	strBytes := byteSliceToString(data)
	return []string{"-object=" + name, "test", fmt.Sprintf("%d", offset), strBytes}
}

func argsForShmWriteCommand(name string, offset int64, data []byte) []string {
	strBytes := byteSliceToString(data)
	return []string{"-object=" + name, "write", fmt.Sprintf("%d", offset), strBytes}
}

func byteSliceToString(data []byte) string {
	buffer := bytes.NewBuffer(nil)
	for _, value := range data {
		if value < 16 {
			buffer.WriteString(fmt.Sprint("0"))
		}
		buffer.WriteString(fmt.Sprintf("%X", value))
	}
	return buffer.String()
}

func runTestShmProg(args []string) (string, error) {
	args = append([]string{"run", "./test/shared_memory_test_prog.go"}, args...)
	cmd := exec.Command("go", args...)
	var err error
	if err = cmd.Start(); err != nil {
		return "", err
	}
	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				err = fmt.Errorf("%v, status code = %d", err, status)
			}
		}
	} else {
		if cmd.ProcessState.Success() {
			if out, err := cmd.CombinedOutput(); err == nil {
				return string(out), nil
			}
		} else {
			err = fmt.Errorf("process has exited with an error")
		}
	}
	return "", err
}
