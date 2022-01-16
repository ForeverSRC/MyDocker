package container_cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ForeverSRC/MyDocker/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

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
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "CONTAINER ID\tIMAGE\tName\tPID\tSTATUS\tCOMMAND\tCREATED\tNETWORK\tIP\tPort Mapping\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
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

	if err := w.Flush(); err != nil {
		log.Errorf("flush error: %v", err)
	}

}
