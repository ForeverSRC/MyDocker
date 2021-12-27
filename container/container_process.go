package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	containerMntURL        = "/root/my-docker/aufs/mnt/container-%s/"
	containerWriteLayerUrl = "/root/my-docker/aufs/diff/rw-%s/"
	containerRootFs        = "/root/my-docker/aufs/diff/busybox"
)

func NewParentProcess(tty bool, containerID string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("new pipe error %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		dirUrl := fmt.Sprintf(DefaultInfoLocation, containerID)
		if err := os.MkdirAll(dirUrl, 0622); err != nil {
			log.Errorf("new parent process mkdir %s error: %v", dirUrl, err)
			return nil, nil
		}

		stdLogFilePath := dirUrl + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			log.Errorf("new parent process create file %s error: %v", stdLogFilePath, err)
			return nil, nil
		}

		cmd.Stdout = stdLogFile

	}

	// 传入管道文件读取端句柄
	// 外带此句柄去创建子进程
	cmd.ExtraFiles = []*os.File{readPipe}
	// 指定容器初始化后的工作目录
	mntUrl, err := NewWorkSpace(containerID)
	if err != nil {
		log.Errorf("new workspace error: %v", err)
		return nil, nil
	}

	cmd.Dir = mntUrl

	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	return read, write, nil
}

// NewWorkSpace Create a AUFS filesystem as container root workspace
func NewWorkSpace(containerID string) (string, error) {
	writeUrl, err := CreateWriteLayer(containerID)
	if err != nil {
		return "", err
	}

	return CreateMountPoint(containerID, writeUrl)
}

func CreateWriteLayer(containerID string) (string, error) {
	writeURL := fmt.Sprintf(containerWriteLayerUrl, containerID)
	if err := os.Mkdir(writeURL, 0777); err != nil {
		return "", fmt.Errorf("mkdir %s error: %v", writeURL, err)
	}

	return writeURL, nil
}

func CreateMountPoint(containerID, writeUrl string) (string, error) {
	mntUrl := fmt.Sprintf(containerMntURL, containerID)
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		return "", fmt.Errorf("mkdir %s error: %v", mntUrl, err)
	}

	dirs := fmt.Sprintf("dirs=%s:%s", writeUrl, containerRootFs)
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	return mntUrl, err
}

func DeleteWorkSpace(containerID string) error {
	if err := DeleteMountPoint(containerID); err != nil {
		return err
	}

	return DeleteWriterLayer(containerID)
}

func DeleteMountPoint(containerID string) error {
	mntUrl := fmt.Sprintf(containerMntURL, containerID)

	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	if err := os.RemoveAll(mntUrl); err != nil {
		return fmt.Errorf("remove dir %s error: %v", mntUrl, err)
	}

	return nil
}

func DeleteWriterLayer(containerID string) error {
	writeURL := fmt.Sprintf(containerWriteLayerUrl, containerID)
	return os.RemoveAll(writeURL)
}

func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
