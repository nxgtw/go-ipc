// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
)

func rwMutexCtor(name string, flag int, perm os.FileMode) (IPCLocker, error) {
	return NewRWMutex(name, flag, perm)
}

func rwRMutexCtor(name string, flag int, perm os.FileMode) (IPCLocker, error) {
	locker, err := NewRWMutex(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return locker.RLocker(), nil
}

func rwMutexDtor(name string) error {
	return DestroyRWMutex(name)
}

func TestRWMutexOpenMode(t *testing.T) {
	testLockerOpenMode(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexOpenMode2(t *testing.T) {
	testLockerOpenMode2(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexOpenMode3(t *testing.T) {
	testLockerOpenMode3(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexOpenMode4(t *testing.T) {
	testLockerOpenMode4(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexOpenMode5(t *testing.T) {
	testLockerOpenMode5(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexLock(t *testing.T) {
	testLockerLock(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexMemory(t *testing.T) {
	testLockerMemory(t, "rw", false, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexMemory2(t *testing.T) {
	testLockerMemory(t, "rw", true, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexValueInc(t *testing.T) {
	testLockerValueInc(t, "rw", rwMutexCtor, rwMutexDtor)
}

func TestRWMutexPanicsOnDoubleUnlock(t *testing.T) {
	testLockerTwiceUnlock(t, rwMutexCtor, rwMutexDtor)
}

func TestRWMutexPanicsOnDoubleRUnlock(t *testing.T) {
	testLockerTwiceUnlock(t, rwRMutexCtor, rwMutexDtor)
}

func ExampleRWMutex() {
	const (
		writers = 4
		readers = 10
	)
	DestroyRWMutex("rw")
	m, err := NewRWMutex("rw", os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		panic(err)
	}
	// we create a shared array of consistently increasing ints for reading and wriring.
	sharedData := make([]int, 128)
	for i := range sharedData {
		sharedData[i] = i
	}
	var wg sync.WaitGroup
	wg.Add(writers + readers)
	// writers will update the data.
	for i := 0; i < writers; i++ {
		go func() {
			defer wg.Done()
			start := rand.Intn(1024)
			m.Lock()
			for i := range sharedData {
				sharedData[i] = i + start
			}
			m.Unlock()
		}()
	}
	// readers will check the data.
	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			m.RLock()
			for i := 1; i < len(sharedData); i++ {
				if sharedData[i] != sharedData[i-1]+1 {
					panic("bad data")
				}
			}
			m.RUnlock()
		}()
	}
	wg.Wait()
	fmt.Println("done")
	// Output:
	// done
}
