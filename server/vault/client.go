package vault

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/conf"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	HttpDefaultTimeout          = 10 * time.Second
	AwsIamPropagationImpediment = 20 * time.Second
	vaultTokenKey               = "X-Vault-Token"
)

type VaultCredentialProvider struct {
	vaultToken string
	client     *http.Client
	expiry     time.Time
	conf       *conf.VaultConfig
}

func NewVaultCredentialProvider(conf *conf.VaultConfig) (*VaultCredentialProvider, error) {
	if nil == conf {
		return nil, errors.New("empty config provided")
	}

	err := conf.Verify()
	if err != nil {
		return nil, err
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5

	standardClient := retryClient.StandardClient()
	standardClient.Timeout = HttpDefaultTimeout

	return &VaultCredentialProvider{
		client:     standardClient,
		expiry:     time.Now(),
		conf:       conf,
		vaultToken: conf.VaultToken,
	}, nil
}

func (m *VaultCredentialProvider) Retrieve() (credentials.Value, error) {
	err := m.checkLogin()
	if err != nil {
		return credentials.Value{}, fmt.Errorf("could not login at vault: %v", err)
	}
	dynamicCredentials, err := m.readAwsCredentials()
	if err != nil {
		log.Warn().Msgf("could not read dynamic credentials from vault: %v", err)
		return credentials.Value{}, fmt.Errorf("could not read dynamic credentials from vault: %v", err)
	}

	m.expiry = time.Now().Add(time.Duration(dynamicCredentials.LeaseDuration) * time.Second)
	cred := ConvertCredentials(dynamicCredentials.Data)

	// The credentials we receive are usually not effective at AWS, yet, so we need to wait for a bit until
	// the changes on AWS are propagated
	time.Sleep(AwsIamPropagationImpediment)
	return cred, nil
}

func (m *VaultCredentialProvider) IsExpired() bool {
	return time.Now().Before(m.expiry)
}

func (m *VaultCredentialProvider) checkLogin() error {
	err := m.LookupToken()
	if err == nil {
		return nil
	}

	if m.conf.HasAppRoleLoginInfo() {
		err = m.loginAppRole()
		if err != nil {
			return fmt.Errorf("could not login via approle: %v", err)
		}

		return m.LookupToken()
	}

	return errors.New("no more authentication methods left")
}

func buildAppRolePayload(roleId, secretId string) ([]byte, error) {
	payload := struct {
		RoleId   string `json:"role_id"`
		SecretId string `json:"secret_id"`
	}{
		RoleId:   roleId,
		SecretId: secretId,
	}
	return json.Marshal(payload)
}

type AuthReply struct {
	Renewable     bool     `json:"renewable"`
	LeaseDuration int      `json:"lease_duration"`
	Metadata      []string `json:"metadata,omitempty"`
	TokenPolicies []string `json:"token_policies"`
	Accessor      string   `json:"accessor"`
	ClientToken   string   `json:"client_token"`
}

func (m *VaultCredentialProvider) loginAppRole() error {
	encodedPayload, err := buildAppRolePayload(m.conf.AppRoleId, m.conf.AppRoleSecret)
	if err != nil {
		return fmt.Errorf("could not marshal payload: %v", err)
	}

	url := fmt.Sprintf("%s/v1/auth/approle/login", m.conf.VaultAddr)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(encodedPayload))
	if err != nil {
		return fmt.Errorf("could not build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	response, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("error while logging in via approle: %v", err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("couldn't read body: %v", err)
	}

	credentials, err := m.parseAppRoleLoginReply(body)
	if err != nil {
		return err
	}

	m.vaultToken = credentials.Auth.ClientToken
	return nil
}

func (m *VaultCredentialProvider) parseAppRoleLoginReply(body []byte) (*VaultCredentialResponse, error) {
	var response VaultCredentialResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal response: %v", err)
	}

	if len(response.Auth.ClientToken) == 0 {
		return nil, fmt.Errorf("could authenticate against vault, received empty client token")
	}

	return &response, nil
}

