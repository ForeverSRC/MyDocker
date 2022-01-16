package container_cmd

import (
	"github.com/urfave/cli"
)

var ContainerCommands = []cli.Command{
	execCommand,
	initCommand,
	listCommand,
	logCommand,
	removeCommand,
	runCommand,
	stopCommand,
}
