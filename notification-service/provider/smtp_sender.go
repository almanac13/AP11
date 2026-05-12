package provider

import "log"

type SMTPSender struct{}

func NewSMTPSender() *SMTPSender {
	return &SMTPSender{}
}

func (s *SMTPSender) Send(to string, subject string, body string) error {
	log.Printf("[Provider:REAL-placeholder] would send SMTP email to=%s subject=%q body=%q", to, subject, body)
	return nil
}
