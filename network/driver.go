package network

type NetworkDriver interface {
	Name() string
	Create(subnet string, gatewayIP string, name string) (*Network, error)
	Delete(network *Network) error
	Connect(network *Network, endpoint *Endpoint) error
	DisConnect(network *Network, endpoint *Endpoint) error
}
