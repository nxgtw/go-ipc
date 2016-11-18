// Copyright 2015 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/avd/go-ipc/internal/test"
	"bitbucket.org/avd/go-ipc/mq"
)

var (
	objName = flag.String("object", "", "mq name")
	timeout = flag.Int("timeout", -1, "timeout for send/receive/notify wait. in ms.")
	typ     = flag.String("type", "default", "message queue type")
	options = flag.String("options", "", "a set of options for a particular mq")
)

const usage = `  test program for message queues.
available commands:
  create
  destroy
  test {expected values byte array}
  send {values byte array}
  notifywait
byte array should be passed as a continuous string of 2-symbol hex byte values like '01020A'
`

func create() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("create: must provide exactly one argument")
	}
	mq, err := createMqWithType(*objName, 0666, *typ, *options)
	if err == nil {
		mq.Close()
	}
	return err
}

func destroy() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("destroy: must not provide any arguments")
	}
	return destroyMqWithType(*objName, *typ)
}

func test() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("test: must provide exactly one argument")
	}
	msqQueue, err := openMqWithType(*objName, os.O_RDONLY, *typ)
	if err != nil {
		return err
	}
	defer msqQueue.Close()
	expected, err := testutil.StringToBytes(flag.Arg(1))
	if err != nil {
		return err
	}
	received := make([]byte, len(expected))
	var l int
	if *timeout >= 0 {
		if tm, ok := msqQueue.(mq.TimedMessenger); ok {
			l, err = tm.ReceiveTimeout(received, time.Duration(*timeout)*time.Millisecond)
		} else {
			return fmt.Errorf("selected mq implementation does not support timeouts")
		}
	} else {
		l, err = msqQueue.Receive(received)
	}
	if err != nil {
		return err
	}
	if l != len(expected) {
		return fmt.Errorf("invalid len. expected '%d', got '%d'", len(expected), l)
	}
	for i, expectedValue := range expected {
		if expectedValue != received[i] {
			return fmt.Errorf("invalid value at %d. expected '%d', got '%d'", i, expectedValue, received[i])
		}
	}
	return nil
}

func send() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("send: must provide exactly one argument")
	}
	msgQueue, err := openMqWithType(*objName, os.O_WRONLY, *typ)
	if err != nil {
		return err
	}
	defer msgQueue.Close()
	toSend, err := testutil.StringToBytes(flag.Arg(1))
	if err != nil {
		return err
	}
	if *timeout >= 0 {
		if tm, ok := msgQueue.(mq.TimedMessenger); ok {
			err = tm.SendTimeout(toSend, time.Duration(*timeout)*time.Millisecond)
		} else {
			return fmt.Errorf("selected mq implementation does not support timeouts")
		}
	} else {
		err = msgQueue.Send(toSend)
	}
	return err
}

func runCommand() error {
	command := flag.Arg(0)
	switch command {
	case "create":
		return create()
	case "destroy":
		return destroy()
	case "test":
		return test()
	case "send":
		return send()
	case "notifywait":
		if flag.NArg() != 1 {
			return fmt.Errorf("notifywait: must not provide any arguments")
		}
		return notifywait(*objName, *timeout, *typ)
	default:
		return fmt.Errorf("unknown command")
	}
}

func parseTwoInts(opt string) (first int, second int, err error) {
	if len(opt) == 0 {
		return
	}
	parts := strings.Split(opt, ",")
	first, err = strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	if len(parts) > 1 {
		second, err = strconv.Atoi(parts[1])
		if err != nil {
			return
		}
	}
	return
}

func main() {
	flag.Parse()
	if len(*objName) == 0 || flag.NArg() == 0 {
		fmt.Print(usage)
		flag.Usage()
		os.Exit(1)
	}
	if err := runCommand(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
