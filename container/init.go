package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

/*
RunContainerInitProcess 此方法在容器内部执行，生成本容器执行的第一个进程
使用mount挂载proc文件系统，以便后面通过ps等命令查看当前进程资源的情况
*/
func RunContainerInitProcess() error {
	// block
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("run container get user command error, cmdArray is nil")
	}

	if err := setUpMount(); err != nil {
		return err
	}

	// 可在系统的PATH中寻找命令的绝对路径
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		return fmt.Errorf("exec look path error %v", err)
	}

	// init进程读取了父进程传递过来的参数，在子进程内执行，完成了将用户指定命令传递给子进程的操作
	if err = syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf("syscall exec error: %v", err.Error())
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

func setUpMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd error:%v", err)
	}

	if err = pivotRoot(pwd); err != nil {
		return fmt.Errorf("pivot root error: %v", err)
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		return fmt.Errorf("mount proc error: %v", err)
	}

	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		return fmt.Errorf("mount tmpfs error: %v", err)
	}

	return nil

}

func pivotRoot(root string) error {
	// systemd 加入linux之后, mount namespace 就变成 shared by default, 必须显式声明新的mount namespace独立。
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		return err
	}
	// 重新mount root
	// bind mount：将相同内容换挂载点
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}

	// 创建 rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	// pivot_root 到新的rootfs, 老的 old_root挂载在rootfs/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root error: %v", err)
	}

	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir error: %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")

	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir error: %v", err)
	}

	// 删除临时文件夹
	return os.Remove(pivotDir)
}
