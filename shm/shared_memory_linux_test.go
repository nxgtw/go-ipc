// Copyright 2015 Aleksandr Demakin. All rights reserved.

package shm

import (
	"strings"
	"testing"
)

func TestShmFsFromReader(t *testing.T) {
	const (
		testData = `
			#
			# /etc/fstab
			# name dir type opts freq passno
			UUID=cd459033-ae0a-4fb4-96fb-2323365a8e21 /                       ext4    defaults        1 1
			UUID=4542ef12-df3d-4336-9d12-740763854139 /boot                   ext4    defaults        1 2
			UUID=95bd9dce-550c-4622-9466-6cd1e1ffd278 /home                   ext4    defaults        1 2
			UUID=53d61062-7b6b-4f5b-80fd-7baf4017f96d swap                    swap    defaults        0 0
			tmpfs /dev/shm tmpfs rw,seclabel,nosuid,nodev 0 0
		`
		testData2 = "tmpfs /dev/shm nottmpfs rw,seclabel,nosuid,nodev 0 0"
	)
	path := shmFsFromReader(strings.NewReader(testData))
	if path != "/dev/shm/" {
		t.Errorf("shm mountpoints not parsed. expected '/dev/shm/', got '%s'", path)
	}
	path = shmFsFromReader(strings.NewReader(testData2))
	if path != "" {
		t.Errorf("shm mountpoint should not be parsed. got '%s'", path)
	}
}

func TestShmFsFromMountPoints(t *testing.T) {
	path := shmFsFromMounts()
	if len(path) == 0 {
		t.Errorf("couldn't find a correct shm path")
	}
}
