package network

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var (
	drivers  = map[string]NetworkDriver{}
	networks = map[string]*Network{}
)

const defaultNetworkPath = "/root/my-docker/network/network/"

func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(defaultNetworkPath, 0644)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err := filepath.Walk(defaultNetworkPath, func(nwPath string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}

		if err := nw.load(nwPath); err != nil {
			log.Errorf("error load network: %v", err)
			return err
		}

		networks[nwName] = nw
		drivers[nw.Driver].Recover(nw)
		return nil
	})

	return err
}
