package commands

import (
	"github.com/ForeverSRC/MyDocker/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		if err := container.RunContainerInitProcess(); err != nil {
			log.Errorf("run init command error: %v", err)
			return err
		}

		return nil
	},
}
