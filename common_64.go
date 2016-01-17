// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build amd64 arm64 ppc64

package ipc

import "time"

// splitUnixTime splits unix time given in nanoseconds into secs and nsecs parts
func splitUnixTime(utime int64) (int64, int64) {
	return utime / int64(time.Second), utime % int64(time.Second)
}
