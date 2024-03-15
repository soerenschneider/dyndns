package conf

import (
	"errors"
	"os"
	"strings"
)

type EmailConfig struct {
	From             string   `yaml:"from" env:"EMAIL_FROM" validate:"required_with=SmtpHost,omitempty,required_without=FromFile"`
	FromFile         string   `yaml:"from_file" env:"EMAIL_FROM_FILE" validate:"required_with=SmtpHost,omitempty,required_without=From"`
	To               []string `yaml:"to"   env:"EMAIL_TO" envSeparator:";" validate:"required_with=SmtpHost,omitempty,required_without=ToFile"`
	ToFile           string   `yaml:"to_file"   env:"EMAIL_TO" validate:"required_with=SmtpHost,omitempty,required_without=To"`
	SmtpHost         string   `yaml:"host" env:"EMAIL_HOST" validate:"omitempty,required"`
	SmtpPort         int      `yaml:"port" env:"EMAIL_PORT" validate:"omitempty,required"`
	SmtpUsername     string   `yaml:"user" env:"EMAIL_USER" validate:"required_with=SmtpHost,omitempty,required_without=SmtpUsernameFile"`
	SmtpUsernameFile string   `yaml:"user_file" env:"EMAIL_USER_FILE" validate:"required_with=SmtpHost,omitempty,required_without=smtpUsername"`
	SmtpPassword     string   `yaml:"password" env:"EMAIL_PASSWORD" validate:"required_with=SmtpHost,omitempty,required_without=SmtpPasswordFile"`
	SmtpPasswordFile string   `yaml:"password_file" env:"EMAIL_PASSWORD_FILE" validate:"required_with=SmtpHost,omitempty,required_without=SmtpPassword"`
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
	if len(conf.From) == 0 {
		return errors.New("'From' not defined")
	}
	if len(conf.To) == 0 {
		return errors.New("'To' not defined")
	}
	if len(conf.SmtpHost) == 0 {
		return errors.New("'SmtpHost' not defined")
	}
	if len(conf.SmtpUsername) == 0 {
		return errors.New("'smtpUsername' not defined")
	}
	if len(conf.SmtpPassword) == 0 {
		return errors.New("'SmtpPassword' not defined")
	}
	return nil
}
