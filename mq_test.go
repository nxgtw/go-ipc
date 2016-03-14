// Copyright 2016 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testMqName = "go-ipc.mq"
)

type mqCtor func(name string, perm os.FileMode) (Messenger, error)
type mqOpener func(name string, flags int) (Messenger, error)
type mqDtor func(name string) error

func testCreateMq(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0666)
	if a.NoError(err) {
		if dtor != nil {
			a.NoError(dtor(testMqName))
		} else {
			a.NoError(mq.Close())
		}
	}
}

func testCreateMqExcl(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0666)
	if !a.NoError(err) || !a.NotNil(mq) {
		return
	}
	_, err = ctor(testMqName, 0666)
	a.Error(err)
	if d, ok := mq.(Destroyer); ok {
		a.NoError(d.Destroy())
	} else {
		a.NoError(mq.Close())
	}
}

func testCreateMqInvalidPerm(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	_, err := ctor(testMqName, 0777)
	a.Error(err)
}

func testOpenMq(t *testing.T, ctor mqCtor, opener mqOpener, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0666)
	if !a.NoError(err) {
		return
	}
	if dtor != nil {
		a.NoError(dtor(testMqName))
	} else {
		a.NoError(mq.Close())
	}
	_, err = opener(testMqName, O_READ_ONLY)
	a.Error(err)
}

func testMqSendInvalidType(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0666)
	if !a.NoError(err) {
		return
	}
	defer func() {
		if dtor != nil {
			a.NoError(dtor(testMqName))
		} else {
			a.NoError(mq.Close())
		}
	}()
	assert.Error(t, mq.Send("string"))
	structWithString := struct{ a string }{"string"}
	assert.Error(t, mq.Send(structWithString))
	var slslByte [][]byte
	assert.Error(t, mq.Send(slslByte))
}

func testMqSendIntSameProcess(t *testing.T, ctor mqCtor, opener mqOpener, dtor mqDtor) {
	var message uint64 = 0xDEAFBEEFDEAFBEEF
	a := assert.New(t)
	mq, err := ctor(testMqName, 0666)
	if !a.NoError(err) {
		return
	}
	defer func() {
		if dtor != nil {
			a.NoError(dtor(testMqName))
		} else {
			a.NoError(mq.Close())
		}
	}()
	go func() {
		a.NoError(mq.Send(message))
	}()
	var received uint64
	mqr, err := opener(testMqName, O_READ_ONLY)
	a.NoError(err)
	err = mqr.Receive(&received)
	a.NoError(err)
	a.Equal(message, received)
}

func testMqSendStructSameProcess(t *testing.T, ctor mqCtor, opener mqOpener, dtor mqDtor) {
	type testStruct struct {
		arr [16]int
		c   complex128
		s   struct {
			a, b byte
		}
		f float64
	}
	a := assert.New(t)
	message := testStruct{c: complex(2, -3), f: 11.22, s: struct{ a, b byte }{127, 255}}
	mq, err := ctor(testMqName, 0666)
	if !a.NoError(err) {
		return
	}
	go func() {
		a.NoError(mq.Send(message))
	}()
	received := &testStruct{}
	mqr, err := opener(testMqName, O_READ_ONLY)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(mqr.Close())
		a.NoError(dtor(testMqName))
	}()
	a.NoError(mqr.Receive(received))
	a.Equal(message, *received)
	a.NoError(mq.Close())
}

func testMqSendMessageLessThenBuffer(t *testing.T, ctor mqCtor, opener mqOpener, dtor mqDtor) {
	a := assert.New(t)
	mq, err := ctor(testMqName, 0666)
	if !a.NoError(err) {
		return
	}
	message := make([]int, 512)
	for i := range message {
		message[i] = i
	}
	go func() {
		a.NoError(mq.Send(message))
	}()
	received := make([]int, 1024)
	mqr, err := opener(testMqName, O_READ_ONLY)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(mqr.Close())
		a.NoError(dtor(testMqName))
	}()
	a.NoError(mqr.Receive(received))
	a.Equal(message, received[:512])
	a.Equal(received[512:], make([]int, 512))
	a.NoError(mq.Close())
}

func testMqSendNonBlock(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		a.NoError(dtor(testMqName))
	}()
	if blocker, ok := mq.(Blocker); ok {
		a.NoError(blocker.SetBlocking(false))
		endChan := make(chan bool, 1)
		go func() {
			for i := 0; i < 100; i++ {
				a.NoError(mq.Send(0x12345678))
			}
			endChan <- true
		}()
		select {
		case <-endChan:
		case <-time.After(time.Millisecond * 300):
			t.Errorf("send on non-blocking mq blocked")
		}
	} else {
		t.Skipf("current mq impl on %s does not implement Blocker", runtime.GOOS)
	}
}

func testMqReceiveNonBlock(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		a.NoError(dtor(testMqName))
	}()
	if blocker, ok := mq.(Blocker); ok {
		a.NoError(blocker.SetBlocking(false))
		endChan := make(chan bool, 1)
		go func() {
			var data int
			for i := 0; i < 32; i++ {
				a.Error(mq.Receive(&data))
			}
			endChan <- true
		}()
		select {
		case <-endChan:
		case <-time.After(time.Millisecond * 300):
			t.Errorf("receive on non-blocking mq blocked")
		}
	} else {
		t.Skipf("current mq impl on %s does not implement Blocker", runtime.GOOS)
	}
}
