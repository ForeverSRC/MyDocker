package container_cmd

import (
	"fmt"
	"os"

	"github.com/ForeverSRC/MyDocker/cgroups"
	"github.com/ForeverSRC/MyDocker/container"
	"github.com/ForeverSRC/MyDocker/network"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove unused container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container id")
		}

		containerID := context.Args().Get(0)
		RemoveContainer(containerID)

		return nil
	},
}

func RemoveContainer(containerID string) {
	containerInfo, err := container.GetContainerInfoById(containerID)
	if err != nil {
		log.Errorf("get container %s info error %v", containerID, err)
		return
	}

	if containerInfo.Status == container.RUNNING {
		log.Errorf("could not remove running container")
		return
	}

	// remove container config files
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerID)
	if err = os.RemoveAll(dirUrl); err != nil {
		log.Errorf("remove file %s error: %v", dirUrl, err)
		return
	}

	// remove container write layer
	if err = container.DeleteWorkSpace(containerID); err != nil {
		log.Errorf("remove workspace of container %s error: %v", containerID, err)
		return
	}

	// remove container cgroups
	cgroupManager := cgroups.NewCgroupManager(fmt.Sprintf(cgroups.CgroupPathFormat, containerID))
	if err = cgroupManager.Destroy(); err != nil {
		log.Errorf("remove cgroup of container %s error: %v", containerID, err)
	}

	// Release ip allocated for container
	if containerInfo.Network != "" && containerInfo.IpAddress != "" {
		network.Init()
		network.ReleaseIp(containerInfo.Network, containerInfo.IpAddress)
	}
}
