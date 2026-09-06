//go:build windows

package main

import (
	"syscall"
	"time"
)

func currentProcessCPU() (time.Duration, bool) {
	handle, err := syscall.GetCurrentProcess()
	if err != nil {
		return 0, false
	}

	var creationTime syscall.Filetime
	var exitTime syscall.Filetime
	var kernelTime syscall.Filetime
	var userTime syscall.Filetime

	if err := syscall.GetProcessTimes(handle, &creationTime, &exitTime, &kernelTime, &userTime); err != nil {
		return 0, false
	}

	return time.Duration(kernelTime.Nanoseconds() + userTime.Nanoseconds()), true
}
