// Copyright 2016 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/sync"
)

var (
	timeout = flag.Int("timeout", -1, "timeout for wait, in ms.")
	fail    = flag.Bool("fail", false, "operation must fail")
)

const usage = `  test program for events.
available commands:
  wait event_name
  set event_name
`

func wait() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("wait: must provide event name only")
	}
	ev, err := sync.NewEvent(flag.Arg(1), 0, 0666, false)
	if err != nil {
		return err
	}
	if *timeout < 0 {
		ev.Wait()
	} else {
		ok := ev.WaitTimeout(time.Duration(*timeout) * time.Millisecond)
		if ok != !*fail {
			if !ok {
				return fmt.Errorf("timeout exceeded")
			}
			return fmt.Errorf("timeout passed")
		}
	}
	if err = ev.Close(); err != nil {
		return err
	}
	return nil
}

func set() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("signal: must provide event name only")
	}
	ev, err := sync.NewEvent(flag.Arg(1), 0, 0666, false)
	if err != nil {
		return nil
	}
	ev.Set()
	if err = ev.Close(); err != nil {
		return err
	}
	return nil
}

func runCommand() error {
	command := flag.Arg(0)
	switch command {
	case "wait":
		return wait()
	case "set":
		return set()
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
