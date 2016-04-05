// Copyright 2015 Aleksandr Demakin. All rights reserved.
// ignore this for a while, as linux rw mutexes don't work,
// and windows mutexes are not ready yes.

package sync

import (
	"os"
	"runtime"
	"testing"
	"time"

	"bitbucket.org/avd/go-ipc"

	"github.com/stretchr/testify/assert"
)

func mutexCtor(name string, mode int, perm os.FileMode) (IPCLocker, error) {
	return NewMutex(name, mode, perm)
}

func mutexDtor(name string) error {
	return DestroyMutex(name)
}

func TestMutexOpenMode(t *testing.T) {
	testLockerOpenMode(t, mutexCtor, mutexDtor)
}

func TestMutexOpenMode2(t *testing.T) {
	testLockerOpenMode2(t, mutexCtor, mutexDtor)
}

func TestMutexOpenMode3(t *testing.T) {
	testLockerOpenMode3(t, mutexCtor, mutexDtor)
}

func TestMutexOpenMode4(t *testing.T) {
	testLockerOpenMode4(t, mutexCtor, mutexDtor)
}

func TestMutexOpenMode5(t *testing.T) {
	testLockerOpenMode5(t, mutexCtor, mutexDtor)
}

func TestMutexLock(t *testing.T) {
	testLockerLock(t, mutexCtor, mutexDtor)
}

func TestMutexMemory(t *testing.T) {
	testLockerMemory(t, "m", mutexCtor, mutexDtor)
}

func TestMutexValueInc(t *testing.T) {
	testLockerValueInc(t, "m", mutexCtor, mutexDtor)
}

func TestMutexLockTimeout(t *testing.T) {
	a := assert.New(t)
	if !a.NoError(mutexDtor(testLockerName)) {
		return
	}
	m, err := mutexCtor(testLockerName, ipc.O_CREATE_ONLY, 0666)
	if !a.NoError(err) || !a.NotNil(m) {
		return
	}
	defer mutexDtor(testLockerName)
	tl, ok := m.(TimedIPCLocker)
	if !ok {
		t.Skipf("timed mutex is not supported on %s(%s)", runtime.GOOS, runtime.GOARCH)
		return
	}
	tl.Lock()
	defer tl.Unlock()
	before := time.Now()
	timeout := time.Millisecond * 50
	a.False(tl.LockTimeout(timeout))
	a.InEpsilon(int64(time.Now().Sub(before)), int64(timeout), 0.05)
}
