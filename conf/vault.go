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
	VaultAddr string `yaml:"vault_addr,omitempty" env:"VAULT_ADDR" validate:"required_unless=AuthStrategy ''"`

	AuthStrategy VaultAuthStrategy `yaml:"vault_auth_strategy" env:"VAULT_AUTH_STRATEGY" validate:"omitempty,oneof=token approle kubernetes"`

	AwsRoleName  string `yaml:"vault_aws_role_name,omitempty" env:"VAULT_AWS_ROLE_NAME"`
	AwsMountPath string `yaml:"vault_aws_mount_path,omitempty" env:"VAULT_AWS_MOUNT"`

	AppRoleId       string `yaml:"vault_app_role_id,omitempty" env:"VAULT_APPROLE_ROLE_ID"`
	AppRoleSecretId string `yaml:"vault_app_role_secret,omitempty" env:"VAULT_APPROLE_SECRET_ID"`

	VaultToken string `yaml:"vault_token,omitempty" env:"VAULT_TOKEN"`
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
