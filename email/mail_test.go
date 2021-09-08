package email_test

import (
	"context"
	"net/smtp"
	"testing"

	"github.com/stevealexrs/Go-Libra/email"
	_ "golang.org/x/text/message/catalog"
)

type papercutService struct {
	from string
}

func (s *papercutService) Send(to []string, msg []byte) error {
	return smtp.SendMail(
		"127.0.0.1:25",
		nil,
		s.from,
		to,
		msg,
	)
}

var testSMTP = email.Client{
	Service: &papercutService{from: "testing@local.com"},
}

func TestClient_VerifyInvitationEmail(t *testing.T) {
	if err := testSMTP.VerifyInvitationEmail(context.Background(), "yourinvitation@random.com", "123456"); err != nil {
		t.Error(err)
	}
}

func TestClient_VerifyRecoveryEmail(t *testing.T) {
	if err := testSMTP.VerifyRecoveryEmail(context.Background(), "yourinvitation@random.com", "123456"); err != nil {
		t.Error(err)
	}
}

func TestClient_RemindUsername(t *testing.T) {
	if err := testSMTP.RemindUsername(context.Background(), "yourinvitation@random.com", "jane doe", "leka", "grrr", "your name"); err != nil {
		t.Error(err)
	}
}

func TestClient_ResetPassword(t *testing.T) {
	if err := testSMTP.ResetPassword(context.Background(), "yourinvitation@random.com", "the_resetter", "EXTREMELY SECRET CODE"); err != nil {
		t.Error(err)
	}
}
