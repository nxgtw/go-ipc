// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLockRwMutex(t *testing.T) {
	mut, err := NewRwMutex("go-rwm-test", O_OPEN_OR_CREATE|O_READWRITE, 0666)
	if !assert.NoError(t, err) || !assert.NotNil(t, mut) {
		return
	}
	defer mut.Destroy()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			mut.Lock()
			wg.Done()
			mut.Unlock()
		}()
	}
	wg.Wait()
}
