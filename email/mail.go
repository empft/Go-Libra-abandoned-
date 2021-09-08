package email

//go:generate gotext -srclang=en update -out=catalog/catalog.go -lang=en,zh-Hans,ms

import (
	"bytes"
	"context"
	"flag"
	"html/template"
	"os"
	"strings"

	"github.com/stevealexrs/Go-Libra/namespace/reqscope"
	"golang.org/x/text/message"
	_ "golang.org/x/text/message/catalog"
)

var t = &template.Template{}

func init() {
	var directory = "./email/template/*"
	// include both ways to check whether it is testing
	if flag.Lookup("test.v") == nil || strings.HasSuffix(os.Args[0], ".test") {
		directory = "./template/*"
	}
	t = template.Must(template.ParseGlob(directory))
}

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

func (s *Client) VerifyInvitationEmail(ctx context.Context, to, otp string) error {
	p := message.NewPrinter(reqscope.Language(ctx))
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
			  "Content-Type: text/html; charset=\"UTF-8\"\n\n"

	msg := []byte(header + b.String())

	return s.Send([]string{to}, msg)
}

func (s *Client) VerifyRecoveryEmail(ctx context.Context, to, otp string) error {
	p := message.NewPrinter(reqscope.Language(ctx))
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
			  "Content-Type: text/html; charset=\"UTF-8\"\n\n"

	msg := []byte(header + b.String())

	return s.Send([]string{to}, msg)
}

func (s *Client) RemindUsername(ctx context.Context, to string, names ...string) error {
	p := message.NewPrinter(reqscope.Language(ctx))
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
	err := t.ExecuteTemplate(b, "remindname.html", usernameMessage{
		Usernames: names,
		Message: p.Sprintf("Here is a list of usernames associated with your email:"),
		EmailHF: defHF,
	})
	if err != nil {
		return err
	}

	header := "Subject: " + p.Sprintf("Username Reminder") + "\n" +
			  "MIME-version: 1.0\n" +
			  "Content-Type: text/html; charset=\"UTF-8\"\n\n"

	msg := []byte(header + b.String())

	return s.Send([]string{to}, msg)
}

func (s *Client) ResetPassword(ctx context.Context, to, username, token string) error {
	p := message.NewPrinter(reqscope.Language(ctx))
	var defHF = EmailHF{
		Header: p.Sprintf("An Accessible Payment System"),
		Footer: p.Sprintf("Never log into your account through any links provided in an email."),
	}

	type resetMessage struct {
		Message string
		Token   string
		EmailHF
	}

	b := new(bytes.Buffer)
	err := t.ExecuteTemplate(b, "resetpassword.html", resetMessage{
		Message: p.Sprintf("Hi %s, reset your password using the token below.", username),
		Token: token,
		EmailHF: defHF,
	})
	if err != nil {
		return err
	}

	header := "Subject: " + p.Sprintf("Password Reset") + "\n" +
			  "MIME-version: 1.0\n" +
			  "Content-Type: text/html; charset=\"UTF-8\"\n\n"

	msg := []byte(header + b.String())

	return s.Send([]string{to}, msg)
}
