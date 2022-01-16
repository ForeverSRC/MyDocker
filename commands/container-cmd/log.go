package container_cmd

import (
	"fmt"

	"github.com/ForeverSRC/MyDocker/container"
	"github.com/urfave/cli"
)

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("container id missed, please input")
		}
		containerID := context.Args().Get(0)
		container.LogContainer(containerID)
		return nil
	},
}
