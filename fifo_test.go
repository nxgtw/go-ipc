// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFifoCreate(t *testing.T) {
	if !assert.NoError(t, DestroyFifo("go-fifo-test")) {
		return
	}
	fifo, err := NewFifo("go-fifo-test", O_READ_ONLY|O_NONBLOCK, 0666)
	if assert.NoError(t, err) {
		assert.NoError(t, fifo.Destroy())
	}
}

func TestFifoCreateAbsPath(t *testing.T) {
	if !assert.NoError(t, DestroyFifo("/tmp/go-fifo-test")) {
		return
	}
	fifo, err := NewFifo("/tmp/go-fifo-test", O_READ_ONLY|O_NONBLOCK, 0666)
	if assert.NoError(t, err) {
		assert.NoError(t, fifo.Destroy())
	}
}
