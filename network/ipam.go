package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

const ipamDefaultAllocatorPath = "/root/my-docker/network/ipam/subnet.json"

type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}

	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		log.Errorf("error dump allocation info, %v", err)
		return err
	}
	return nil

}

func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(ipamConfigFileDir, 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}

	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}

	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}

	return nil
}

func (ipam *IPAM) CreateSubnet(subnet *net.IPNet) (net.IP, error) {
	ipam.Subnets = &map[string]string{}
	err := ipam.load()
	if err != nil {
		log.Errorf("error load allocation info: %v", err)
		return nil, err
	}

	ones, size := subnet.Mask.Size()

	_, exist := (*ipam.Subnets)[subnet.String()]
	if exist {
		return nil, fmt.Errorf("pool overlaps with other one on this address space: %s", subnet.String())
	}

	// 如果之前没有分配过指定网段，则初始化网段的分配配置
	// 用 0 填满网段配置，1<<uint8(size-ones)表示网段中的可用地址数目
	// size-ones表示子网掩码后买呢的网络位数，2^(size-ones)表示可用ip数目
	(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-ones))

	return ipam.Allocate(subnet)

}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (net.IP, error) {
	var ip net.IP
	for idx, ch := range (*ipam.Subnets)[subnet.String()] {
		if ch == '0' {
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[idx] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)

			// 初始ip
			ip = subnet.IP

			// 计算当前数组偏移量对应的ip
			// 示例：
			// 原始数组[172，16，0，0]，偏移量idx=65555
			// 则需要在各个部分依次加上 [uint8(65555>>24),uint8(65555>>16)，uint8(65555>>8)，uint8(65555>>0)]
			// 结果为[0,1,0,19] 则偏移后的ip为[172,17,0,19]
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(idx >> ((t - 1) * 8))
			}

			// 计算下一个可用的ip
			ip[3] += 1
			break
		}
	}

	err := ipam.dump()
	if err != nil {
		log.Errorf("error dump allocation info: %v", err)
		return nil, err
	}

	return ip, nil
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipAddr *net.IP) error {
	ipam.Subnets = &map[string]string{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		log.Errorf("error dump allocation info %v", err)
	}

	c := 0
	// 4字节表示方式
	releaseIP := ipAddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1] - subnet.IP[t-1]) << ((4-t) * 8)
	}

	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	err = ipam.dump()
	if err != nil {
		log.Errorf("error dump allocation info: %v", err)
		return err
	}

	return nil
}
