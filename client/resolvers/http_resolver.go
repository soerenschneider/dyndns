package resolvers

import (
	"dyndns/internal/common"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	retries        = 3
	timeout        = 2 * time.Second
	AddrFamilyIpv6 = "ip6"
	AddrFamilyIpv4 = "ip4"
)

var providers = []string{
	"https://icanhazip.com",
	"https://ifconfig.me",
	"https://ifconfig.co",
	"https://ipinfo.io/ip",
	"https://api.ipify.org",
	"https://ipecho.net/plain",
	"https://checkip.amazonaws.com",
}

var (
	serverAddresses = map[string]string{
		AddrFamilyIpv4: "8.8.8.8:53",
		AddrFamilyIpv6: "[2001:4860:4860::8888]:53",
	}
)

type HttpResolver struct {
	client         *http.Client
	host           string
	localAddresses map[string]string
}

func NewHttpResolver(domain string) (*HttpResolver, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retries

	standardClient := retryClient.StandardClient()
	standardClient.Timeout = timeout

	return &HttpResolver{client: standardClient, host: domain}, nil
}

func getLocalAddress(serverAddr string) (net.Addr, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.TCPAddr)
	return &net.TCPAddr{IP: localAddr.IP, Port: 0, Zone: ""}, nil
}

func (resolver *HttpResolver) Host() string {
	return resolver.host
}

func (resolver *HttpResolver) Name() string {
	return "HttpResolver"
}

func (resolver *HttpResolver) Resolve() (*common.ResolvedIp, error) {
	shuffleProviders()
	detectedIps := &common.ResolvedIp{
		Host:      resolver.host,
		Timestamp: time.Now(),
	}

	for addressFamily, serverAddress := range serverAddresses {
		localAddr, err := getLocalAddress(serverAddress)
		if err != nil {
			// system has no support for addressFamily apparently, continue
			continue
		}
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				LocalAddr: localAddr,
			}).DialContext,
		}
		if transport == nil {
			continue
		}
		resolver.client.Transport = transport
		for _, url := range providers {
			detectedIp, err := resolveSingle(url, resolver.client)
			if err == nil {
				// Check if the resolved IP is actually a valid IP
				if net.ParseIP(detectedIp) == nil {
					continue
				}

				// Set the correct address family and stop iterating providers
				if addressFamily == AddrFamilyIpv6 {
					detectedIps.IpV6 = detectedIp
				} else {
					detectedIps.IpV4 = detectedIp
				}
				break
			} else {
				log.Printf("Error while resolving IP: %v", err)
			}
		}
	}

	return detectedIps, nil
}

func shuffleProviders() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(providers), func(i, j int) { providers[i], providers[j] = providers[j], providers[i] })
}

func resolveSingle(url string, client *http.Client) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("error talking to '%s': %v", url, err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't read response: %v", err)
	}
	detectedIp := repair(string(body))
	return detectedIp, nil
}

func repair(body string) string {
	body = strings.TrimSuffix(body, "\n")
	return body
}
