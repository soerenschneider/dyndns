package conf

import (
	"errors"
	"fmt"
)

type EmailConfig struct {
	From         string   `json:"from" env:"DYNDNS_EMAIL_FROM"`
	To           []string `json:"to"   env:"DYNDNS_EMAIL_TO" envSeparator:";"`
	SmtpHost     string   `json:"host" env:"DYNDNS_EMAIL_HOST"`
	SmtpPort     int      `json:"port" env:"DYNDNS_EMAIL_PORT"`
	SmtpUsername string   `json:"user" env:"DYNDNS_EMAIL_USER"`
	SmtpPassword string   `json:"password" env:"DYNDNS_EMAIL_PASSWORD"`
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
		return errors.New("'SmtpUsername' not defined")
	}
	if len(conf.SmtpPassword) == 0 {
		return errors.New("'SmtpPassword' not defined")
	}
	return nil
}

func (conf *EmailConfig) String() string {
	return fmt.Sprintf("from=%s, to=%v, smtpHost=%s, smtpPort=%d, smtpUsername=%s", conf.From, conf.To, conf.SmtpHost, conf.SmtpPort, conf.SmtpUsername)
}
