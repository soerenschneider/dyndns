package verification

import (
	"encoding/base64"

	"github.com/soerenschneider/dyndns/internal/common"
)

func DecodeBase64(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

func EncodeBase64(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}

type SignatureKeypair interface {
	AsJson() ([]byte, error)
	Sign(ip common.DnsRecord) string
	VerificationKey
}

type VerificationKey interface {
	Verify(signature string, ip common.DnsRecord) bool
}
