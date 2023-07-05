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
	VaultAddr string `json:"vault_addr,omitempty" validate:"required_unless=AuthStrategy ''"`

	AuthStrategy VaultAuthStrategy `json:"vault_auth_strategy" validate:"omitempty,oneof=token approle kubernetes"`

	AwsRoleName  string `json:"vault_aws_role_name,omitempty"`
	AwsMountPath string `json:"vault_aws_mount_path,omitempty"`

	AppRoleId       string `json:"vault_app_role_id,omitempty"`
	AppRoleSecretId string `json:"vault_app_role_secret,omitempty"`

	VaultToken string `json:"vault_token,omitempty"`
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
