package container_cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/ForeverSRC/MyDocker/cgroups"
	"github.com/ForeverSRC/MyDocker/cgroups/subsystems"
	"github.com/ForeverSRC/MyDocker/container"
	img "github.com/ForeverSRC/MyDocker/image"
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
		cli.StringSliceFlag{
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

		if err := run(runCmdArgs); err != nil {
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

func run(args *runArgs) (err error) {
	var mntUrl string
	var parent *exec.Cmd = nil
	var writePipe *os.File = nil
	var cgroupManager *cgroups.CgroupManager = nil

	cinfo := &container.ContainerInfo{}

	defer func() {
		rcv := recover()
		if err != nil || rcv != nil {
			if parent != nil {
				_ = syscall.Kill(parent.Process.Pid, syscall.SIGTERM)
			}

			if writePipe != nil {
				_ = writePipe.Close()
			}

			if mntUrl != "" {
				_ = container.DeleteWorkSpace(cinfo.Id)
			}

			if cgroupManager != nil {
				_ = cgroupManager.Destroy()
			}
		}
	}()

	imageID := checkImage(args.image)
	if imageID == "" {
		log.Errorf("image:%s is not exist", args.image)
		return
	}

	cinfo.Image = args.image

	cID, cName := container.GenerateContainerIDAndName(args.containerName)
	cinfo.Id = cID
	cinfo.Name = cName

	mntUrl, err = createContainerWorkspace(imageID, cID)
	if err != nil {
		return
	}

	parent, writePipe = container.NewParentProcess(mntUrl, args.tty, cID, args.envSlice)
	if parent == nil {
		err = fmt.Errorf("new parent process error")
		return
	}

	// 只有Start()后才有Process，才能有pid
	if err = parent.Start(); err != nil {
		return
	}

	cinfo.Pid = strconv.Itoa(parent.Process.Pid)

	cgroupManager = cgroups.NewCgroupManager(fmt.Sprintf(cgroups.CgroupPathFormat, cID))
	if err = createCgroups(cgroupManager, args.res, parent.Process.Pid); err != nil {
		return
	}

	if args.network != "" {
		cinfo.PortMapping = args.portMapping

		var ip net.IP
		ip, err = processNetwork(args.network, cinfo)
		if err != nil {
			err = fmt.Errorf("error connect network %s: %v", args.network, err)
			return
		}

		cinfo.Network = args.network
		cinfo.IpAddress = ip.To4().String()
	}

	cinfo.Command = strings.Join(args.cmdArray, "")
	if err = container.RecordContainerInfo(cinfo); err != nil {
		err = fmt.Errorf("record container info error: %v", err)
		return
	}

	if err = sendInitCommand(args.cmdArray, writePipe); err != nil {
		return
	}

	if args.tty {
		err2 := parent.Wait()
		if err2 != nil {
			log.Errorf("parent wait return error: %v", err2)
		}

		container.SetContainerStateToStop(cID)
	}

	return
}

func checkImage(image string) string {
	imageID, err := img.GetImageID(image)
	if err != nil {
		log.Errorf("get image id of image: %s error: %v", image, err)
		return ""
	}

	return imageID
}

func createContainerWorkspace(imageID string, containerID string) (string, error) {
	mntUrl, err := container.NewWorkspace(imageID, containerID)
	if err != nil {
		log.Errorf("new workspace error: %v", err)
		return "", nil
	}

	return mntUrl, nil
}

func createCgroups(cgroupManager *cgroups.CgroupManager, res *subsystems.ResourceConfig, pid int) error {
	if err := cgroupManager.Set(res); err != nil {
		return err
	}

	if err := cgroupManager.Apply(pid); err != nil {
		return err
	}

	return nil
}

func processNetwork(nw string, cinfo *container.ContainerInfo) (net.IP, error) {
	network.Init()
	ip, err := network.Connect(nw, cinfo)
	if err != nil {
		return nil, fmt.Errorf("error connect network %s: %v", nw, err)
	}

	return ip, nil
}

func sendInitCommand(cmdArray []string, writePipe *os.File) error {
	defer writePipe.Close()
	command := strings.Join(cmdArray, " ")
	if _, err := writePipe.WriteString(command); err != nil {
		return fmt.Errorf("send init command [%s] error:%v", command, err)
	}

	return nil

}
