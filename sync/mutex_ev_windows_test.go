// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"testing"
)

func BenchmarkEventMutex(b *testing.B) {
	benchmarkLocker(b, func(name string, mode int, perm os.FileMode) (IPCLocker, error) {
		return NewEventMutex(name, mode, perm)
	}, func(name string) error {
		return DestroyEventMutex(name)
	})
}
