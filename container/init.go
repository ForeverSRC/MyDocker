package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

/*
RunContainerInitProcess 此方法在容器内部执行，生成本容器执行的第一个进程
使用mount挂载proc文件系统，以便后面通过ps等命令查看当前进程资源的情况
*/
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("run container get user command error, cmdArray is nil")
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err:=syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		log.Errorf("syscall mount error %v", err)
		return err
	}
	// 可在系统的PATH中寻找命令的绝对路径
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("exec look path error %v", err)
		return err
	}
	log.Infof("find path %s", path)
	// init进程读取了父进程传递过来的参数，在子进程内执行，完成了将用户指定命令传递给子进程的操作
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}

func readUserCommand() []string {
	// 0-stdin
	// 1-stdout
	// 2-stderr
	// 3-pipe
	pipe := os.NewFile(uintptr(3), "pipe")

	// block read
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}

	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
