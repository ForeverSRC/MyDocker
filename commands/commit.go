package commands

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ForeverSRC/MyDocker/container"
	"github.com/ForeverSRC/MyDocker/image"
	"github.com/go-basic/uuid"
	"github.com/urfave/cli"
)

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container to image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing container id or image repository:tag")
		}

		args := context.Args()
		containerID := args[0]
		imgInfo := args[1]
		if !strings.ContainsRune(imgInfo, ':') {
			imgInfo = imgInfo + ":latest"
		}

		return commitContainer(containerID, imgInfo)

	},
}

func commitContainer(containerID, imageInfo string) error {
	containerInfo, err := container.GetContainerInfoById(containerID)
	if err != nil {
		return err
	}

	imageID, err := image.GetImageID(containerInfo.Image)
	if err != nil {
		return fmt.Errorf("get image ID error: %v", err)
	}

	layers, err := image.GetImageLayers(imageID)
	if err != nil {
		return err
	}

	writeURL := fmt.Sprintf(container.ContainerWriteLayerUrl, containerID)
	roURL := fmt.Sprintf(container.ContainerAUFSRootUrl, uuid.New())
	cmd := exec.Command("cp", "-r", writeURL, roURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	newLayers := make([]string, len(layers))
	copy(newLayers, layers)
	newLayers = append(newLayers, roURL)

	tmp := strings.Split(imageInfo, ":")
	repo, tag := tmp[0], tmp[1]
	imgCfg := &image.ImageConfig{
		Repository: repo,
		Os:         "linux",
		Arch:       "amd64",
		Tag:        tag,
		Layers:     newLayers,
	}

	jsonByte, err := json.Marshal(imgCfg)
	if err != nil {
		return err
	}

	newImageID := fmt.Sprintf("%x", sha256.Sum256(jsonByte))
	jsonStr := string(jsonByte)
	err = image.CommitNewImage(repo, tag, newImageID, jsonStr)
	if err != nil {
		return err
	}

	fmt.Println(newImageID)

	return nil

}
