package commands

import (
	"github.com/ForeverSRC/MyDocker/container"
	"github.com/urfave/cli"
)

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all containers",
	Action: func(context *cli.Context) error {
		container.ListContainers()
		return nil
	},
}
