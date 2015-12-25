// Copyright 2015 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	ipc "bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/test"
)

var (
	objName = flag.String("object", "", "mq name")
	timeout = flag.Int("timeout", -1, "timeout for send/receive/notify wait. in ms.")
	prio    = flag.Int("prio", 0, "message prioroty")
)

const usage = `  test program for message queues.
available commands:
  create {max_size} {max_msg_len}
  destroy
  test {expected values byte array}
  send {values byte array}
  notifywait
byte array should be passed as a continuous string of 2-symbol hex byte values like '01020A'
`

func create() error {
	if flag.NArg() != 3 {
		return fmt.Errorf("create: must provide exactly two argument")
	}
	maxSize, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	maxMsgLen, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return err
	}
	mq, err := ipc.CreateMessageQueue(*objName, false, 0666, maxSize, maxMsgLen)
	if err != nil {
		mq.Close()
	}
	return err
}

func destroy() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("destroy: must not provide any arguments")
	}
	return ipc.DestroyMessageQueue(*objName)
}

func test() error {
	if flag.NArg() != 2 {
		return fmt.Errorf("test: must provide exactly one argument")
	}
	mq, err := ipc.OpenMessageQueue(*objName, ipc.O_READ_ONLY)
	if err != nil {
		return err
	}
	defer mq.Close()
	expected, err := ipc_test.StringToBytes(flag.Arg(1))
	if err != nil {
		return err
	}
	received := make([]byte, len(expected))
	if *timeout >= 0 {
		err = mq.ReceiveTimeout(received, nil, time.Duration(*timeout)*time.Millisecond)
	} else {
		err = mq.Receive(received, nil)
	}
	if err != nil {
		return err
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
	mq, err := ipc.OpenMessageQueue(*objName, ipc.O_WRITE_ONLY)
	if err != nil {
		return err
	}
	defer mq.Close()
	toSend, err := ipc_test.StringToBytes(flag.Arg(1))
	if err != nil {
		return err
	}
	if *timeout >= 0 {
		err = mq.SendTimeout(toSend, *prio, time.Duration(*timeout)*time.Millisecond)
	} else {
		err = mq.Send(toSend, *prio)
	}
	return nil
}

func notifywait() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("notifywait: must not provide any arguments")
	}
	mq, err := ipc.OpenMessageQueue(*objName, ipc.O_WRITE_ONLY)
	if err != nil {
		return err
	}
	defer mq.Close()
	notifyChan := make(chan int, 1)
	if err = mq.Notify(notifyChan); err != nil {
		return err
	}
	var timeChan <-chan time.Time
	if *timeout > 0 {
		timeChan = time.After(time.Duration(*timeout) * time.Millisecond)
	}
	select {
	case id := <-notifyChan:
		if id != mq.Id() {
			return fmt.Errorf("expected mq with id %q, got with %q", mq.Id(), id)
		}
	case <-timeChan:
		return fmt.Errorf("operation timeout")
	}
	return nil
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
		return notifywait()
	default:
		return fmt.Errorf("unknown command")
	}
	return nil
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
