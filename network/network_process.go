package network

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/ForeverSRC/MyDocker/container"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func CreateNetwork(driver, subnet, name string) error {
	_, exist := networks[name]
	if exist {
		return fmt.Errorf("network with name %s already exists", name)
	}

	// 将网段字符串转换为net.IPNet
	_, cider, _ := net.ParseCIDR(subnet)

	gatewayIp, err := ipAllocator.CreateSubnet(cider)
	if err != nil {
		return err
	}

	nw, err := drivers[driver].Create(subnet, gatewayIp.To4().String(), name)

	if err != nil {
		return err
	}

	return nw.dump(defaultNetworkPath)
}

func Connect(networkName string, cinfo *container.ContainerInfo) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("not such network %s", networkName)
	}

	ip, err := ipAllocator.Allocate(network.Subnet)
	log.Infof("ip allocated for container %s is: %s", cinfo.Id, ip.String())
	if err != nil {
		return err
	}

	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}

	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}

	if err = configEndpointIpAddrAndRoute(ep, cinfo); err != nil {
		return err
	}

	return configPortMapping(ep)

}

func configEndpointIpAddrAndRoute(ep *Endpoint, cinfo *container.ContainerInfo) error {
	vethPeerName := ep.Device.PeerName
	peerLink, err := netlink.LinkByName(vethPeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	// 将容器的网络端点加入到容器的网络空间中
	// 并使当前函数下面的操作都在此网络空间中进行，当前函数执行完毕后，恢复为默认的网络空间
	defer enterContainerNetns(&peerLink, cinfo)()

	// 获取到容器的IP地址及网段，用于配置容器内部接口地址
	interfaceIP := ep.Network.getIPNet()
	interfaceIP.IP = ep.IPAddress

	if err = setInterfaceIP(vethPeerName, interfaceIP); err != nil {
		return fmt.Errorf("set network %s interface ip [%s] error: %v", ep.Network.Name, interfaceIP.IP.String(), err)
	}

	// 启动容器内的veth端点
	if err = setInterfaceUP(vethPeerName); err != nil {
		return err
	}

	// net ns中默认本地地址127.0.0.1的lo网卡默认关闭，需要启动
	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	// 设置容器内的外部请求都通过容器内的veth端点访问
	// 0.0.0.0/0 表示所有的ip地址

	_, cider, _ := net.ParseCIDR("0.0.0.0/0")

	// 构建需要添加的路由数据
	// bash: route add  -net 0.0.0.0/0 gw {bridge addr} dev {veth in container}
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        net.ParseIP(ep.Network.Gateway),
		Dst:       cider,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil

}

func enterContainerNetns(enLink *netlink.Link, cinfo *container.ContainerInfo) func() {
	// /proc/[pid]/ns/net 打开此文件描述符，即可操作 Net Namespace
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("error get container net namespace, %v", err)
	}

	nsFD := f.Fd()

	// 锁定当前程序所执行的线程，如果不锁定操作系统线程
	// goroutine可能会被调度到别的线程上，无法保证一直在所需的net namespace中
	runtime.LockOSThread()

	// 修改veth的另一端，将其移动到容器进程的net ns中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		log.Errorf("error set link netns: %v", err)
	}

	origns, err := netns.Get()
	if err != nil {
		log.Errorf("error get current netns: %v", err)
	}

	// setns
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		log.Errorf("error set netns: %v", err)
	}

	return func() {
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}

}

func configPortMapping(ep *Endpoint) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			log.Errorf("port mapping format error, %v", pm)
			continue
		}

		iptableCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptableCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			log.Errorf("iptables output %v", output)
			continue
		}

	}

	return nil
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tSubnet\tGateway\tDriver\n")

	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			nw.Name,
			nw.Subnet,
			nw.Gateway,
			nw.Driver,
		)
	}

	if err := w.Flush(); err != nil {
		log.Errorf("flush error: %v", err)
	}

}

func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	if err := ipAllocator.Delete(nw.Subnet); err != nil {
		return fmt.Errorf("error remove network driver: %v", err)
	}

	if err := drivers[nw.Driver].Delete(nw); err != nil {
		return fmt.Errorf("error remove network driver %s error: %v", nw.Driver, err)
	}

	return nw.remove(defaultNetworkPath)
}
