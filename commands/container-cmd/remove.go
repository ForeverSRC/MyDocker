package container_cmd

import (
	"fmt"

	"github.com/ForeverSRC/MyDocker/container"
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
		container.RemoveContainer(containerID)
		return nil
	},
}
