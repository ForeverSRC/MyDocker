package main

import (
	"os"

	"github.com/ForeverSRC/MyDocker/commands"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const usage = `a simple container runtime implementation. Just for fun.`

func main() {
	app := cli.NewApp()
	app.Name = "my-docker"
	app.Usage = usage

	app.Commands = commands.AllCommands

	app.Before = func(context *cli.Context) error {
		log.SetReportCaller(true)
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
