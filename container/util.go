package container

import (
	"fmt"
	"math/rand"
	"time"
)

const letterBytes = "1234567890"

func randStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}

func getContainerConfigFilePath(containerID string) string {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerID)
	configFilePath := dirUrl + ConfigName

	return configFilePath
}
