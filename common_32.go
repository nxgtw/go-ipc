// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build 386 arm

package ipc

import "time"

// splitUnixTime splits unix time given in nanoseconds into secs and nsecs parts
func splitUnixTime(utime int64) (int32, int32) {
	return int32(utime / int64(time.Second)), int32(utime % int64(time.Second))
}
