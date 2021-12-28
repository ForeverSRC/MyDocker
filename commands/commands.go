package commands

import "github.com/urfave/cli"

var AllCommands = []cli.Command{
	initCommand,
	runCommand,
	listCommand,
	logCommand,
	stopCommand,
	removeCommand,
	execCommand,
}
