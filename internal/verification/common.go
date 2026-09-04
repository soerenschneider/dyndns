package verification

import (
	"encoding/base64"

	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

func DecodeBase64(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

func EncodeBase64(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}

type SignatureKeypair interface {
	Sign(ip update.DnsRecord) string
}

type VerificationKey interface {
	Verify(signature string, ip update.DnsRecord) bool
}
