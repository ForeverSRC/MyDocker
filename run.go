package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ForeverSRC/MyDocker/cgroups"
	"github.com/ForeverSRC/MyDocker/cgroups/subsystems"
	"github.com/ForeverSRC/MyDocker/container"
	log "github.com/sirupsen/logrus"
)

const cgroupPathFormat = "my-docker-cgroup/%s"

func Run(tty bool, cmdArray []string, containerName string, res *subsystems.ResourceConfig) error {
	cID, cName := container.GenerateContainerIDAndName(containerName)

	parent, writePipe := container.NewParentProcess(tty, cID)
	if parent == nil {
		return fmt.Errorf("new parent process error")
	}

	if err := parent.Start(); err != nil {
		return err
	}

	if err := container.RecordContainerInfo(cID, parent.Process.Pid, cmdArray, cName); err != nil {
		syscall.Kill(parent.Process.Pid,syscall.SIGTERM)
		writePipe.Close()
		return fmt.Errorf("record container info error: %v", err)
	}

	cgroupManager := cgroups.NewCgroupManager(fmt.Sprintf(cgroupPathFormat, cID))
	//defer cgroupManager.Destroy()

	if err := cgroupManager.Set(res); err != nil {
		syscall.Kill(parent.Process.Pid,syscall.SIGTERM)
		writePipe.Close()
		return err
	}

	if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
		syscall.Kill(parent.Process.Pid,syscall.SIGTERM)
		writePipe.Close()
		return err
	}

	if err := sendInitCommand(cmdArray, writePipe); err != nil {
		syscall.Kill(parent.Process.Pid,syscall.SIGTERM)
		return err
	}

	if tty {
		parent.Wait()
	}

	//container.DeleteWorkSpace(cID)

	return nil
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