type TokenInfo struct {
	IssueTime   time.Time         `json:"issue_time"`
	ExpireTime  time.Time         `json:"expire_time"`
	DisplayName string            `json:"display_name"`
	Policies    []string          `json:"policies"`
	Metadata    map[string]string `json:"meta"`
	NumUses     int               `json:"num_uses"`
}

func (m *VaultCredentialProvider) LookupToken() error {
	body, err := m.lookupTokenCall()
	if err != nil {
		return fmt.Errorf("error performing lookup: %v", err)
	}

	_, err = m.parseLookupReply(body)
	return err
}

func (m *VaultCredentialProvider) lookupTokenCall() ([]byte, error) {
	url := fmt.Sprintf("%s/v1/auth/token/lookup-self", m.conf.VaultAddr)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set(vaultTokenKey, m.vaultToken)
	if err != nil {
		return nil, fmt.Errorf("could not build request: %v", err)
	}
	response, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %v", err)
	}

	if response.StatusCode > 204 {
		return nil, fmt.Errorf("bad status code: %d", response.StatusCode)
	}

	return ioutil.ReadAll(response.Body)
}

func (m *VaultCredentialProvider) parseLookupReply(body []byte) (*TokenInfo, error) {
	var wrapper struct {
		Data TokenInfo `json:"data"`
	}
	err := json.Unmarshal(body, &wrapper)
	if err != nil {
		// "not great, not terrible"
		log.Warn().Msgf("could not unmarshal response: %v", err)
		return nil, err
	}

	until := wrapper.Data.ExpireTime.Sub(time.Now())
	log.Info().Msgf("Token is valid for %v (%v)", until, wrapper.Data.ExpireTime)
	metrics.VaultTokenLifetime.Set(float64(wrapper.Data.ExpireTime.Unix()))
	return &wrapper.Data, nil
}

type Credentials struct {
	AccessKey     string `json:"access_key"`
	SecretKey     string `json:"secret_key"`
	SecurityToken string `json:"security_token"`
}

func (c *Credentials) isValid() bool {
	return len(c.SecretKey) > 0 && len(c.AccessKey) > 0
}

type VaultCredentialResponse struct {
	Data          Credentials `json:"data,omitempty"`
	Auth          AuthReply   `json:"auth,omitempty"`
	Renewable     bool        `json:"renewable"`
	LeaseDuration int         `json:"lease_duration"`
	Warnings      []string    `json:"warnings"`
	RequestId     string      `json:"request_id"`
	LeaseId       string      `json:"lease_id"`
}

func (m *VaultCredentialProvider) readAwsCredentials() (*VaultCredentialResponse, error) {
	log.Info().Msgf("Generating dynamic AWS credentials for role %s", m.conf.RoleName)

	url := fmt.Sprintf("%s/v1/aws/creds/%s", m.conf.VaultAddr, m.conf.RoleName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}

	req.Header.Set(vaultTokenKey, m.vaultToken)
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error from vault: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("couldn't read response: %v", err)
	}

	return m.parseAwsCredentialsReply(body)
}

func (m *VaultCredentialProvider) parseAwsCredentialsReply(body []byte) (*VaultCredentialResponse, error) {
	var res VaultCredentialResponse
	err := json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal json: %v", err)
	}

	log.Info().Msgf("Received credentials with id %s, valid for %ds", res.Data.AccessKey, res.LeaseDuration)
	return &res, nil
}

func ConvertCredentials(dynamicCredentials Credentials) credentials.Value {
	return credentials.Value{
		AccessKeyID:     dynamicCredentials.AccessKey,
		SecretAccessKey: dynamicCredentials.SecretKey,
		SessionToken:    dynamicCredentials.SecurityToken,
		ProviderName:    "vault",
	}
}
