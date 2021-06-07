package verification

import (
	"crypto/ed25519"
	"crypto/rand"
	"dyndns/internal/common"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
)

type Ed25519Keypair struct {
	PubKey     ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

// TODO: Test with garbage input
func PubkeyFromString(pub string) (VerificationKey, error) {
	raw, err := DecodeBase64(pub)
	if err != nil {
		return nil, err
	}

	return &Ed25519Keypair{
		PubKey: raw,
	}, nil
}

type serializedKeypair struct {
	PubKey     string `json:"public_key"`
	PrivateKey string `json:"private_key"`
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

func FromFile(path string) (*Ed25519Keypair, error) {
	content, err := ioutil.ReadFile(path)
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
	} else {
		log.Info().Msgf("Read keypair with pub key %s from %s", conf.PubKey, path)
	}

	return &Ed25519Keypair{
		PubKey:     pub,
		privateKey: priv,
	}, nil
}

func ToFile(path string, keypair *Ed25519Keypair) error {
	serialized := serializedKeypair{
		PubKey:     EncodeBase64(keypair.PubKey),
		PrivateKey: EncodeBase64(keypair.privateKey),
	}

	marshalled, err := json.Marshal(serialized)
	if err != nil {
		return fmt.Errorf("could not unmarshal json to config: %v", err)
	}

	err = ioutil.WriteFile(path, marshalled, 0640)
	if err != nil {
		return fmt.Errorf("can not write config to path %s: %v", path, err)
	}

	return nil
}

func NewKeyPair() (*Ed25519Keypair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Ed25519Keypair{
		PubKey:     pub,
		privateKey: priv,
	}, nil
}
