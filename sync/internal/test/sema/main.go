// Copyright 2016 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"bitbucket.org/avd/go-ipc/sync"
)

var (
	timeout = flag.Int("timeout", -1, "timeout for wait, in ms.")
	fail    = flag.Bool("fail", false, "operation must fail")
)

const usage = `  test program for semaphores.
available commands:
  wait sema_name
  signal sema_name count
`

func wait() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("wait: must provide sema name only")
	}
	s, err := sync.NewSemaphore(flag.Arg(1), 0, 0666, 0)
	if err != nil {
		return err
	}
	if *timeout < 0 {
		s.Wait()
	} else {
		ok := s.WaitTimeout(time.Duration(*timeout) * time.Millisecond)
		if ok != !*fail {
			if !ok {
				return fmt.Errorf("timeout exceeded")
			}
			return fmt.Errorf("timeout passed")
		}
	}
	return s.Close()
}

func signal() error {
	if flag.NArg() != 3 {
		return fmt.Errorf("signal: must provide sema name and count")
	}
	s, err := sync.NewSemaphore(flag.Arg(1), 0, 0666, 0)
	if err != nil {
		return err
	}
	count, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		return err
	}
	s.Signal(count)
	return s.Close()
}

func runCommand() error {
	command := flag.Arg(0)
	switch command {
	case "wait":
		return wait()
	case "signal":
		return signal()
	default:
		return fmt.Errorf("unknown command")
	}
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Print(usage)
		flag.Usage()
		os.Exit(1)
	}
	if err := runCommand(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
