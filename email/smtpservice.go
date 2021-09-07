package email

import "net/smtp"

type SMTPService struct {
	server string
	from  string
	auth   smtp.Auth
}

func NewSMTPService(email, password, host, server string) *SMTPService {
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
		s.auth,
		s.from,
		to,
		msg,
	)
}