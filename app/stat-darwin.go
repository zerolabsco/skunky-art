//go:build darwin

package app

import (
	"syscall"
	"time"
)

// macOS names the ctime field Ctimespec rather than Ctim, so it needs its own
// variant to let the project build and test on a Mac (deploys are Linux).
func statTime(stat *syscall.Stat_t) int64 {
	return time.Unix(stat.Ctimespec.Unix()).UnixMilli()
}
