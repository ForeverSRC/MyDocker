package utils

import (
	"syscall"
)

func ProcessExist(containerPid int) bool {
	if err := syscall.Kill(containerPid, 0); err != nil {
		return false
	}

	return true
}
