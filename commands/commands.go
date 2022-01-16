package commands

import (
	containerCmd "github.com/ForeverSRC/MyDocker/commands/container-cmd"
	"github.com/urfave/cli"
)

var AllCommands []cli.Command

func init() {
	AllCommands = append(AllCommands, containerCmd.ContainerCommands...)
	AllCommands = append(AllCommands, networkCommand, commitCommand)
}
