package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

type ContainerInfo struct {
	Pid         string   `json:"pid"`
	Id          string   `json:"id"`
	Image       string   `json:"image"`
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	CreateTime  string   `json:"createTime"`
	Status      string   `json:"status"`
	Network     string   `json:"network"`
	IpAddress   string   `json:"ipAddress"`
	PortMapping []string `json:"portMapping"`
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

func (c *ContainerInfo) dump() error {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, c.Id)
	if _, err := os.Stat(dirUrl); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dirUrl, 0644)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	fileName := path.Join(dirUrl, ConfigName)
	configFile, err := os.OpenFile(fileName, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
	defer configFile.Close()
	if err != nil {
		return fmt.Errorf("open %s error: %v", fileName, err)
	}

	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	if _, err = configFile.Write(jsonBytes); err != nil {
		return err
	}

	return nil
}

func (c *ContainerInfo) load(dumpPath string) error {
	configJson, err := ioutil.ReadFile(dumpPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(configJson, c)
	if err != nil {
		log.Errorf("error load container info: %v", err)
		return err
	}

	return nil
}
