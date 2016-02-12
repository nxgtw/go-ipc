// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	cShmfsSuperMagic = 0x01021994
	cRamfsMagic      = 0x858458f6
)

type mntent struct {
	fsname string /* Device or server for filesystem.  */
	dir    string /* Directory mounted on.  */
	fstype string /* Type of filesystem: ufs, nfs, etc.  */
	opts   string /* Comma-separated options for fs.  */
	freq   int    /* Dump frequency (in days).  */
	passno int    /* Pass number for `fsck'.  */
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
