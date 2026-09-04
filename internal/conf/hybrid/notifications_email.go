package hybrid

import (
	"os"
	"strings"
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

func (conf *EmailConfig) BuildEmailNotification() bool {
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
