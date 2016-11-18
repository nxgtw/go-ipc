// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux windows

package mq

import (
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
}

type prioMqCtor func(name string, flag int, perm os.FileMode, maxQueueSize, maxMsgSize int) (PriorityMessenger, error)
type prioMqOpener func(name string, flag int) (PriorityMessenger, error)

func testPrioMq1(t *testing.T, ctor prioMqCtor, opener prioMqOpener, dtor mqDtor) {
	prios := [...]int{8, 4, 7, 1, 0, 15, 2, 4}
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, O_NONBLOCK, 0666, len(prios), 8)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(mq.Close())
		if dtor != nil {
			a.NoError(dtor(testMqName))
		}
	}()
	message := make([]byte, 8)
	for _, prio := range prios {
		message[0] = byte(prio)
		if !a.NoError(mq.SendPriority(message, prio)) {
			return
		}
	}
	sort.Ints(prios[:])
	for i := len(prios) - 1; i >= 0; i-- {
		message := make([]byte, 8)
		l, prio, err := mq.ReceivePriority(message)
		if !a.NoError(err) {
			continue
		} else {
			a.Equal(len(message), l)
			a.Equal(prios[i], prio, "error at %d", i)
			a.Equal(byte(prio), message[0])
		}
	}
}

type prioBenchmarkParams struct {
	readers int
	writers int
	mqSize  int
	msgSize int
	flag    int
}

func benchmarkPrioMq1(b *testing.B, ctor prioMqCtor, opener prioMqOpener, dtor mqDtor, params *prioBenchmarkParams) {
	if dtor != nil {
		dtor(testMqName)
	}
	mq, err := ctor(testMqName, params.flag|O_NONBLOCK, 0666, params.mqSize, params.msgSize)
	if err != nil {
		b.Error(err)
		return
	}
	defer func() {
		mq.Close()
		if dtor != nil {
			dtor(testMqName)
		}
	}()
	var wgw, wgr sync.WaitGroup
	wgw.Add(params.writers)
	wgr.Add(params.readers)
	var sent, received, done int32
	for i := 0; i < params.writers; i++ {
		go func() {
			defer wgw.Done()
			inst, err := opener(testMqName, params.flag)
			if err != nil {
				b.Error(err)
				return
			}
			defer inst.Close()
			rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
			mess := make([]byte, params.msgSize)
			for j := 0; j < b.N; j++ {
				mess[0] = byte(rnd.Intn(100))
				if err := inst.SendPriority(mess, int(mess[0])); err != nil {
					if !(params.flag&O_NONBLOCK != 0 && IsTemporary(err)) {
						b.Error(err)
					}
				} else {
					atomic.AddInt32(&sent, 1)
				}
			}
		}()
	}
	for i := 0; i < params.readers; i++ {
		go func() {
			defer wgr.Done()
			inst, err := opener(testMqName, params.flag)
			if err != nil {
				b.Error(err)
				return
			}
			defer inst.Close()
			mess := make([]byte, params.msgSize)
			for j := 0; j < b.N; j++ {
				if l, prio, err := inst.ReceivePriority(mess); err != nil {
					if !(params.flag&O_NONBLOCK != 0 && IsTemporary(err)) {
						b.Error(err)
					}
				} else if prio == int(mess[0]) {
					if l != params.msgSize {
						b.Errorf("error in msg len: %d != %d", params.msgSize, l)
					} else {
						atomic.AddInt32(&received, 1)
					}
				} else {
					b.Errorf("error in prio: %d != %d", prio, int(mess[0]))
				}
			}
		}()
	}
	wgw.Wait()
	atomic.StoreInt32(&done, 1)
	if params.flag&O_NONBLOCK == 0 { // wake up all readers
		mess := make([]byte, params.msgSize)
		for i := 0; i < params.readers; i++ {
			if err := mq.SendPriority(mess, 0); err != nil && !IsTemporary(err) {
				b.Error(err)
			} else {
				atomic.AddInt32(&sent, 1)
			}
		}
	}
	wgr.Wait()
	if sent > 0 {
		b.Logf("%d of %d (%2.2f%%) messages received for N = %d", received, sent, float64(received)/float64(sent)*100, b.N)
	}
}
