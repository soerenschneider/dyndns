package common

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"time"
)

type UpdateRecordRequest struct {
	PublicIp  DnsRecord `json:"public_ip"`
	Signature string    `json:"signature"`
}

func (r *UpdateRecordRequest) Validate() error {
	if len(r.Signature) == 0 {
		return errors.New("signature is missing")
	}

	return r.PublicIp.Validate()
}

type DnsRecord struct {
	IpV4      string    `json:"ipv4,omitempty"`
	IpV6      string    `json:"ipv6,omitempty"`
	Host      string    `json:"host"`
	Timestamp time.Time `json:"timestamp"`
}

func NewResolvedIp(host string) *DnsRecord {
	return &DnsRecord{
		Host:      host,
		Timestamp: time.Now(),
	}
}

// Equals checks for equality and ignores timestamps
func (resolved *DnsRecord) Equals(ip *DnsRecord) bool {
	if ip == nil || resolved == nil {
		return false
	}

	if resolved.IpV4 != ip.IpV4 {
		return false
	}

	if resolved.IpV6 != ip.IpV6 {
		return false
	}

	if resolved.Host != ip.Host {
		return false
	}

	return true

}

func (resolved *DnsRecord) HasIpV6() bool {
	return len(resolved.IpV6) > 0
}

func (resolved *DnsRecord) HasIpV4() bool {
	return len(resolved.IpV4) > 0
}

func (resolved *DnsRecord) IsValid() bool {
	if !resolved.HasIpV4() {
		return false
	}
	addr := net.ParseIP(resolved.IpV4)
	if addr == nil {
		return false
	}

	if addr.IsPrivate() {
		return false
	}

	if addr.IsLoopback() {
		return false
	}

	return true
}

func (resolved *DnsRecord) String() string {
	if resolved.HasIpV4() && resolved.HasIpV6() {
		return fmt.Sprintf("%s: %s (v4), %s (v6)", resolved.Host, resolved.IpV4, resolved.IpV6)
	}

	if resolved.HasIpV4() {
		return fmt.Sprintf("%s: %s (v4)", resolved.Host, resolved.IpV4)
	}

	return fmt.Sprintf("%s: %s (v6)", resolved.Host, resolved.IpV6)
}

func (resolved *DnsRecord) Hash() string {
	value := fmt.Sprintf("%d%s%s%s", resolved.Timestamp.Unix(), resolved.Host, resolved.IpV4, resolved.IpV6)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}

func (resolved *DnsRecord) Validate() error {
	if len(resolved.Host) == 0 {
		return errors.New("domain is missing")
	}

	if len(resolved.IpV4) == 0 && len(resolved.IpV6) == 0 {
		return errors.New("both ipv4 and ipv6 are empty")
	}

	if resolved.Timestamp.IsZero() {
		return fmt.Errorf("timestamp empty")
	}

	return nil
}
