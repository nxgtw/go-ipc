// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package sync

const (
	cSemUndo = 0x1000
)

type sembuf struct {
	semnum uint16
	semop  int16
	semflg int16
}
