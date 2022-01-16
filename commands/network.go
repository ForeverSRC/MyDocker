package commands

import (
	"fmt"

	"github.com/ForeverSRC/MyDocker/network"
	"github.com/ForeverSRC/MyDocker/utils"
	"github.com/urfave/cli"
)

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		networkCreateCmd,
		networkListCmd,
		networkRemoveCmd,
	},
}

var networkCreateCmd = cli.Command{
	Name:  "create",
	Usage: "create a container network",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "driver",
			Usage: "network driver",
		},
		cli.StringFlag{
			Name:  "subnet",
			Usage: "subnet cider",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing network name")
		}

		network.Init()
		err := network.CreateNetwork(context.String("driver"), context.String("subnet"), context.Args()[0])
		if err != nil {
			return fmt.Errorf("create network error %v", err)
		}

		return nil
	},
}

const networkTableTitle = "NAME\tSubnet\tGateway\tDriver\n"

var networkListCmd = cli.Command{
	Name:  "ls",
	Usage: "list container network",
	Action: func(context *cli.Context) error {
		network.Init()
		infos := network.ListNetwork()
		utils.PrintInfoTable(networkTableTitle, infos)
		return nil
	},
}

var networkRemoveCmd = cli.Command{
	Name:  "rm",
	Usage: "remove container network",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing network name")
		}

		network.Init()
		err := network.DeleteNetwork(context.Args()[0])
		if err != nil {
			return fmt.Errorf("remove network error: %v", err)
		}

		return nil
	},
}
