package common

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"time"
)

type Envelope struct {
	PublicIp  ResolvedIp `json:"public_ip"`
	Signature string     `json:"signature"`
}

func (env *Envelope) Validate() error {
	if len(env.Signature) == 0 {
		return errors.New("signature is missing")
	}

	return env.PublicIp.Validate()
}

type ResolvedIp struct {
	IpV4      string    `json:"ipv4,omitempty"`
	IpV6      string    `json:"ipv6,omitempty"`
	Host      string    `json:"host"`
	Timestamp time.Time `json:"timestamp"`
}

func NewResolvedIp(host string) *ResolvedIp {
	return &ResolvedIp{
		Host:      host,
		Timestamp: time.Now(),
	}
}

// Equals checks for equality and ignores timestamps
func (resolved *ResolvedIp) Equals(ip *ResolvedIp) bool {
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

func (resolved *ResolvedIp) HasIpV6() bool {
	return len(resolved.IpV6) > 0
}

func (resolved *ResolvedIp) HasIpV4() bool {
	return len(resolved.IpV4) > 0
}

func (resolved *ResolvedIp) IsValid() bool {
	validIpV4 := false
	if resolved.HasIpV4() {
		addr := net.ParseIP(resolved.IpV4)
		validIpV4 = addr != nil
		if !validIpV4 {
			resolved.IpV4 = ""
		}
	}

	validIpV6 := false
	if resolved.HasIpV6() {
		addr := net.ParseIP(resolved.IpV6)
		validIpV6 = addr != nil
		if !validIpV6 {
			resolved.IpV6 = ""
		}
	}

	return validIpV4 || validIpV6
}

func (resolved *ResolvedIp) getFormattedDate() string {
	return resolved.Timestamp.Format(time.RFC3339)
}

func (resolved *ResolvedIp) String() string {
	if resolved.HasIpV4() && resolved.HasIpV6() {
		return fmt.Sprintf("%s: %s (v4), %s (v6)", resolved.Host, resolved.IpV4, resolved.IpV6)
	}

	if resolved.HasIpV4() {
		return fmt.Sprintf("%s: %s (v4)", resolved.Host, resolved.IpV4)
	}

	return fmt.Sprintf("%s: %s (v6)", resolved.Host, resolved.IpV6)
}

func (resolved *ResolvedIp) Hash() string {
	value := fmt.Sprintf("%d%s%s%s", resolved.Timestamp.Unix(), resolved.Host, resolved.IpV4, resolved.IpV6)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}

func (resolved ResolvedIp) Validate() error {
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
