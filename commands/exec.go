package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/ForeverSRC/MyDocker/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	_ "github.com/ForeverSRC/MyDocker/nsenter"
)

const ENV_EXEC_PID = "my_docker_pid"
const ENV_EXEC_CMD = "my_docker_cmd"

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		if os.Getenv(ENV_EXEC_PID) != "" {
			log.Infof("pid callback pid %d", os.Getgid())
			return nil
		}

		if len(context.Args()) < 2 {
			return fmt.Errorf("missing container id or command")
		}
		containerID := context.Args().Get(0)

		var cmdArray []string
		for _, arg := range context.Args().Tail() {
			cmdArray = append(cmdArray, arg)
		}

		execContainer(containerID, cmdArray)
		return nil
	},
}

func execContainer(containerID string, cmdArray []string) {
	containerInfo, err := container.GetContainerInfoById(containerID)
	if err != nil {
		log.Errorf("get container info error:%v", err)
		return
	}

	if containerInfo.Status != container.RUNNING {
		log.Errorf("container %s is not running", containerID)
		return
	}

	cmdStr := strings.Join(cmdArray, " ")
	log.Infof("exec container pid %s, command: %s", containerInfo.Pid, cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := os.Setenv(ENV_EXEC_PID, containerInfo.Pid); err != nil {
		log.Errorf("set ENV_EXEC_PID = %s error: %v", containerInfo.Pid, err)
		return
	}
	if err := os.Setenv(ENV_EXEC_CMD, cmdStr); err != nil {
		log.Errorf("set ENV_EXEC_CMD = %s error: %v", cmdStr, err)
		return
	}

	containerEnvs := getEnvByPid(containerInfo.Pid)
	cmd.Env = append(os.Environ(), containerEnvs...)

	if err := cmd.Run(); err != nil {
		log.Errorf("exec container %s error %v", containerID, err)
	}

	return

}

func getEnvByPid(pid string) []string {
	// 进程环境变量存放位置：/proc/PID/environ
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(path)

	if err != nil {
		log.Errorf("read file %s error: %v", path, err)
		return nil
	}

	envs := strings.Split(string(contentBytes), "\u0000")
	return envs
}
