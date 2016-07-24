// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux

package sync

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testCondName    = "ipccond"
	testCondMutName = "ipccondmut"
)

func TestCondWait(t *testing.T) {
	a := assert.New(t)
	a.NoError(DestroyCond(testCondName))
	a.NoError(DestroyMutex(testCondMutName))
	l, err := NewMutex(testCondMutName, os.O_CREATE|os.O_EXCL, 0666)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(DestroyMutex(testCondMutName))
	}()
	cond, err := NewCond(testCondName, os.O_CREATE|os.O_EXCL, 0666, l)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(DestroyCond(testCondName))
	}()
	l.Lock()
	go func() {
		time.Sleep(time.Millisecond * 50)
		cond.Signal()
	}()
	cond.Wait()
}

func TestCondWaitTimeout(t *testing.T) {
	a := assert.New(t)
	a.NoError(DestroyCond(testCondName))
	a.NoError(DestroyMutex(testCondMutName))
	l, err := NewMutex(testCondMutName, os.O_CREATE|os.O_EXCL, 0666)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(DestroyMutex(testCondMutName))
	}()
	cond, err := NewCond(testCondName, os.O_CREATE|os.O_EXCL, 0666, l)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(DestroyCond(testCondName))
	}()
	l.Lock()
	cond.WaitTimeout(time.Millisecond * 50)
}
