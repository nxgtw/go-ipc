// Copyright 2016 Aleksandr Demakin. All rights reserved.

//+build linux freebsd windows darwin

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/sync"
)

var (
	timeout = flag.Int("timeout", -1, "timeout for condwait. in ms.")
	fail    = flag.Bool("fail", false, "operation must fail")
)

const usage = `  test program for condvars.
available commands:
  wait cond_name locker_name
  signal cond_name
  broadcast cond_name
`

func makeCond(condName, lockerName string) (cond *sync.Cond, l sync.IPCLocker, err error) {
	l, err = sync.NewMutex(lockerName, 0, 0666)
	if err != nil {
		return
	}
	cond, err = sync.NewCond(condName, 0, 0666, l)
	return
}

func wait() error {
	if flag.NArg() != 3 {
		return fmt.Errorf("wait: must provide cond and locker name only")
	}
	cond, l, err := makeCond(flag.Arg(1), flag.Arg(2))
	if err != nil {
		return err
	}
	l.Lock()
	if *timeout < 0 {
		cond.Wait()
	} else {
		ok := cond.WaitTimeout(time.Duration(*timeout) * time.Millisecond)
		if ok != !*fail {
			return fmt.Errorf("WaitTimeout returned %v, but expected %v", ok, !*fail)
		}
	}
	return nil
}

func signal() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("signal: must provide cond name only")
	}
	condName := flag.Arg(1)
	cond, err := sync.NewCond(condName, 0, 0666, nil)
	if err != nil {
		return nil
	}
	cond.Signal()
	return nil
}

func broadcast() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("broadcast: must provide cond name only")
	}
	condName := flag.Arg(1)
	cond, err := sync.NewCond(condName, 0, 0666, nil)
	if err != nil {
		return nil
	}
	cond.Broadcast()
	return nil
}

func runCommand() error {
	command := flag.Arg(0)
	switch command {
	case "wait":
		return wait()
	case "signal":
		return signal()
	case "broadcast":
		return broadcast()
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
