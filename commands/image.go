package commands

import (
	"github.com/ForeverSRC/MyDocker/image"
	"github.com/ForeverSRC/MyDocker/utils"
	"github.com/urfave/cli"
)

var imageCommand = cli.Command{
	Name:  "image",
	Usage: "image commands",
	Subcommands: []cli.Command{
		imageListCmd,
	},
}

const imageTableTitle = "REPOSITORY\tTAG\tIMAGE ID\n"

var imageListCmd = cli.Command{
	Name:  "ls",
	Usage: "list images",
	Action: func(context *cli.Context) error {
		infos := image.ListImages()
		utils.PrintInfoTable(imageTableTitle, infos)
		return nil
	},
}
