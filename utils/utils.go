package utils

import (
	"fmt"
	"os"
	"syscall"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
)

func ProcessExist(containerPid int) bool {
	if err := syscall.Kill(containerPid, 0); err != nil {
		return false
	}

	return true
}

func PrintInfoTable(title string, infos []string) {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, title)
	for _, item := range infos {
		fmt.Fprint(w, item)
	}

	if err := w.Flush(); err != nil {
		log.Errorf("flush error: %v", err)
	}
}
