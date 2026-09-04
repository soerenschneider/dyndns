package verification

import "github.com/soerenschneider/dyndns/v2/pkg/update"

type Edge struct {
	Host string
}

func (e Edge) Sign(ip update.DnsRecord) string {
	return "edge"
}

func (e Edge) Verify(signature string, ip update.DnsRecord) bool {
	return signature == "edge" && e.Host == ip.Host
}
