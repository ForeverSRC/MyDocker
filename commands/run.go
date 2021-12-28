package commands

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ForeverSRC/MyDocker/cgroups"
	"github.com/ForeverSRC/MyDocker/cgroups/subsystems"
	"github.com/ForeverSRC/MyDocker/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name: "run",
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
	},

	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		// 用户指定运行的命令
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}

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
		if err := Run(tty, cmdArray, containerName, resConf); err != nil {
			log.Error(err)
			return err
		}

		return nil
	},
}

func Run(tty bool, cmdArray []string, containerName string, res *subsystems.ResourceConfig) error {
	cID, cName := container.GenerateContainerIDAndName(containerName)

	parent, writePipe := container.NewParentProcess(tty, cID)
	if parent == nil {
		return fmt.Errorf("new parent process error")
	}

	if err := parent.Start(); err != nil {
		return err
	}

	if err := container.RecordContainerInfo(cID, parent.Process.Pid, cmdArray, cName); err != nil {
		syscall.Kill(parent.Process.Pid, syscall.SIGTERM)
		writePipe.Close()
		return fmt.Errorf("record container info error: %v", err)
	}

	cgroupManager := cgroups.NewCgroupManager(fmt.Sprintf(cgroups.CgroupPathFormat, cID))

	if err := cgroupManager.Set(res); err != nil {
		syscall.Kill(parent.Process.Pid, syscall.SIGTERM)
		writePipe.Close()
		return err
	}

	if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
		syscall.Kill(parent.Process.Pid, syscall.SIGTERM)
		writePipe.Close()
		return err
	}

	if err := sendInitCommand(cmdArray, writePipe); err != nil {
		syscall.Kill(parent.Process.Pid, syscall.SIGTERM)
		return err
	}

	if tty {
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
