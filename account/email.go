package account

import "golang.org/x/text/language"

type ExternalComm interface {
	VerifyInvitationEmail(loc language.Tag, to string, otp string) error
	VerifyRecoveryEmail(loc language.Tag, to string, otp string) error
	RemindUsername(loc language.Tag, to string, names ...string) error
	ResetPassword(loc language.Tag, to string, username, link string) error
}