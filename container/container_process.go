package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	img "github.com/ForeverSRC/MyDocker/image"

	log "github.com/sirupsen/logrus"
)

const (
	ContainerMntURL        = "/root/my-docker/aufs/mnt/container-%s/"
	ContainerWriteLayerUrl = "/root/my-docker/aufs/diff/rw-%s/"
	ContainerAUFSRootUrl   = "/root/my-docker/aufs/diff/%s/"
)

func NewParentProcess(image string, tty bool, containerID string, envSlice []string) (*exec.Cmd, *os.File) {
	// 指定容器初始化后的工作目录
	mntUrl, err := NewWorkSpace(image, containerID)
	if err != nil {
		log.Errorf("new workspace error: %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("new pipe error %v", err)
		return nil, nil
	}

	// 传入管道文件读取端句柄，外带此句柄去创建子进程
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), envSlice...)
	cmd.Dir = mntUrl

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
func NewWorkSpace(image, containerID string) (string, error) {
	imageID, err := img.GetImageID(image)
	if err != nil {
		return "", fmt.Errorf("get image ID error: %v", err)
	}

	writeUrl, err := CreateWriteLayer(containerID)
	if err != nil {
		return "", err
	}

	return CreateMountPoint(imageID, containerID, writeUrl)
}

func CreateWriteLayer(containerID string) (string, error) {
	writeURL := fmt.Sprintf(ContainerWriteLayerUrl, containerID)
	if err := os.Mkdir(writeURL, 0777); err != nil {
		return "", fmt.Errorf("mkdir %s error: %v", writeURL, err)
	}

	return writeURL, nil
}

func CreateMountPoint(imageID, containerID, writeUrl string) (string, error) {
	mntUrl := fmt.Sprintf(ContainerMntURL, containerID)
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		return "", fmt.Errorf("mkdir %s error: %v", mntUrl, err)
	}

	imageLayers, err := img.GetImageLayers(imageID)
	if err != nil {
		return "", err
	}

	roLayers := make([]string, len(imageLayers))
	l := len(imageLayers)
	for i := 0; i < l; i++ {
		roLayers[i] = imageLayers[l-1-i]
	}

	roLayerStr := strings.Join(roLayers, ":")

	dirs := fmt.Sprintf("dirs=%s:%s", writeUrl, roLayerStr)

	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	return mntUrl, err
}

func DeleteWorkSpace(containerID string) error {
	if err := DeleteMountPoint(containerID); err != nil {
		return err
	}

	return DeleteWriterLayer(containerID)
}

func DeleteMountPoint(containerID string) error {
	mntUrl := fmt.Sprintf(ContainerMntURL, containerID)

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
	writeURL := fmt.Sprintf(ContainerWriteLayerUrl, containerID)
	return os.RemoveAll(writeURL)
}
