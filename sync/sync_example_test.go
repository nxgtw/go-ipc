// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"math/rand"
	"os"
	"sync"
	"time"
)

func ExampleIPCLocker() {
	DestroyMutex("mut")
	mut, err := NewMutex("mut", os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		panic("new")
	}
	defer mut.Close()
	var sharedValue uint64
	var wg sync.WaitGroup
	wg.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			defer wg.Done()
			mut, err := NewMutex("mut", 0, 0)
			if err != nil {
				panic("new")
			}
			defer mut.Close()
			for i := 0; i < 1000000; i++ {
				mut.Lock()
				sharedValue++
				mut.Unlock()
			}
		}()
	}
	wg.Wait()
	if sharedValue != 8*1000000 {
		panic("invalid value ")
	}
}

func ExampleTimedIPCLocker() {
	DestroyMutex("mut")
	mut, err := NewMutex("mut", os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		panic("new")
	}
	defer mut.Close()
	tmut, ok := mut.(TimedIPCLocker)
	if !ok {
		panic("not a timed locker")
	}
	var sharedValue int
	rand.Seed(time.Now().Unix())
	go func() {
		mut, err := NewMutex("mut", 0, 0)
		if err != nil {
			panic("new")
		}
		defer mut.Close()
		mut.Lock()
		// change value after [0..500] ms delay.
		time.Sleep(time.Duration(rand.Int()%6) * time.Millisecond * 100)
		sharedValue = 1
		mut.Unlock()
	}()
	// give another goroutine some time to lock the mutex.
	time.Sleep(10 * time.Millisecond)
	if tmut.LockTimeout(250 * time.Millisecond) {
		if sharedValue != 1 {
			panic("bad value")
		}
		tmut.Unlock()
	} else {
	} // timeout elapsed
}

func ExampleCond() {
	DestroyMutex("mut")
	mut, err := NewMutex("mut", os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		panic("new")
	}
	defer mut.Close()
	DestroyCond("cond")
	cond, err := NewCond("cond", os.O_CREATE|os.O_EXCL, 0666, mut)
	if err != nil {
		panic("new")
	}
	defer cond.Close()
	var sharedValue int
	go func() {
		mut.Lock()
		defer mut.Unlock()
		sharedValue = 1
		cond.Signal()
	}()
	mut.Lock()
	defer mut.Unlock()
	if sharedValue == 0 {
		cond.Wait()
		if sharedValue == 0 {
			panic("bad value")
		}
	}
}

func ExampleEvent() {
	event, err := NewEvent("event", os.O_CREATE|os.O_EXCL, 0666, false)
	if err != nil {
		return
	}
	go func() {
		event.Set()
	}()
	if event.WaitTimeout(time.Millisecond * 250) {
		// event has been set
	} else {
		// timeout elapsed
	}
	event.Destroy()
}
