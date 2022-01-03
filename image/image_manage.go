package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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

func CommitNewImage(repo string, tag string, imageID string, config string) error {
	dirUrl := ImageRootPath + imageID + "/"
	if err := os.Mkdir(dirUrl, 0622); err != nil {
		return err
	}

	configFileName := dirUrl + ImageConfigFileName

	file, err := os.Create(configFileName)
	defer file.Close()
	if err != nil {
		return err
	}

	if _, err := file.WriteString(config); err != nil {
		return err
	}

	return updateRepositories(repo, tag, imageID)

}

func updateRepositories(repo string, tag string, imageID string) error {
	repositories, err := getRepositories()
	if err != nil {
		return err
	}

	repoInfo, ok := repositories.Repositories[repo]
	if ok {
		repoInfo[repo+":"+tag] = "sha256:" + imageID
	} else {
		tmpMap := make(map[string]string)
		tmpMap[repo+":"+tag] = "sha256:" + imageID
		repositories.Repositories[repo] = tmpMap
	}

	reposJsonByte, err := json.Marshal(repositories)
	if err != nil {
		return fmt.Errorf("marshal repositories json error: %v", err)
	}

	err = ioutil.WriteFile(ImageRepositoriesPath, reposJsonByte, 0766)
	if err != nil {
		return fmt.Errorf("write to %s error: %v", err)
	}

	return nil
}
