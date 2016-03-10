// Copyright 2016 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"os"
	"testing"

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
	var message = 0xDEADBEEF
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
	var received int
	mqr, err := opener(testMqName, O_READ_ONLY)
	a.NoError(err)
	err = mqr.Receive(&received)
	a.NoError(err)
}
