// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd netbsd openbsd solaris

package ipc

// glibc/sysdeps/posix/shm-directory.c
func locateShmFs() {
	shmPath = defaultShmPath
}
