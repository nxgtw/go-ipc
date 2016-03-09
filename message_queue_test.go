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
	_, err := ctor(testMqName, 0666)
	if a.NoError(err) {
		if dtor != nil {
			a.NoError(dtor(testMqName))
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
	a.NoError(dtor(testMqName))
	_, err := ctor(testMqName, 0666)
	a.NoError(err)
	a.NoError(dtor(testMqName))
	_, err = opener(testMqName, O_READ_ONLY)
	a.Error(err)
}
