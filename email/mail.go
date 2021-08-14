package email

//go:generate gotext -srclang=en update -out=catalog/catalog.go -lang=en,zh-Hans,ms

import (
	"bytes"
	"html/template"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	_ "golang.org/x/text/message/catalog"
)


const directory = "./email/template/*"
var t = template.Must(template.ParseGlob(directory))

// Email Header and Footer
type EmailHF struct {
	Header string
	Footer string
}

type otpMessage struct {
	EmailHF
	Message string
	Otp 	string
}

func (s *Client) VerifyInvitationEmail(loc language.Tag, to string, otp string) error {
	p := message.NewPrinter(loc)
	var defHF = EmailHF{
		Header: p.Sprintf("An Accessible Payment System"),
		Footer: p.Sprintf("Never log into your account through any links provided in an email."),
	}

	b := new(bytes.Buffer)
	err := t.ExecuteTemplate(b, "otp.html", otpMessage{
		Message: "Here is the code for verifying your email:",
		Otp: otp, 
		EmailHF: defHF,
	})
	if err != nil {
		return err
	}
	
	
	header := "Subject: " + p.Sprintf("Verification Code for Invitation Email") + "\n" +
			  "MIME-version: 1.0\n" +
			  "Content-Type: text/html; charset=\"UTF-8\"\n"

	msg := []byte(header + b.String())

	return s.Send([]string{to}, msg)
}

func (s *Client) VerifyRecoveryEmail(loc language.Tag, to string, otp string) error {
	p := message.NewPrinter(loc)
	var defHF = EmailHF{
		Header: p.Sprintf("An Accessible Payment System"),
		Footer: p.Sprintf("Never log into your account through any links provided in an email."),
	}

	b := new(bytes.Buffer)
	err := t.ExecuteTemplate(b, "otp.html", otpMessage{
		Message: "Here is the code for verifying your email:",
		Otp: otp, 
		EmailHF: defHF,
	})
	if err != nil {
		return err
	}

	header := "Subject: " + p.Sprintf("Verification Code for Recovery Email") + "\n" +
			  "MIME-version: 1.0\n" +
			  "Content-Type: text/html; charset=\"UTF-8\"\n"

	msg := []byte(header + b.String())

	return s.Send([]string{to}, msg)
}

func (s *Client) RemindUsername(loc language.Tag, to string, names ...string) error {
	p := message.NewPrinter(loc)
	var defHF = EmailHF{
		Header: p.Sprintf("An Accessible Payment System"),
		Footer: p.Sprintf("Never log into your account through any links provided in an email."),
	}

	type usernameMessage struct {
		Usernames []string
		Message   string
		EmailHF
	}

	b := new(bytes.Buffer)
	err := t.ExecuteTemplate(b, "otp.html", usernameMessage{
		Usernames: names,
		Message: p.Sprintf("Here is the list of usernames associated with your email:"),
		EmailHF: defHF,
	})
	if err != nil {
		return err
	}

	header := "Subject: " + p.Sprintf("Username Reminder") + "\n" +
			  "MIME-version: 1.0\n" +
			  "Content-Type: text/html; charset=\"UTF-8\"\n"

	msg := []byte(header + b.String())

	return s.Send([]string{to}, msg)
}

func (s *Client) ResetPassword(loc language.Tag, to string, username, link string) error {
	p := message.NewPrinter(loc)
	var defHF = EmailHF{
		Header: p.Sprintf("An Accessible Payment System"),
		Footer: p.Sprintf("Never log into your account through any links provided in an email."),
	}

	type resetMessage struct {
		Message string
		Link  	string
		EmailHF
	}

	b := new(bytes.Buffer)
	err := t.ExecuteTemplate(b, "otp.html", resetMessage{
		Message: p.Sprintf("Hi %s, reset your password using the link below.", username),
		Link: link,
		EmailHF: defHF,
	})
	if err != nil {
		return err
	}

	header := "Subject: " + p.Sprintf("Password Reset") + "\n" +
			  "MIME-version: 1.0\n" +
			  "Content-Type: text/html; charset=\"UTF-8\"\n"

	msg := []byte(header + b.String())

	return s.Send([]string{to}, msg)
}
