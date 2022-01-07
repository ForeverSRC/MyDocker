package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct {
}

func (b *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (b *BridgeNetworkDriver) Create(subnet string, gatewayIP string, name string) (*Network, error) {
	n := &Network{
		Name:    name,
		Driver:  b.Name(),
		Subnet:  subnet,
		Gateway: gatewayIP,
	}

	err := b.initBridge(n)
	if err != nil {
		log.Errorf("error init bridge: %v", err)
	}

	return n, err
}

func (b *BridgeNetworkDriver) Delete(network *Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	return netlink.LinkDel(br)
}

func (b *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]

	// 设置veth的一端挂载到网络对应的Linux Bridge上
	la.MasterIndex = br.Attrs().Index
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}

	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("error add endpoint device: %v", err)
	}

	// bash: ip link set xxx up
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("error set endpoint device up: %v", err)
	}

	return nil

}

func (b *BridgeNetworkDriver) DisConnect(network *Network, endpoint *Endpoint) error {
	return nil
}

func (b *BridgeNetworkDriver) initBridge(n *Network) error {
	//1.创建bridge虚拟设备
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("error add bridge %s, error: %v", bridgeName, err)
	}

	//2.设置bridge设备的路由和地址
	if err := setInterfaceIP(bridgeName, n.getIPNet()); err != nil {
		return fmt.Errorf("error assigning address %s on bridge %s with an error of: %v", n.Subnet, bridgeName, err)
	}

	//3.启动bridge设备
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("error set bridge %s up, error: %v", bridgeName, err)
	}

	//4.设置iptables的SNAT规则
	if err := setupIpTables(bridgeName, n.Subnet); err != nil {
		return fmt.Errorf("error setting iptables for %s, error: %v", bridgeName, err)
	}

	return nil
}

// createBridgeInterface 创建linux Bridge设备
func createBridgeInterface(bridgeName string) error {

	_, err := net.InterfaceByName(bridgeName)
	interfaceExist := err == nil
	if interfaceExist || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	br := &netlink.Bridge{
		LinkAttrs: la,
	}

	// 添加bridge网络设备
	if err = netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge creation failed for bridge %s: %v", bridgeName, err)
	}

	return nil
}

// setInterfaceIP 设置一个网络接口的IP地址
func setInterfaceIP(name string, ipNet *net.IPNet) error {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("error get interface: %v", err)
	}

	addr := &netlink.Addr{
		IPNet: ipNet,
		Label: "",
		Flags: 0,
		Scope: 0,
	}

	// bash: ip addr add xxx
	return netlink.AddrAdd(iface, addr)
}

// setInterfaceUP 设置网络借口状态为UP
func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)

	if err != nil {
		return fmt.Errorf("error retrieving a link named [%s]: %v", iface.Attrs().Name, err)
	}

	// 启用网络设备
	// bash: ip link set xxx up
	if err = netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("error enableing interface for %s: %v", interfaceName, err)
	}

	return nil
}

// setupIpTables 设置iptables对应bridge的MASQUERADE规则
func setupIpTables(bridgeName string, subnet string) error {
	// bash: iptables -t nat -A POSTROUTING -s <subnetName> ! -o <bridgeName> -j MASQUERADE
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet, bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		log.Errorf("iptables output: %v", output)
	}

	return nil
}
