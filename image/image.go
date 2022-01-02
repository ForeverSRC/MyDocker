package image

type ImageConfig struct {
	Repository string   `json:"repository"`
	Os         string   `json:"os"`
	Arch       string   `json:"arch"`
	Tag        string   `json:"tag"`
	Layers     []string `json:"layers"`
}

type ImageRepositories struct {
	Repositories map[string]map[string]string `json:"repositories"`
}

const ImageRepositoriesPath = "/root/my-docker/image/aufs/repositories.json"

const ImageRootPath = "/root/my-docker/image/aufs/imagedb/content/sha256/"
