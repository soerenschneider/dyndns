package conf

import (
	"os"
)

type VaultAuthStrategy string

var (
	VaultAuthStrategyToken      VaultAuthStrategy = "token"
	VaultAuthStrategyApprole    VaultAuthStrategy = "approle"
	VaultAuthStrategyKubernetes VaultAuthStrategy = "kubernetes"
)

type VaultConfig struct {
	VaultAddr string `json:"vault_addr,omitempty" env:"DYNDNS_VAULT_ADDR" validate:"required_unless=AuthStrategy ''"`

	AuthStrategy VaultAuthStrategy `json:"vault_auth_strategy" env:"DYNDNS_VAULT_AUTH_STRATEGY" validate:"omitempty,oneof=token approle kubernetes"`

	AwsRoleName  string `json:"vault_aws_role_name,omitempty" env:"DYNDNS_VAULT_AWS_ROLE_NAME"`
	AwsMountPath string `json:"vault_aws_mount_path,omitempty" env:"DYNDNS_VAULT_AWS_MOUNT"`

	AppRoleId       string `json:"vault_app_role_id,omitempty" env:"DYNDNS_VAULT_APPROLE_ROLE_ID"`
	AppRoleSecretId string `json:"vault_app_role_secret,omitempty" env:"DYNDNS_VAULT_APPROLE_SECRET_ID"`

	VaultToken string `json:"vault_token,omitempty" env:"DYNDNS_VAULT_TOKEN"`
}

func GetDefaultVaultConfig() *VaultConfig {
	return &VaultConfig{
		AwsRoleName:  "dyndns",
		AwsMountPath: "aws",
		VaultAddr:    os.Getenv("VAULT_ADDR"),
	}
}

func (c *VaultConfig) UseVaultCredentialsProvider() bool {
	return len(c.AuthStrategy) > 0
}
