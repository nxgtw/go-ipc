// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux windows

package mq

import (
	"os"
	"reflect"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/test"
	ipc_sync "bitbucket.org/avd/go-ipc/sync"

	"github.com/stretchr/testify/assert"
)

const (
	testMqName = "go-ipc.mq"
	mqProgPath = "./internal/test/"
)

type mqCtor func(name string, flag int, perm os.FileMode) (Messenger, error)
type mqOpener func(name string, flag int) (Messenger, error)
type mqDtor func(name string) error

var (
	mqProgArgs       []string
	defaultMutexType = "m"
)

func detectMutexType() {
	ipc_sync.DestroyMutex("testLocker")
	m, err := ipc_sync.NewMutex("testLocker", os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	t := reflect.ValueOf(m)
	if t.Elem().Type().Name() == "SemaMutex" {
		defaultMutexType = "msysv"
	}
	m.Close()
	ipc_sync.DestroyMutex("testLocker")
}

func locate(path string) []string {
	files, err := testutil.LocatePackageFiles(path)
	if err != nil {
		panic(err)
	}
	if len(files) == 0 {
		panic("no locker test files")
	}
	for i, name := range files {
		files[i] = path + name
	}
	if defaultMutexType == "msysv" {
		files = append([]string{`-tags="sysv_mutex_linux"`}, files...)
	}
	return files
}

func init() {
	detectMutexType()
	mqProgArgs = locate(mqProgPath)
}

func testCreateMq(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if a.NoError(err) {
		a.NoError(mq.Close())
		if dtor != nil {
			a.NoError(dtor(testMqName))
		}
	}
}

func testCreateMqExcl(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !a.NoError(err) || !a.NotNil(mq) {
		return
	}
	_, err = ctor(testMqName, os.O_EXCL, 0666)
	a.Error(err)
	if d, ok := mq.(ipc.Destroyer); ok {
		a.NoError(d.Destroy())
	} else {
		a.NoError(mq.Close())
	}
}

func testCreateMqInvalidPerm(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	_, err := ctor(testMqName, os.O_EXCL, 0777)
	a.Error(err)
}

func testOpenMq(t *testing.T, ctor mqCtor, opener mqOpener, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !a.NoError(err) || !a.NotNil(mq) {
		return
	}
	a.NoError(mq.Close())
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	_, err = opener(testMqName, 0)
	a.Error(err)
}

func testMqSendIntSameProcess(t *testing.T, ctor mqCtor, opener mqOpener, dtor mqDtor) {
	var message = uint64(0xDEADBEEFDEADBEEF)
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(mq.Close())
		if dtor != nil {
			a.NoError(dtor(testMqName))
		}
	}()
	data, _ := allocator.ObjectData(&message)
	if !a.NoError(mq.Send(data)) {
		return
	}
	var received uint64
	mqr, err := opener(testMqName, 0)
	if !a.NoError(err) {
		return
	}
	data, _ = allocator.ObjectData(&received)
	l, err := mqr.Receive(data)
	a.NoError(err)
	a.Equal(int(unsafe.Sizeof(message)), l)
	a.Equal(message, received)
	allocator.UseValue(data)
	a.NoError(mqr.Close())
}

func testMqSendStructSameProcess(t *testing.T, ctor mqCtor, opener mqOpener, dtor mqDtor) {
	type testStruct struct {
		arr [16]int
		c   complex128
		s   struct {
			a, b byte
		}
		f float64
	}
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	message := testStruct{
		arr: [...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		c:   complex(2, -3),
		f:   11.22,
		s:   struct{ a, b byte }{127, 255},
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !a.NoError(err) {
		return
	}
	go func() {
		data, _ := allocator.ObjectData(message)
		a.NoError(mq.Send(data))
	}()
	received := testStruct{}
	mqr, err := opener(testMqName, 0)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(mqr.Close())
		a.NoError(dtor(testMqName))
	}()
	data, _ := allocator.ObjectData(&received)
	l, err := mqr.Receive(data)
	a.NoError(err)
	a.Equal(int(unsafe.Sizeof(message)), l)
	a.Equal(message, received)
	a.NoError(mq.Close())
	allocator.UseValue(data)
}

func testMqSendMessageLessThenBuffer(t *testing.T, ctor mqCtor, opener mqOpener, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !a.NoError(err) {
		return
	}
	message := make([]byte, 512)
	for i := range message {
		message[i] = byte(i)
	}
	go func() {
		a.NoError(mq.Send(message))
	}()
	received := make([]byte, 1024)
	mqr, err := opener(testMqName, 0)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(mqr.Close())
		a.NoError(dtor(testMqName))
	}()
	l, err := mqr.Receive(received)
	a.NoError(err)
	a.Equal(len(message), l)
	a.Equal(message, received[:512])
	a.Equal(received[512:], make([]byte, 512))
	a.NoError(mq.Close())
}

func testMqSendNonBlock(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		a.NoError(dtor(testMqName))
	}()
	if b, ok := mq.(Blocker); ok {
		a.NoError(b.SetBlocking(false))
		endChan := make(chan bool, 1)
		go func() {
			data := make([]byte, 8)
			for i := 0; i < 100; i++ {
				err := mq.Send(data)
				a.True(err == nil || IsTemporary(err))
			}
			endChan <- true
		}()
		select {
		case <-endChan:
		case <-time.After(time.Millisecond * 300):
			t.Errorf("send on non-blocking mq blocked")
		}
	} else {
		t.Skipf("current mq impl on %s does not implement Blocker", runtime.GOOS)
	}
}

