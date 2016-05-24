// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build linux

package shm

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

const (
	maxNameLen       = 255
	defaultShmPath   = "/dev/shm/"
	cShmfsSuperMagic = 0x01021994
	cRamfsMagic      = 0x858458f6
)

var (
	shmPathOnce sync.Once
	shmPath     string
)

type mntent struct {
	fsname string /* Device or server for filesystem.  */
	dir    string /* Directory mounted on.  */
	fstype string /* Type of filesystem: ufs, nfs, etc.  */
	opts   string /* Comma-separated options for fs.  */
	freq   int    /* Dump frequency (in days).  */
	passno int    /* Pass number for `fsck'.  */
}

func doDestroyMemoryObject(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// glibc/sysdeps/posix/shm_open.c
func shmOpen(path string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, flag, perm)
}

// glibc/sysdeps/posix/shm-directory.h
func shmName(name string) (string, error) {
	name = strings.TrimLeft(name, "/")
	nameLen := len(name)
	if nameLen == 0 || nameLen >= maxNameLen || strings.Contains(name, "/") {
		return "", errors.New("invalid shm name")
	}
	var dir string
	var err error
	if dir, err = shmDirectory(); err != nil {
		return "", errors.Wrap(err, "error building shared memory name")
	}
	return dir + name, nil
}

func shmDirectory() (string, error) {
	shmPathOnce.Do(locateShmFs)
	if len(shmPath) == 0 {
		return shmPath, errors.New("error locating the shared memory path")
	}
	return shmPath, nil
}

// glibc/sysdeps/unix/sysv/linux/shm-directory.c
func locateShmFs() {
	if checkShmPath(defaultShmPath) {
		shmPath = defaultShmPath
	} else {
		shmPath = shmFsFromMounts()
	}
}

func checkShmPath(path string) bool {
	if len(path) == 0 {
		return false
	}
	var statfs unix.Statfs_t
	if err := unix.Statfs(path, &statfs); err != nil {
		return false
	}
	// unconvert says 'warning: redundant type conversion',
	// however, it is not, as statfs.Type has different types on different platforms.
	return isShmFs(int64(statfs.Type))
}

func isShmFs(fsType int64) bool {
	return fsType == cShmfsSuperMagic || fsType == cRamfsMagic
}

func shmFsFromMounts() string {
	var fsFile *os.File
	var err error
	if fsFile, err = os.Open("/proc/mounts"); err != nil {
		if fsFile, err = os.Open("/etc/fstab"); err != nil {
			return ""
		}
	}
	return shmFsFromReader(fsFile)
}

func shmFsFromReader(r io.Reader) string {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if record := scanMountRecord(line); record != nil {
			if record.fstype == "tmpfs" || record.fstype == "shm" {
				result := record.dir
				if checkShmPath(result) {
					if !strings.HasSuffix(result, "/") {
						result = result + "/"
					}
					return result
				}
			}
		}
	}
	return ""
}

func scanMountRecord(record string) (result *mntent) {
	wordScanner := bufio.NewScanner(strings.NewReader(record))
	wordScanner.Split(bufio.ScanWords)
	if wordScanner.Scan() {
		word := wordScanner.Text()
		if strings.HasPrefix(word, "#") {
			return
		}
		result = &mntent{fsname: word}
	} else {
		return
	}
	if wordScanner.Scan() {
		result.dir = wordScanner.Text()
	} else {
		return
	}
	if wordScanner.Scan() {
		result.fstype = wordScanner.Text()
	} else {
		return
	}
	if wordScanner.Scan() {
		result.opts = wordScanner.Text()
	} else {
		return
	}
	var err error
	if wordScanner.Scan() {
		if result.freq, err = strconv.Atoi(wordScanner.Text()); err != nil {
			return
		}
	} else {
		return
	}
	if wordScanner.Scan() {
		if result.passno, err = strconv.Atoi(wordScanner.Text()); err != nil {
			return
		}
	} else {
		return
	}
	return
}
