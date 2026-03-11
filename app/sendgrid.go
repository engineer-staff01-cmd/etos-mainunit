package app

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	from *mail.Email
}

func NewSendGridMailer() *SendGridMailer {
	return &SendGridMailer{
		from: mail.NewEmail("ecoRAMDAR", "eco-notice@ecoramdar.jp"),
	}
}

func (m *SendGridMailer) Send(apiKey, to, subject, body string) error {
	toEmail := mail.NewEmail(to, to)
	message := mail.NewSingleEmail(m.from, subject, toEmail, body, "")
	client := sendgrid.NewSendClient(apiKey)
	_, err := client.Send(message)
	if err != nil {
		return err
	}
	return nil
}
