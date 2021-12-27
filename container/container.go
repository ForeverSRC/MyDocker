package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"
)

type ContainerInfo struct {
	Pid        string `json:"pid"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	Command    string `json:"command"`
	CreateTime string `json:"createTime"`
	Status     string `json:"status"`
}

const (
	RUNNING = "running"
	STOP    = "stopped"
	EXIT    = "exited"
)

const (
	DefaultInfoLocation = "/root/my-docker/containers/%s/"
	ConfigName          = "config.json"
	ContainerLogFile    = "container.log"
)

func GenerateContainerIDAndName(containerName string) (string, string) {
	id := randStringBytes(10)
	if containerName == "" {
		containerName = id
	}

	return id, containerName
}

func RecordContainerInfo(containerID string, containerPID int, commandArray []string, containerName string) error {

	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")

	containerInfo := &ContainerInfo{
		Id:         containerID,
		Pid:        strconv.Itoa(containerPID),
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

	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.Mkdir(dirUrl, 0622); err != nil {
		log.Errorf("mkdir error: %v", err)
		return err
	}

	fileName := dirUrl + "/" + ConfigName
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

func DeleteContainerInfo(containerId string) {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Errorf("remove dir %s error %v", dirUrl, err)
	}
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
	containerName := file.Name()

	configFileDir := fmt.Sprintf(DefaultInfoLocation, containerName)
	configFileDir = configFileDir + ConfigName

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
	fmt.Fprint(w, "ID\tName\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
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

func LogContainer(containerName string) {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
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
