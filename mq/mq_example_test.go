// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mq

import "os"

func ExampleMessenger() {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	mq, err := New("mq", os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return
	}
	defer mq.Close()
	go func() {
		if err := mq.Send(data); err != nil {
			return
		}
	}()
	mq2, err := Open("mq", 0)
	if err != nil {
		return
	}
	defer mq2.Close()
	received := make([]byte, len(data))
	if err := mq2.Receive(data); err != nil {
		return
	}
	for i, b := range received {
		if b != data[i] {
			panic("wrong data")
		}
	}
}

func ExampleTimedMessenger() {
	mq, err := New("mq", os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return
	}
	defer mq.Close()
	received := make([]byte, 8)
	if err := mq.Receive(received); err != nil {
		return
	}
}
