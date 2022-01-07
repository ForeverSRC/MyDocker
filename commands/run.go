package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/ForeverSRC/MyDocker/cgroups"
	"github.com/ForeverSRC/MyDocker/cgroups/subsystems"
	"github.com/ForeverSRC/MyDocker/container"
	"github.com/ForeverSRC/MyDocker/network"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: `create a container: my-docker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment",
		},
		cli.StringFlag{
			Name:  "net",
			Usage: "container network",
		},
		cli.StringFlag{
			Name:  "p",
			Usage: "port mapping",
		},
	},

	Action: func(context *cli.Context) error {
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing image or container command")
		}

		var args []string
		for _, arg := range context.Args() {
			args = append(args, arg)
		}

		image := args[0]
		cmdArray := args[1:]

		tty := context.Bool("ti")
		detach := context.Bool("d")
		if tty && detach {
			return fmt.Errorf("-ti and -d parameter can not both provided")
		}

		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("mem"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}

		containerName := context.String("name")

		envSlice := context.StringSlice("e")
		network := context.String("net")
		portMapping := context.StringSlice("p")

		runCmdArgs := &runArgs{
			image:         image,
			tty:           tty,
			cmdArray:      cmdArray,
			containerName: containerName,
			res:           resConf,
			envSlice:      envSlice,
			network:       network,
			portMapping:   portMapping,
		}

		if err := Run(runCmdArgs); err != nil {
			log.Error(err)
			return err
		}

		return nil
	},
}

type runArgs struct {
	image         string
	tty           bool
	cmdArray      []string
	containerName string
	res           *subsystems.ResourceConfig
	envSlice      []string
	network       string
	portMapping   []string
}

func Run(args *runArgs) error {
	cID, cName := container.GenerateContainerIDAndName(args.containerName)

	parent, writePipe := container.NewParentProcess(args.image, args.tty, cID, args.envSlice)
	if parent == nil {
		return fmt.Errorf("new parent process error")
	}

	if err := parent.Start(); err != nil {
		return err
	}

	cancelParent := func() {
		syscall.Kill(parent.Process.Pid, syscall.SIGTERM)
		writePipe.Close()
	}

	if err := container.RecordContainerInfo(args.image, cID, parent.Process.Pid, args.cmdArray, cName); err != nil {
		cancelParent()
		return fmt.Errorf("record container info error: %v", err)
	}

	cgroupManager := cgroups.NewCgroupManager(fmt.Sprintf(cgroups.CgroupPathFormat, cID))

	if err := cgroupManager.Set(args.res); err != nil {
		cancelParent()
		return err
	}

	if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
		cancelParent()
		return err
	}

	if args.network != "" {
		network.Init()
		cinfo := &container.ContainerInfo{
			Id:          cID,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        cName,
			PortMapping: args.portMapping,
		}

		if err := network.Connect(args.network, cinfo); err != nil {
			log.Errorf("error connect network %s: %v", args.network, err)
			cancelParent()
			return err
		}
	}

	if err := sendInitCommand(args.cmdArray, writePipe); err != nil {
		syscall.Kill(parent.Process.Pid, syscall.SIGTERM)
		return err
	}

	if args.tty {
		err := parent.Wait()
		if err != nil {
			log.Errorf("parent wait return error: %v", err)
		}

		container.StopContainerForTty(cID)
	}

	return nil
}

func sendInitCommand(cmdArray []string, writePipe *os.File) error {
	defer writePipe.Close()
	command := strings.Join(cmdArray, " ")
	log.Infof("command all is [ %s ]", command)
	if _, err := writePipe.WriteString(command); err != nil {
		return fmt.Errorf("send init command [%s] error:%v", command, err)
	}

	return nil

}
