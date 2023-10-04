package util

import (
	"fmt"

	"github.com/soerenschneider/dyndns/conf"
	"github.com/soerenschneider/dyndns/internal/common"
	"gopkg.in/gomail.v2"
)

type EmailNotification struct {
	From         string
	To           []string
	SmtpHost     string
	SmtpPort     int
	SmtpUsername string
	smtpPassword string
}

func NewEmailNotification(emailConf *conf.EmailConfig) (*EmailNotification, error) {
	return &EmailNotification{
		From:         emailConf.From,
		To:           emailConf.To,
		SmtpHost:     emailConf.SmtpHost,
		SmtpPort:     emailConf.SmtpPort,
		SmtpUsername: emailConf.SmtpUsername,
		smtpPassword: emailConf.SmtpPassword,
	}, nil
}

func (e *EmailNotification) NotifyUpdatedIpDetected(ip *common.DnsRecord) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", e.To...)
	subject := fmt.Sprintf("DynDNS new IP detected for host %s", ip.Host)
	m.SetHeader("Subject", subject)

	body := fmt.Sprintf("New IP detected for host %s: %s", ip.Host, ip.IpV4)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(e.SmtpHost, e.SmtpPort, e.SmtpUsername, e.smtpPassword)

	return d.DialAndSend(m)
}

func (e *EmailNotification) NotifyUpdatedIpApplied(ip *common.DnsRecord) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", e.To...)
	subject := fmt.Sprintf("DynDNS applied IP for host %s", ip.Host)
	m.SetHeader("Subject", subject)

	body := fmt.Sprintf("New IP applied for host %s: %s", ip.Host, ip.IpV4)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(e.SmtpHost, e.SmtpPort, e.SmtpUsername, e.smtpPassword)

	return d.DialAndSend(m)
}
