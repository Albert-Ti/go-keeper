package email

import (
	"github.com/wneessen/go-mail"
)

type Sender struct {
	client *mail.Client
	from   string
}

func NewSender(host string, port int, username, pass, from string) (*Sender, error) {
	client, err := mail.NewClient(host,
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username),
		mail.WithPassword(pass),
		mail.WithSSL(),
	)
	if err != nil {
		return nil, err
	}
	return &Sender{client: client, from: from}, nil
}

func (s *Sender) SendConfirmationCode(to, code string) error {
	msg := mail.NewMsg()
	if err := msg.From(s.from); err != nil {
		return err
	}
	if err := msg.To(to); err != nil {
		return err
	}
	msg.Subject("Confirm your email")
	msg.SetBodyString(mail.TypeTextPlain, "Your confirmation code: "+code)

	return s.client.DialAndSend(msg)
}
