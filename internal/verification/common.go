package verification

import (
	"github.com/soerenschneider/dyndns/internal/common"
	"encoding/base64"
)

func DecodeBase64(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

func EncodeBase64(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}

type SignatureKeypair interface {
	Sign(ip common.ResolvedIp) string
	VerificationKey
}

type VerificationKey interface {
	Verify(signature string, ip common.ResolvedIp) bool
}
