package account

import (
	"context"

	"github.com/stevealexrs/Go-Libra/random"
)

// Verify that the email belongs to the user
type RecoveryEmailVerification struct {
	UserId int
	Email  string
	Token  string // for verificating email
}

func NewRecoveryEmailVerification(userId int, email string) (*RecoveryEmailVerification, error) {
	token, err := random.Token20Byte()
	if err != nil {
		return nil, err
	}

	obj := &RecoveryEmailVerification{
		UserId: userId,
		Email:  email,
		Token:  token,
	}
	return obj, nil
}

func (r *RecoveryEmailVerification) Verify(token string) bool {
	return r.Token == token
}

type RecoveryEmailVerificationRepository interface {
	Store(ctx context.Context, verification *RecoveryEmailVerification) error
	Fetch(ctx context.Context, accountId int, email string) (string, error)
	Delete(ctx context.Context, accountId int, email string) error
	Exist(ctx context.Context, accountId int, email string) (bool, error)
}

// Reset account password
type Recovery struct {
	UserId int
	Token string
}

func (r *Recovery) Verify(token string) bool {
	return r.Token == token
}

type RecoveryRepository interface {
	Store(ctx context.Context, recovery *Recovery) error
	Fetch(ctx context.Context, accountId int) (string, error)
	Delete(ctx context.Context, accountId int) error
	Exist(ctx context.Context, accountId int) (bool, error)
}

func NewAccountRecovery(userId int) (*Recovery, error) {
	token, err := random.Token20Byte()
	if err != nil {
		return nil, err
	}

	obj := &Recovery{
		UserId: userId,
		Token: token,
	}
	return obj, nil
}