package account

import (
	"context"
)

type ExternalComm interface {
	VerifyInvitationEmail(ctx context.Context, to, otp string) error
	VerifyRecoveryEmail(ctx context.Context, to, otp string) error
	RemindUsername(ctx context.Context, to string, names ...string) error
	ResetPassword(ctx context.Context, to, username, token string) error
}