// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"bitbucket.org/avd/go-ipc/internal/test"
)

const (
	shmProgName  = "./internal/test/shared_memory/main.go"
	fifoProgName = "./internal/test/fifo/main.go"
)

type testAppResult struct {
	output string
	err    error
}

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

func boolStr(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func runTestApp(args []string, killChan <-chan bool) (result testAppResult) {
	args = append([]string{"run"}, args...)
	cmd := exec.Command("go", args...)
	var out []byte
	if killChan != nil {
		go func() {
			if kill, ok := <-killChan; kill && ok {
				if cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
					cmd.Process.Kill()
				}
			}
		}()
	}
	out, result.err = cmd.CombinedOutput()
	if result.err != nil {
		if exiterr, ok := result.err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				result.err = fmt.Errorf("%v, status code = %d", result.err, status)
			}
		}
	} else {
		if cmd.ProcessState.Success() {
			result.output = string(out)
		} else {
			result.err = fmt.Errorf("process has exited with an error")
		}
	}
	return
}

func runTestAppAsync(args []string, killChan <-chan bool) <-chan testAppResult {
	ch := make(chan testAppResult, 1)
	go func() {
		ch <- runTestApp(args, killChan)
	}()
	return ch
}

func waitForFunc(f func(), d time.Duration) bool {
	ch := make(chan bool, 1)
	go func() {
		f()
		ch <- true
	}()
	select {
	case <-ch:
		return true
	case <-time.After(d):
		return false
	}
}

func waitForAppResultChan(ch <-chan testAppResult, d time.Duration) (testAppResult, bool) {
	select {
	case value := <-ch:
		return value, true
	case <-time.After(d):
		return testAppResult{}, false
	}
}
