package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetImageLayers(image string) ([]string, error) {
	imageID, err := getImageID(image)
	if err != nil {
		return nil, fmt.Errorf("get image ID error: %v", err)
	}

	imgID := strings.Split(imageID, ":")

	imageConfigDir := ImageRootPath + imgID[1] + "/config.json"
	content, err := ioutil.ReadFile(imageConfigDir)
	if err != nil {
		return nil, fmt.Errorf("get image %s config file error: %v", imageID, err)
	}

	var imageConfig ImageConfig
	if err := json.Unmarshal(content, &imageConfig); err != nil {

		return nil, fmt.Errorf("unmarshal image %s config info error: %v", imageID, err)
	}

	return imageConfig.Layers, nil

}

func getImageID(image string) (string, error) {
	imgRepo, err := getRepositories()
	if err != nil {
		log.Errorf("get image repositories info error: %v", err)
		return "", err
	}

	info := getImageRepoNameAndTag(image)
	name := info[0]

	imageRepo, ok := imgRepo.Repositories[name]
	if !ok {
		err = fmt.Errorf("image %s not exist", name)
		return "", err
	}

	imageID, ok := imageRepo[image]
	if !ok {
		err = fmt.Errorf("image ID of %s not exist", image)
		return "", err
	}

	return imageID, nil
}

func getRepositories() (*ImageRepositories, error) {
	content, err := ioutil.ReadFile(ImageRepositoriesPath)
	if err != nil {
		return nil, err
	}

	var imageRepo ImageRepositories
	err = json.Unmarshal(content, &imageRepo)

	return &imageRepo, err
}

func getImageRepoNameAndTag(image string) []string {
	return strings.Split(image, ":")
}
