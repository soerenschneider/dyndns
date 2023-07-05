package verification

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
)

type Ed25519Keypair struct {
	PubKey     ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

func NewKeyPair() (*Ed25519Keypair, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	return &Ed25519Keypair{
		PubKey:     pub,
		privateKey: priv,
	}, nil
}

func PubkeyFromString(pub string) (VerificationKey, error) {
	raw, err := DecodeBase64(pub)
	if err != nil {
		return nil, err
	}

	return &Ed25519Keypair{
		PubKey: raw,
	}, nil
}

func (keypair *Ed25519Keypair) Verify(signature string, ip common.ResolvedIp) bool {
	signatureRaw, err := DecodeBase64(signature)
	if err != nil {
		return false
	}

	return ed25519.Verify(keypair.PubKey, []byte(ip.Hash()), signatureRaw)
}

func (keypair *Ed25519Keypair) Sign(ip common.ResolvedIp) string {
	if nil == keypair.privateKey {
		return ""
	}

	signature := ed25519.Sign(keypair.privateKey, []byte(ip.Hash()))
	return base64.StdEncoding.EncodeToString(signature)
}

func (keypair *Ed25519Keypair) AsJson() ([]byte, error) {
	serialized := serializedKeypair{
		PubKey:     EncodeBase64(keypair.PubKey),
		PrivateKey: EncodeBase64(keypair.privateKey),
	}

	return json.Marshal(serialized)
}

type serializedKeypair struct {
	PubKey     string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func FromFile(path string) (*Ed25519Keypair, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var conf serializedKeypair
	err = json.Unmarshal(content, &conf)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal json to config: %v", err)
	}

	priv, err := DecodeBase64(conf.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode private key: %v", err)
	}

	pub, err := DecodeBase64(conf.PubKey)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode public key: %v", err)
	}

	log.Info().Msgf("Read keypair with pub key '%s' from %s", conf.PubKey, path)
	return &Ed25519Keypair{
		PubKey:     pub,
		privateKey: priv,
	}, nil
}

func WriteToFile(path string, keypair *Ed25519Keypair) error {
	marshalled, err := keypair.AsJson()
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, marshalled, 0600); err != nil {
		return fmt.Errorf("can not write config to path %s: %v", path, err)
	}

	return nil
}
