package email

import "net/smtp"

type SMTPService struct {
	server string
	from  string
	auth   smtp.Auth
}

func NewSMTPService(server, email, password, host string) *SMTPService {
	auth := smtp.PlainAuth("", email, password, host)

	return &SMTPService{
		server: server,
		from:  email,
		auth:   auth,
	}
}

func (s *SMTPService) Send(to []string, msg []byte) error {
	return smtp.SendMail(
		s.server,
		nil,
		s.from,
		to,
		msg,
	)
}