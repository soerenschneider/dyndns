package resolvers

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/conf"
	"github.com/soerenschneider/dyndns/internal/metrics"
)

const (
	retries = 3
	timeout = 2 * time.Second
)

var (
	serverAddresses = map[string]string{
		conf.AddrFamilyIpv4: "8.8.8.8:53",
		conf.AddrFamilyIpv6: "[2001:4860:4860::8888]:53",
	}
)

type HttpResolver struct {
	client             *http.Client
	host               string
	preferredProviders []string
	backupProviders    []string
	providers          []string
	addressFamilies    []string
	random             *rand.Rand
}

func NewHttpResolver(domain string, preferredUrls []string, fallbackUrls []string, addressFamilies []string) (*HttpResolver, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retries

	standardClient := retryClient.StandardClient()
	standardClient.Timeout = timeout

	if fallbackUrls == nil {
		fallbackUrls = make([]string, 0)
	}

	if len(preferredUrls)+len(fallbackUrls) == 0 {
		return nil, errors.New("neither preferred- nor fallback-urls provided")
	}

	if addressFamilies == nil {
		return nil, errors.New("empty addressFamily slice provided")
	}

	resolver := &HttpResolver{
		host:               domain,
		client:             standardClient,
		preferredProviders: preferredUrls,
		backupProviders:    fallbackUrls,
		addressFamilies:    addressFamilies,
	}
	resolver.providers = make([]string, len(preferredUrls)+len(fallbackUrls))
	return resolver, nil
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

func (resolver *HttpResolver) Resolve() (*common.DnsRecord, error) {
	resolver.shuffleProviders()
	detectedIps := &common.DnsRecord{
		Host:      resolver.host,
		Timestamp: time.Now(),
	}

	for _, addressFamily := range resolver.addressFamilies {
		serverAddress, ok := serverAddresses[addressFamily]
		if !ok {
			// TODO: add metric?
			log.Warn().Str("component", "http_resolver").Str("address_family", addressFamily).Msg("unknown address family, check your configuration")
			continue
		}

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

		resolver.client.Transport = transport
		for index, url := range resolver.providers {
			detectedIp, err := resolveSingle(url, resolver.client)
			if err == nil {
				// Check if the resolved IP is actually a valid IP
				if net.ParseIP(detectedIp) == nil {
					log.Error().Str("component", "http_resolver").Str("detected_ip", detectedIp).Msg("could not parse detected IP address")
					metrics.InvalidResolvedIps.WithLabelValues(resolver.Host(), resolver.Name(), url).Inc()
					continue
				}

				// Set the correct address family and stop iterating backupProviders
				if addressFamily == conf.AddrFamilyIpv6 {
					detectedIps.IpV6 = detectedIp
				} else {
					detectedIps.IpV4 = detectedIp
				}
				metrics.IpsResolved.WithLabelValues(resolver.host, resolver.Name(), url).Inc()
				break
			} else {
				metrics.IpResolveErrors.WithLabelValues(resolver.host, resolver.Name(), url).Inc()
				log.Error().Err(err).Str("component", "http_resolver").Msg("Error while resolving IP")
				if index == len(resolver.preferredProviders)-1 {
					log.Warn().Str("component", "http_resolver").Msgf("Exhausted list of preferred providers")
				}
			}
		}
	}

	return detectedIps, nil
}

func (resolver *HttpResolver) shuffleProviders() {
	resolver.random = rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404
	resolver.random.Shuffle(len(resolver.preferredProviders), func(i, j int) {
		resolver.preferredProviders[i], resolver.preferredProviders[j] = resolver.preferredProviders[j], resolver.preferredProviders[i]
	})
	resolver.random.Shuffle(len(resolver.backupProviders), func(i, j int) {
		resolver.backupProviders[i], resolver.backupProviders[j] = resolver.backupProviders[j], resolver.backupProviders[i]
	})

	for i := 0; i < len(resolver.providers); i++ {
		if i < len(resolver.preferredProviders) {
			resolver.providers[i] = resolver.preferredProviders[i]
		} else {
			resolver.providers[i] = resolver.backupProviders[i-len(resolver.preferredProviders)]
		}
	}
}

func resolveSingle(url string, client *http.Client) (string, error) {
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("error talking to '%s': %v", url, err)
	}
	timeTaken := time.Since(start)
	metrics.ResponseTime.WithLabelValues(url).Observe(timeTaken.Seconds())

	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't read response: %v", err)
	}
	detectedIp := repair(string(body))
	return detectedIp, nil
}

func repair(body string) string {
	return strings.TrimSpace(body)
}
