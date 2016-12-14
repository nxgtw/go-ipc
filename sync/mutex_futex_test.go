// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux freebsd

package sync

import (
	"os"
	"testing"
)

func BenchmarkFutexMutex(b *testing.B) {
	benchmarkLocker(b, func(name string, mode int, perm os.FileMode) (IPCLocker, error) {
		return NewFutexMutex(name, mode, perm)
	}, func(name string) error {
		return DestroyFutexMutex(name)
	})
}
