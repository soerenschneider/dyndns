package resolvers

import (
	"dyndns/internal/common"
	"errors"
	"fmt"
	"net"
	"time"
)

type InterfaceResolver struct {
	watchedInterface string
	host             string
}

func NewInterfaceResolver(watchedInterface, host string) (*InterfaceResolver, error) {
	return &InterfaceResolver{
		watchedInterface: watchedInterface,
		host:             host,
	}, nil
}

func (resolver *InterfaceResolver) Name() string {
	return fmt.Sprintf("InterfaceResolver (%s)", resolver.watchedInterface)
}

func (resolver *InterfaceResolver) Host() string {
	return resolver.host
}

func (resolver *InterfaceResolver) Resolve() (*common.ResolvedIp, error) {
	ipv4, err := GetInterfaceIpv4Addr(resolver.watchedInterface)
	if err != nil {
		return nil, fmt.Errorf("could not resolve ip for interface: %v", err)
	}

	return &common.ResolvedIp{
		IpV4:      ipv4,
		Host:      resolver.host,
		Timestamp: time.Now(),
	}, nil
}

func GetInterfaceIpv4Addr(interfaceName string) (addr string, err error) {
	var (
		ief       *net.Interface
		addresses []net.Addr
		ipv4Addr  net.IP
	)

	if ief, err = net.InterfaceByName(interfaceName); err != nil { // get interface
		return
	}

	if addresses, err = ief.Addrs(); err != nil { // get addresses
		return
	}

	for _, addr := range addresses { // get ipv4 address
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			return ipv4Addr.String(), nil
		}
	}

	return "", errors.New(fmt.Sprintf("interface %s doesn't have an ipv4 address\n", interfaceName))
}
