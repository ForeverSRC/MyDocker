package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/ForeverSRC/MyDocker/cgroups"
	log "github.com/sirupsen/logrus"
)

func GenerateContainerIDAndName(containerName string) (string, string) {
	id := randStringBytes(10)
	if containerName == "" {
		containerName = id
	}

	return id, containerName
}

func RecordContainerInfo(image string, containerID string, containerPID int, commandArray []string, containerName string) error {

	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")

	containerInfo := &ContainerInfo{
		Id:         containerID,
		Pid:        strconv.Itoa(containerPID),
		Image:      image,
		Command:    command,
		CreateTime: createTime,
		Status:     RUNNING,
		Name:       containerName,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("record container info error: %v", err)
		return err
	}

	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerID)
	if err := os.Mkdir(dirUrl, 0622); err != nil && !os.IsExist(err) {
		log.Errorf("mkdir error: %v", err)
		return err
	}

	fileName := dirUrl + ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("create file error: %v", err)
		return err
	}

	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("file write string error %v", err)
		return err
	}

	return nil
}

func ListContainers() {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, "")
	dirUrl = dirUrl[:len(dirUrl)-1]

	files, err := ioutil.ReadDir(dirUrl)
	if err != nil {
		log.Errorf("read dir %s error:%v", dirUrl, err)
		return
	}

	var containers []*ContainerInfo

	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			log.Errorf("get container info error: %v", err)
			continue
		}

		containers = append(containers, tmpContainer)
	}

	printContainerInfoTable(containers)

}

func getContainerInfo(file os.FileInfo) (*ContainerInfo, error) {
	containerID := file.Name()

	configFileDir := getContainerConfigFilePath(containerID)

	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		log.Errorf("read file %s error: %v", configFileDir, err)
		return nil, err
	}

	var containerInfo ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("json unmarshal error: %v", err)
		return nil, err
	}

	return &containerInfo, nil
}

func printContainerInfoTable(containers []*ContainerInfo) {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "CONTAINER ID\tIMAGE\tName\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Image,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreateTime)
	}

	if err := w.Flush(); err != nil {
		log.Errorf("flush error: %v", err)
	}

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

	pid, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		log.Errorf("convert pid from string to int error: %v", err)
		return
	}

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		log.Errorf("stop container %s error: %v", containerID, err)
		return
	}

	changeContainerInfoForStop(containerInfo)

}

func StopContainerForTty(containerID string) {
	containerInfo, err := GetContainerInfoById(containerID)
	if err != nil {
		log.Errorf("get container info of %s error: %v", containerID, err)
		return
	}

	changeContainerInfoForStop(containerInfo)
}

func changeContainerInfoForStop(containerInfo *ContainerInfo) {
	containerInfo.Status = STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("json marshal container %s error: %v", containerInfo.Id, err)
		return
	}

	configFilePath := getContainerConfigFilePath(containerInfo.Id)
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		log.Errorf("write file %s error: %v", configFilePath, err)
	}
}

func GetContainerInfoById(containerID string) (*ContainerInfo, error) {
	configFilePath := getContainerConfigFilePath(containerID)

	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("read file %s error: %v", configFilePath, err)
		return nil, err
	}

	var containerInfo ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		log.Errorf("get container info by id, unmarshal error: %v", err)
		return nil, err
	}

	return &containerInfo, nil
}

func RemoveContainer(containerID string) {
	containerInfo, err := GetContainerInfoById(containerID)
	if err != nil {
		log.Errorf("get container %s info error %v", containerID, err)
		return
	}

	if containerInfo.Status != STOP {
		log.Errorf("could not remove running container")
		return
	}

	// remove container config files
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerID)
	if err = os.RemoveAll(dirUrl); err != nil {
		log.Errorf("remove file %s error: %v", dirUrl, err)
		return
	}

	// remove container write layer
	if err = DeleteWorkSpace(containerID); err != nil {
		log.Errorf("remove workspace of container %s error: %v", containerID, err)
		return
	}

	// remove container cgroups
	cgroupManager := cgroups.NewCgroupManager(fmt.Sprintf(cgroups.CgroupPathFormat, containerID))
	if err = cgroupManager.Destroy(); err != nil {
		log.Errorf("remove cgroup of container %s error: %v", containerID, err)
	}

}
