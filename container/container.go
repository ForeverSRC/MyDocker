package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/ForeverSRC/MyDocker/utils"
	log "github.com/sirupsen/logrus"
)

func GenerateContainerIDAndName(containerName string) (string, string) {
	id := randStringBytes(10)
	if containerName == "" {
		containerName = id
	}

	return id, containerName
}

func RecordContainerInfo(containerInfo *ContainerInfo) error {
	containerInfo.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	containerInfo.Status = RUNNING

	if err := containerInfo.dump(); err != nil {
		log.Errorf("record container info error: %v", err)
		return err
	}

	return nil
}

func GetContainerInfoById(containerID string) (*ContainerInfo, error) {
	configFilePath := getContainerConfigFilePath(containerID)

	var containerInfo = &ContainerInfo{}
	if err := containerInfo.load(configFilePath); err != nil {
		return nil, err
	}

	return containerInfo, nil
}

func ListContainers() ([]*ContainerInfo, error) {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, "")
	dirUrl = dirUrl[:len(dirUrl)-1]

	dirs, err := ioutil.ReadDir(dirUrl)
	if err != nil {
		return nil, fmt.Errorf("read dir %s error:%v", dirUrl, err)

	}

	var containers []*ContainerInfo

	for _, dir := range dirs {
		containerID := dir.Name()
		tmpContainer, err := GetContainerInfoById(containerID)
		if err != nil {
			log.Errorf("get container info error: %v", err)
			continue
		}

		if tmpContainer.Status == RUNNING {
			pid, err := strconv.Atoi(tmpContainer.Pid)
			containerProcNotExist := !utils.ProcessExist(pid)
			if err != nil || containerProcNotExist {
				log.Warnf("process of container %s is not exist", containerID)
				tmpContainer = SetContainerStateToStop(containerID)
			}
		}

		containers = append(containers, tmpContainer)
	}

	return containers, nil
}

func LogContainer(containerID string) {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerID)
	logFileLocation := dirUrl + ContainerLogFile

	file, err := os.Open(logFileLocation)
	defer file.Close()

	if err != nil {
		log.Errorf("log container open file %s error: %v", logFileLocation, err)
		return
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("log container read file %s error: %v", logFileLocation, err)
		return
	}

	fmt.Fprint(os.Stdout, string(content))
}

func StopContainer(containerID string) {
	containerInfo, err := GetContainerInfoById(containerID)
	if err != nil {
		log.Errorf("get container info of %s error: %v", containerID, err)
		return
	}

	if containerInfo.Status == STOP {
		return
	}

	pid, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		log.Errorf("convert pid from string to int error: %v", err)
		return
	}

	exist := utils.ProcessExist(pid)
	if exist {
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			log.Errorf("stop container %s error: %v", containerID, err)
			return
		}

		changeContainerInfoToStop(containerInfo)
	}
}

func SetContainerStateToStop(containerID string) *ContainerInfo {
	containerInfo, err := GetContainerInfoById(containerID)
	if err != nil {
		log.Errorf("get container info of %s error: %v", containerID, err)
		return nil
	}

	return changeContainerInfoToStop(containerInfo)
}

func changeContainerInfoToStop(containerInfo *ContainerInfo) *ContainerInfo {
	containerInfo.Status = STOP
	containerInfo.Pid = " "
	err := containerInfo.dump()
	if err != nil {
		log.Errorf("change container %s to stop error: %v", containerInfo.Id, err)
	}

	return containerInfo
}
