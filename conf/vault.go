package conf

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

type VaultConfig struct {
	RoleName      string `json:"vault_role_name,omitempty"`
	VaultAddr     string `json:"vault_addr,omitempty"`
	AppRoleId     string `json:"vault_app_role_id,omitempty"`
	AppRoleSecret string `json:"vault_app_role_secret,omitempty"`
	VaultToken    string `json:"vault_token,omitempty"`
}

func GetDefaultVaultConfig() VaultConfig {
	return VaultConfig{
		RoleName:  "dyndns",
		VaultAddr: os.Getenv("VAULT_ADDR"),
	}
}

func (c *VaultConfig) Print() {
	log.Info().Msgf("RoleName=%s", c.RoleName)
	log.Info().Msgf("VaultAddr=%s", c.VaultAddr)
	if len(c.AppRoleId) > 0 {
		log.Info().Msgf("RoleName=%s", "***")
	}
	if len(c.AppRoleSecret) > 0 {
		log.Info().Msgf("RoleName=%s", "***")
	}
	if len(c.VaultToken) > 0 {
		log.Info().Msgf("VaultToken=%s", "***")
	}
}

func (c *VaultConfig) WithRoleName(roleName string) *VaultConfig {
	c.RoleName = roleName
	return c
}

func (c *VaultConfig) WithAppRole(appRoleId, appRoleSecret string) *VaultConfig {
	c.AppRoleId = appRoleId
	c.AppRoleSecret = appRoleSecret
	return c
}

func (c *VaultConfig) WithVaultToken(vaultToken string) *VaultConfig {
	c.VaultToken = vaultToken
	return c
}

func (c *VaultConfig) HasAppRoleLoginInfo() bool {
	return len(c.AppRoleSecret) > 0 && len(c.AppRoleId) > 0
}

func (c *VaultConfig) HasTokenInfo() bool {
	return len(c.VaultToken) > 0
}

func (c *VaultConfig) Verify() error {
	if !IsValidUrl(c.VaultAddr) {
		return fmt.Errorf("%s is not a valid url", c.VaultAddr)
	}

	hasVaultToken := c.HasTokenInfo()
	hasAppRole := c.HasAppRoleLoginInfo()
	if !hasVaultToken && !hasAppRole {
		return errors.New("no vault token and no app role information provided")
	}

	if len(c.RoleName) == 0 {
		return errors.New("no role name provided")
	}

	return nil
}
