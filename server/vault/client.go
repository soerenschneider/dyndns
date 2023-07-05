package vault

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/conf"
)

const (
	AwsIamPropagationImpediment = 20 * time.Second
)

type VaultCredentialProvider struct {
	client *api.Client
	expiry time.Time
	config *conf.VaultConfig
	auth   Auth
}

type Auth interface {
	Login(ctx context.Context, client *api.Client) (*api.Secret, error)
}

func NewVaultCredentialProvider(client *api.Client, auth Auth, conf *conf.VaultConfig) (*VaultCredentialProvider, error) {
	if client == nil {
		return nil, errors.New("empty client provided")
	}

	if auth == nil {
		return nil, errors.New("empty auth provided")
	}

	if conf == nil {
		return nil, errors.New("empty vault config provided")
	}

	return &VaultCredentialProvider{
		client: client,
		expiry: time.Now(),
		auth:   auth,
		config: conf,
	}, nil
}

func (m *VaultCredentialProvider) Retrieve() (credentials.Value, error) {
	secret, err := m.readAwsCredentials()
	if err != nil {
		return credentials.Value{}, fmt.Errorf("error getting dynamic credentials: %w", err)
	}

	m.expiry = time.Now().Add(time.Duration(secret.LeaseDuration) * time.Second)
	cred, err := parseAwsCredentialsReply(secret)
	if err != nil {
		return credentials.Value{}, err
	}

	// Wait for eventual consistency on AWS side
	time.Sleep(AwsIamPropagationImpediment)
	return cred, nil
}

func (m *VaultCredentialProvider) IsExpired() bool {
	return time.Now().Before(m.expiry)
}

func (m *VaultCredentialProvider) checkLogin() error {
	_, err := m.auth.Login(context.Background(), m.client)
	return err
}

func (m *VaultCredentialProvider) readAwsCredentials() (*api.Secret, error) {
	log.Info().Msgf("Generating dynamic AWS credentials for role %s", m.config.AwsRoleName)

	_, err := m.client.Auth().Login(context.Background(), m.auth)
	if err != nil {
		return nil, fmt.Errorf("auth against vault failed: %w", err)
	}

	path := fmt.Sprintf("%s/creds/%s", m.config.AwsMountPath, m.config.AwsRoleName)
	return m.client.Logical().Read(path)

}

func parseAwsCredentialsReply(secret *api.Secret) (credentials.Value, error) {
	ret := credentials.Value{ProviderName: "vault"}

	if secret == nil {
		return ret, errors.New("empty secret response provided")
	}

	var ok bool
	ret.AccessKeyID, ok = secret.Data["access_key"].(string)
	if !ok {
		return ret, errors.New("could not convert 'access_key' to string")
	}

	ret.SecretAccessKey, ok = secret.Data["secret_key"].(string)
	if !ok {
		return ret, errors.New("could not convert 'secret_key' to string")
	}

	ret.SessionToken, ok = secret.Data["security_token"].(string)
	if !ok {
		return ret, errors.New("could not convert 'security_token' to string")
	}

	log.Info().Msgf("Received credentials with id %s, valid for %ds", ret.AccessKeyID, secret.LeaseDuration)
	return ret, nil
}