func testMqSendTimeout(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		a.NoError(dtor(testMqName))
	}()
	if tmq, ok := mq.(TimedMessenger); ok {
		data := make([]byte, 8)
		tm := time.Millisecond * 200
		if buf, ok := mq.(Buffered); ok {
			cap, err := buf.Cap()
			if !a.NoError(err) {
				return
			}
			for i := 0; i < cap; i++ {
				if !a.NoError(mq.Send(data)) {
					return
				}
			}
		}
		now := time.Now()
		err := tmq.SendTimeout(data, tm)
		a.Error(err)
		a.True(IsTemporary(err))
		a.Condition(func() bool {
			return time.Since(now) >= tm
		})
	} else {
		t.Skipf("current mq impl on %s does not implement TimedMessenger", runtime.GOOS)
	}
}

func testMqReceiveTimeout(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		a.NoError(dtor(testMqName))
	}()
	if tmq, ok := mq.(TimedMessenger); ok {
		received := make([]byte, 8)
		tm := time.Millisecond * 200
		now := time.Now()
		_, err := tmq.ReceiveTimeout(received, tm)
		a.Error(err)
		if sysErr, ok := err.(syscall.Errno); ok {
			a.True(sysErr.Temporary())
		}
		a.Condition(func() bool {
			return time.Since(now) >= tm
		})
	} else {
		t.Skipf("current mq impl on %s does not implement TimedMessenger", runtime.GOOS)
	}
}

func testMqReceiveNonBlock(t *testing.T, ctor mqCtor, dtor mqDtor) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		a.NoError(dtor(testMqName))
	}()
	if b, ok := mq.(Blocker); ok {
		a.NoError(b.SetBlocking(false))
		_ = b
		endChan := make(chan bool, 1)
		go func() {
			data := make([]byte, 8)
			for i := 0; i < 32; i++ {
				_, err := mq.Receive(data)
				a.Error(err)
			}
			endChan <- true
		}()
		select {
		case <-endChan:
		case <-time.After(time.Millisecond * 300):
			t.Errorf("receive on non-blocking mq blocked")
		}
	} else {
		t.Skipf("current mq impl on %s does not implement Blocker", runtime.GOOS)
	}
}

func testMqSendToAnotherProcess(t *testing.T, ctor mqCtor, dtor mqDtor, typ string) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		a.NoError(mq.Close())
		if dtor != nil {
			a.NoError(dtor(testMqName))
		}
	}()
	data := make([]byte, 2048)
	for i := range data {
		data[i] = byte(i)
	}
	args := argsForMqTestCommand(testMqName, -1, typ, "", data)
	go func() {
		a.NoError(mq.Send(data))
	}()
	result := testutil.RunTestApp(args, nil)
	if !a.NoError(result.Err) {
		t.Logf("program output is: '%s'", result.Output)
	}
}

func testMqReceiveFromAnotherProcess(t *testing.T, ctor mqCtor, dtor mqDtor, typ string) {
	a := assert.New(t)
	if dtor != nil {
		a.NoError(dtor(testMqName))
	}
	mq, err := ctor(testMqName, 0, 0666)
	if !a.NoError(err) {
		return
	}
	defer func() {
		a.NoError(dtor(testMqName))
	}()
	data := make([]byte, 2048)
	for i := range data {
		data[i] = byte(i)
	}
	args := argsForMqSendCommand(testMqName, -1, typ, "", data)
	result := testutil.RunTestApp(args, nil)
	if !a.NoError(result.Err) {
		t.Logf("program output is %s", result.Output)
	}
	received := make([]byte, 2048)
	l, err := mq.Receive(received)
	a.NoError(err)
	a.Equal(len(data), l)
	a.Equal(data, received)
}

// Mq test program

func argsForMqCreateCommand(name string, mqMaxSize, msgMazSize int) []string {
	return append(mqProgArgs, "-object="+name, "create", strconv.Itoa(mqMaxSize), strconv.Itoa(msgMazSize))
}

func argsForMqDestroyCommand(name string) []string {
	return append(mqProgArgs, "-object="+name, "destroy")
}

func argsForMqSendCommand(name string, timeout int, typ, options string, data []byte) []string {
	return append(mqProgArgs,
		"-object="+name,
		"-type="+typ,
		"-options="+options,
		"-timeout="+strconv.Itoa(timeout),
		"send",
		testutil.BytesToString(data),
	)
}

func argsForMqTestCommand(name string, timeout int, typ, options string, data []byte) []string {
	return append(mqProgArgs,
		"-object="+name,
		"-type="+typ,
		"-options="+options,
		"-timeout="+strconv.Itoa(timeout),
		"test",
		testutil.BytesToString(data),
	)
}

func argsForMqNotifyWaitCommand(name string, timeout int, typ, options string) []string {
	return append(mqProgArgs,
		"-object="+name,
		"-type="+typ,
		"-options="+options,
		"-timeout="+strconv.Itoa(timeout),
		"notifywait",
	)
}
