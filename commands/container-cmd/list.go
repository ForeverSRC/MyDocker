package container_cmd

import (
	"fmt"

	"github.com/ForeverSRC/MyDocker/container"
	"github.com/ForeverSRC/MyDocker/utils"
	"github.com/urfave/cli"
)

const containerTableTitle = "CONTAINER ID\tIMAGE\tName\tPID\tSTATUS\tCOMMAND\tCREATED\tNETWORK\tIP\tPort Mapping\n"

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all containers",
	Action: func(context *cli.Context) error {
		cInfos, err := container.ListContainers()
		if err != nil {
			return err
		}

		printContainerInfoTable(cInfos)

		return nil
	},
}

func printContainerInfoTable(containers []*container.ContainerInfo) {
	infos := make([]string, len(containers))
	for idx, item := range containers {
		infos[idx] = fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Image,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreateTime,
			item.Network,
			item.IpAddress,
			item.PortMapping,
		)
	}

	utils.PrintInfoTable(containerTableTitle, infos)
}
