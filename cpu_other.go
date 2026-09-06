//go:build !windows && !unix

package main

import "time"

func currentProcessCPU() (time.Duration, bool) {
	return 0, false
}
