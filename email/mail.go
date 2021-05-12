package email

import (
	"bytes"
	"html/template"
	"net/smtp"
)

// TODO: switch to an actual email service when it is live

// actual directory
// const directory = "./mail/template/*"

//debug directory
const directory = "./template/*"
var t = template.Must(template.ParseGlob(directory))

type Sender struct {
	server string
	email string
	auth smtp.Auth
}

func NewSender(server, email, password, host string) *Sender {
	auth := smtp.PlainAuth("", email, password, host)
	
	return &Sender{
		server: server,
		email: email,
		auth: auth,
	}
}

func (sender *Sender) sendMail(to []string, msg []byte) error {
	return smtp.SendMail(
				sender.server, 
				nil, 
				sender.email,
				to,
				msg,
			)
}

type OTP struct {
	Code string
}

func (sender *Sender) SendOTP(to []string, otp OTP) error {
	b := new(bytes.Buffer)
	err := t.ExecuteTemplate(b, "otp.html", otp)
	if err != nil {
		return err
	}

	header := 	"Subject: Your requested otp for Libra!\n" +
				"MIME-version: 1.0\n" +
				"Content-Type: text/html; charset=\"UTF-8\"\n" +
				"From: " + "Cloudy" + " <" + sender.email + ">\n\n"

	a := []byte(header + b.String())

	return sender.sendMail(to, a)
}

