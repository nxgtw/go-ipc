// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"os"

	"golang.org/x/sys/unix"
)

type Fifo struct {
	file *os.File
}

func NewFifo(name string, mode int, perm os.FileMode) (*Fifo, error) {
	if err := unix.Mkfifo(name, uint32(perm)); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(name, mode, perm)
	if err != nil {
		return nil, err
	}
	return &Fifo{file: file}, nil
}

func (f *Fifo) Read(b []byte) (n int, err error) {
	return f.file.Read(b)
}

func (f *Fifo) Write(b []byte) (n int, err error) {
	return f.file.Write(b)
}
