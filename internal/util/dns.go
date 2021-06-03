package util

import (
	"fmt"
	"log"
	"net"
)

func HostnameMatchesIp(host, ipv4, ipv6 string) bool {
	ips, err := LookupDns(host)
	if err != nil {
		log.Printf("Error looking up dns record %s: %v", host, err)
		return false
	}

	for _, hostIp := range ips {
		if hostIp == ipv4 || hostIp == ipv6 {
			log.Printf("DNS record %s verified", host)
			return true
		}
	}

	return false
}

func LookupDns(host string) ([]string, error) {
	response, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("could not resolve host %s: %v", host, err)
	}

	var ips []string
	for _, ip := range response {
		ips = append(ips, ip.String())
	}

	return ips, nil
}
