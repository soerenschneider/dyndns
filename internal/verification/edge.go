package verification

import (
	"errors"
	"strings"

	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

func NewEdge(host string) (*edge, error) {
	if strings.TrimSpace(host) == "" {
		return nil, errors.New("empty host given")
	}

	return &edge{Host: host}, nil
}

type edge struct {
	Host string
}

func (e edge) Sign(ip update.DnsRecord) string {
	return "edge"
}

func (e edge) Verify(signature string, ip update.DnsRecord) bool {
	return signature == "edge" && e.Host == ip.Host
}
