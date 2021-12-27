package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ForeverSRC/MyDocker/cgroups"
	"github.com/ForeverSRC/MyDocker/cgroups/subsystems"
	"github.com/ForeverSRC/MyDocker/container"
	log "github.com/sirupsen/logrus"
)

func Run(tty bool, cmdArray []string, containerName string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty, containerName)
	if parent == nil {
		log.Errorf("new parent process error")
		return
	}

	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	containerName, err := container.RecordContainerInfo(parent.Process.Pid, cmdArray, containerName)
	if err != nil {
		log.Errorf("record container info error: %v", err)
		return
	}

	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()

	if err := cgroupManager.Set(res); err != nil {
		log.Error(err)
	}

	if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
		log.Error(err)
	}

	if err := sendInitCommand(cmdArray, writePipe); err != nil {
		log.Error(err)
	}

	if tty {
		parent.Wait()
	}

	mntURL := "/root/mnt"
	rootURL := "/root/"
	container.DeleteWorkSpace(rootURL, mntURL)
}

func sendInitCommand(cmdArray []string, writePipe *os.File) error {
	defer writePipe.Close()
	command := strings.Join(cmdArray, " ")
	log.Infof("command all is [ %s ]", command)
	if _, err := writePipe.WriteString(command); err != nil {
		return fmt.Errorf("send init command [%s] error:%v", command, err)
	}

	return nil

}
