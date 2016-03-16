// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc_testing

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// TestAppResult is a result of a 'go run' program launch
type TestAppResult struct {
	Output string
	Err    error
}

// StringToBytes takes an input string in a 2-hex-symbol per byte format
// and returns corresponding byte array.
// Input must not contain any symbols except [A-F0-9]
func StringToBytes(input string) ([]byte, error) {
	if len(input)%2 != 0 {
		return nil, fmt.Errorf("invalid byte array len")
	}
	var err error
	var b byte
	buff := bytes.NewBuffer(nil)
	for err == nil {
		if len(input) < 2 {
			err = io.EOF
		} else if _, err = fmt.Sscanf(input[:2], "%X", &b); err == nil {
			buff.WriteByte(b)
			input = input[2:]
		}
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buff.Bytes(), nil
}

// BytesToString convert a byte slice into its string representation.
// Each byte is represented as a 2 upper-case letters for A-F
func BytesToString(data []byte) string {
	buff := bytes.NewBuffer(nil)
	for _, value := range data {
		if value < 16 { // force leading 0 for 1-digit values
			buff.WriteString("0")
		}
		buff.WriteString(fmt.Sprintf("%X", value))
	}
	return buff.String()
}

// launch helpers

func startTestApp(args []string, killChan <-chan bool) (*exec.Cmd, *bytes.Buffer, error) {
	args = append([]string{"run"}, args...)
	cmd := exec.Command("go", args...)
	buff := bytes.NewBuffer(nil)
	cmd.Stderr = buff
	cmd.Stdout = buff
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	if killChan != nil {
		go func() {
			if kill, ok := <-killChan; kill && ok {
				if cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
					cmd.Process.Kill()
				}
			}
		}()
	}
	fmt.Printf("started new process [%d]\n", cmd.Process.Pid)
	return cmd, buff, nil
}

func waitForCommand(cmd *exec.Cmd, buff *bytes.Buffer) (result TestAppResult) {
	if result.Err = cmd.Wait(); result.Err != nil {
		if exiterr, ok := result.Err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				result.Err = fmt.Errorf("%v, status code = %d", result.Err, status)
			}
		}
	} else {
		if !cmd.ProcessState.Success() {
			result.Err = fmt.Errorf("process has exited with an error")
		}
	}
	result.Output = buff.String()
	return
}

// RunTestApp starts a go program via 'go run'.
// To kill the process, send to killChan
func RunTestApp(args []string, killChan <-chan bool) (result TestAppResult) {
	if cmd, buff, err := startTestApp(args, killChan); err == nil {
		result = waitForCommand(cmd, buff)
	} else {
		result.Err = err
	}
	return
}

// RunTestAppAsync starts a go program via 'go run' and returns immediately.
// To kill the process, send to killChan.
// To wait for the program to finish, receive on TestAppResult chan.
func RunTestAppAsync(args []string, killChan <-chan bool) <-chan TestAppResult {
	ch := make(chan TestAppResult, 1)
	if cmd, buff, err := startTestApp(args, killChan); err != nil {
		ch <- TestAppResult{Err: err}
	} else {
		go func() {
			ch <- waitForCommand(cmd, buff)
		}()
	}
	return ch
}

// WaitForFunc calls f asynchronously leaving it some time to finish.
// It returns true, if f completed.
func WaitForFunc(f func(), d time.Duration) bool {
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

// WaitForAppResultChan waits for a value from ch with a timeout
func WaitForAppResultChan(ch <-chan TestAppResult, d time.Duration) (TestAppResult, bool) {
	select {
	case value := <-ch:
		return value, true
	case <-time.After(d):
		return TestAppResult{}, false
	}
}

// LocatePackageFiles returns a slice of all the buildable source files in the given directory
func LocatePackageFiles(path string) ([]string, error) {
	args := []string{"list", "-f", "{{.GoFiles}}", path}
	cmd := exec.Command("go", args...)
	buff := bytes.NewBuffer(nil)
	cmd.Stderr = buff
	cmd.Stdout = buff
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	result := waitForCommand(cmd, buff)
	if result.Err != nil {
		return nil, result.Err
	}
	return buildFilesFromOutput(result.Output), nil
}

func buildFilesFromOutput(output string) []string {
	output = strings.TrimSpace(output)
	output = strings.Trim(output, "[]")
	parts := strings.Split(output, " ")
	for i := 0; i < len(parts); i++ {
		if !strings.HasSuffix(parts[i], ".go") {
			for j := i + 1; j < len(parts); j++ {
				needBrake := strings.HasSuffix(parts[j], ".go")
				parts[i] += parts[j]
				parts[j] = ""
				if needBrake {
					break
				}
			}
		}
	}
	for i := len(parts) - 1; i >= 0 && len(parts[i]) == 0; i-- {
		parts = parts[:i]
	}
	return parts
}
