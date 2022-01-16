package container_cmd

import (
	"fmt"

	"github.com/ForeverSRC/MyDocker/container"
	"github.com/urfave/cli"
)

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("container name missed, please input")
		}
		containerID := context.Args().Get(0)
		container.StopContainer(containerID)
		return nil
	},
}
