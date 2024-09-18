package conf

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
)

type EmailConfig struct {
	From             string   `yaml:"from" env:"EMAIL_FROM" validate:"omitempty,email"`
	FromFile         string   `yaml:"from_file" env:"EMAIL_FROM_FILE" validate:"omitempty,filepath"`
	To               []string `yaml:"to" env:"EMAIL_TO" envSeparator:";" validate:"omitempty,dive,email"`
	ToFile           string   `yaml:"to_file" env:"EMAIL_TO" validate:"omitempty,filepath"`
	SmtpHost         string   `yaml:"host" env:"EMAIL_HOST" validate:"omitempty,hostname"`
	SmtpPort         int      `yaml:"port" env:"EMAIL_PORT" validate:"omitempty,gte=25,lte=65535"`
	SmtpUsername     string   `yaml:"user" env:"EMAIL_USER"`
	SmtpUsernameFile string   `yaml:"user_file" env:"EMAIL_USER_FILE" validate:"omitempty,filepath"`
	SmtpPassword     string   `yaml:"password" env:"EMAIL_PASSWORD"`
	SmtpPasswordFile string   `yaml:"password_file" env:"EMAIL_PASSWORD_FILE" validate:"omitempty,filepath"`
}

func (e *EmailConfig) String() string {
	var sb strings.Builder

	sb.WriteString("EmailConfig {")
	appendIfNotEmpty(&sb, "From", e.From)
	appendIfNotEmpty(&sb, "FromFile", e.FromFile)
	if len(e.To) > 0 {
		sb.WriteString(fmt.Sprintf(" To: %v,", e.To))
	}
	appendIfNotEmpty(&sb, "ToFile", e.ToFile)
	appendIfNotEmpty(&sb, "SmtpHost", e.SmtpHost)
	if e.SmtpPort != 0 {
		sb.WriteString(fmt.Sprintf(" SmtpPort: %d,", e.SmtpPort))
	}
	appendIfNotEmpty(&sb, "SmtpUsername", e.SmtpUsername)
	appendIfNotEmpty(&sb, "SmtpUsernameFile", e.SmtpUsernameFile)
	// Note: We deliberately exclude SmtpPassword from the output
	appendIfNotEmpty(&sb, "SmtpPasswordFile", e.SmtpPasswordFile)
	sb.WriteString(" }")

	return sb.String()
}

func appendIfNotEmpty(sb *strings.Builder, fieldName, value string) {
	if value != "" {
		sb.WriteString(fmt.Sprintf(" %s: %s,", fieldName, value))
	}
}

func (conf *EmailConfig) IsConfigured() bool {
	return (len(conf.From) > 0 || len(conf.FromFile) > 0) && (len(conf.To) > 0 || len(conf.ToFile) > 0)
}

func (conf *EmailConfig) GetFrom() (string, error) {
	if len(conf.From) > 0 {
		return conf.From, nil
	}

	data, err := os.ReadFile(conf.FromFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (conf *EmailConfig) GetTo() ([]string, error) {
	if len(conf.To) > 0 {
		return conf.To, nil
	}

	data, err := os.ReadFile(conf.ToFile)
	if err != nil {
		return nil, err
	}
	sData := string(data)
	return strings.Split(sData, ","), nil
}

func (conf *EmailConfig) GetUsername() (string, error) {
	if len(conf.SmtpUsername) > 0 {
		return conf.SmtpUsername, nil
	}

	data, err := os.ReadFile(conf.SmtpUsernameFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (conf *EmailConfig) GetPassword() (string, error) {
	if len(conf.SmtpPassword) > 0 {
		return conf.SmtpPassword, nil
	}

	data, err := os.ReadFile(conf.SmtpPasswordFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (conf *EmailConfig) Validate() error {
	return ValidateConfig(conf)
}

func EmailConfigStructLevelValidation(sl validator.StructLevel) {
	config := sl.Current().Interface().(EmailConfig)

	if config.SmtpPort == 0 && len(config.SmtpHost)+len(config.SmtpUsername)+len(config.SmtpUsernameFile)+len(config.SmtpPasswordFile)+len(config.SmtpPassword)+len(config.From)+len(config.FromFile)+len(config.To)+len(config.ToFile) == 0 {
		return
	}

	if config.SmtpUsername == "" && config.SmtpUsernameFile == "" {
		sl.ReportError(config.SmtpUsername, "SmtpUsername", "SmtpUsername", "usernameOrFileRequired", "")
		sl.ReportError(config.SmtpUsernameFile, "SmtpUsernameFile", "SmtpUsernameFile", "usernameOrFileRequired", "")
	}
	if config.SmtpPassword == "" && config.SmtpPasswordFile == "" {
		sl.ReportError(config.SmtpPassword, "SmtpPassword", "SmtpPassword", "passwordOrFileRequired", "")
		sl.ReportError(config.SmtpPasswordFile, "SmtpPasswordFile", "SmtpPasswordFile", "passwordOrFileRequired", "")
	}

	if config.From == "" && config.FromFile == "" {
		sl.ReportError(config.From, "From", "From", "requiredWithoutFromFile", "")
		sl.ReportError(config.FromFile, "FromFile", "FromFile", "requiredWithoutFrom", "")
	}

	if len(config.To) == 0 && config.ToFile == "" {
		sl.ReportError(config.To, "To", "To", "requiredWithoutToFile", "")
		sl.ReportError(config.ToFile, "ToFile", "ToFile", "requiredWithoutTo", "")
	}
}
