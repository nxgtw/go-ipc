// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mq

import (
	"math/rand"
	"os"
	"time"
)

func ExampleMessenger() {
	mq, err := New("mq", os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		panic("new queue")
	}
	defer mq.Close()
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	go func() {
		if err := mq.Send(data); err != nil {
			panic("send")
		}
	}()
	mq2, err := Open("mq", 0)
	if err != nil {
		panic("open")
	}
	defer mq2.Close()
	received := make([]byte, len(data))
	l, err := mq2.Receive(data)
	if err != nil {
		panic("receive")
	}
	if l != len(data) {
		panic("wrong len")
	}
	for i, b := range received {
		if b != data[i] {
			panic("wrong data")
		}
	}
}

func ExampleTimedMessenger() {
	Destroy("mq")
	mq, err := New("mq", os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		panic("new queue")
	}
	defer mq.Close()
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	go func() {
		// send after [0..500] ms delay.
		time.Sleep(time.Duration((rand.Int() % 6)) * time.Millisecond * 100)
		if err := mq.Send(data); err != nil {
			panic("send")
		}
	}()
	mq2, err := Open("mq", 0)
	if err != nil {
		panic("open")
	}
	defer mq2.Close()
	// not all implementations support timed send/receive.
	tmq, ok := mq2.(TimedMessenger)
	if !ok {
		panic("not a timed messenger")
	}
	received := make([]byte, len(data))
	// depending on send delay we either get a timeout error, or receive the data.
	l, err := tmq.ReceiveTimeout(received, 500*time.Millisecond)
	if err != nil {
		if !IsTemporary(err) {
			panic(err)
		} else { // handle timeout.
			return
		}
	}
	if l != len(data) {
		panic("wrong len")
	}
	for i, b := range received {
		if b != data[i] {
			panic("wrong data")
		}
	}
}

func ExamplePrioMessenger() {
	Destroy("mq")
	mq, err := New("mq", os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		panic("new queue")
	}
	defer mq.Close()
	// not all implementations support prioritized send/receive.
	tmq, ok := mq.(PriorityMessenger)
	if !ok {
		panic("not a prio messenger")
	}
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	go func() {
		if err := tmq.SendPriority(data, 0); err != nil {
			panic("send")
		}
		if err := tmq.SendPriority(data, 1); err != nil {
			panic("send")
		}
	}()
	mq2, err := Open("mq", 0)
	if err != nil {
		panic("open")
	}
	defer mq2.Close()
	tmq2, ok := mq2.(PriorityMessenger)
	if !ok {
		panic("not a prio messenger")
	}
	received := make([]byte, len(data))
	_, prio, err := tmq2.ReceivePriority(received)
	if err != nil || prio != 1 {
		panic("receive")
	}
	_, prio, err = tmq2.ReceivePriority(received)
	if err != nil || prio != 0 {
		panic("receive")
	}
}
