package account

import (
	"context"

	"github.com/stevealexrs/Go-Libra/random"
)

type User struct {
	Base
	InvitationEmail string
	DisplayName     string
}

func NewUserAccountWithPassword(invitationEmail, username, displayName, password, email string) (*User, error) {
	hash, err := random.GenerateHash(password)
	if err != nil {
		return nil, err
	}
	
	acc := &User{
		Base: Base{
			Id:              nil,
			Username:        username,
			PasswordHash:    hash,
			Email:           "",
			UnverifiedEmail: email,
		},
		InvitationEmail: invitationEmail,
		DisplayName: displayName,
	}
	return acc, nil
}

type InvitationEmail struct {
	Email string
	Code string
}

func NewInvitationEmail(email string) (*InvitationEmail, error) {
	otp, err := random.OTP()
	if err != nil {
		return nil, err
	}
	invitation := &InvitationEmail{
		Email: email,
		Code: otp,
	}
	return invitation, err
}


type UserAccountRepository interface {
	// returns id
	Store(ctx context.Context, account *User) (int, error)
	FetchById(ctx context.Context, id int) (*User, error)
	FetchByUsername(ctx context.Context, name string) (*User, error)
	FetchByEmail(ctx context.Context, email string) ([]User, error)
	Update(ctx context.Context, account *User) error
	HasUsername(ctx context.Context, name string) (bool, error)
	HasInvitationEmail(ctx context.Context, email string) (bool, error)
}

type InvitationEmailVerificationRepository interface {
	Store(ctx context.Context, invitation InvitationEmail) error
	Fetch(ctx context.Context, email string) (string, error)
	Exist(ctx context.Context, email string) (bool, error)
}