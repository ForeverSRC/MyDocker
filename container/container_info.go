package container

type ContainerInfo struct {
	Pid         string   `json:"pid"`
	Id          string   `json:"id"`
	Image       string   `json:"image"`
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	CreateTime  string   `json:"createTime"`
	Status      string   `json:"status"`
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
