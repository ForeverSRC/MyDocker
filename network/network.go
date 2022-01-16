package network

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type Network struct {
	Name    string
	Driver  string
	Subnet  string
	Gateway string
}

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string         `json:"portMapping"`
	Network     *Network
}

func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dumpPath, 0644)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer nwFile.Close()
	if err != nil {
		log.Errorf("open %s error: %v", nwPath, err)
		return err
	}

	nwJson, err := json.Marshal(nw)
	if err != nil {
		return err
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		return err
	}

	return nil
}

func (nw *Network) load(dumpPath string) error {
	nwJson, err := ioutil.ReadFile(dumpPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(nwJson, nw)
	if err != nil {
		log.Errorf("error load nw info: %v", err)
		return err
	}

	return nil
}

func (nw *Network) remove(dumpPath string) error {
	nwPath := path.Join(dumpPath, nw.Name)
	if _, err := os.Stat(nwPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(nwPath)
	}

}

func (nw *Network) getIPNet() *net.IPNet {
	_, cider, _ := net.ParseCIDR(nw.Subnet)
	cider.IP = net.ParseIP(nw.Gateway)
	return cider
}
