// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux freebsd

package sync

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkFutexMutex(b *testing.B) {
	benchmarkLocker(b, func(name string, mode int, perm os.FileMode) (IPCLocker, error) {
		return NewFutexMutex(name, mode, perm)
	}, func(name string) error {
		return DestroyFutexMutex(name)
	})
}

func BenchmarkFutexMutexAsRW(b *testing.B) {
	a := assert.New(b)
	DestroyFutexMutex(testLockerName)
	m, err := NewFutexMutex(testLockerName, os.O_CREATE|os.O_EXCL, 0666)
	if !a.NoError(err) {
		return
	}
	defer m.Close()
	benchmarkRWLocker(b, m, m)
}
