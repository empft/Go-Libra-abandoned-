package account

import (
	"context"
	"errors"

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
	Store(ctx context.Context, account *User) (int, error)
	FetchById(ctx context.Context, id int) (*User, error)
	FetchByUsername(ctx context.Context, name string) (*User, error)
	FetchByEmail(ctx context.Context, email string) ([]User, error)
	Update(ctx context.Context, account *User) error
	HasUsername(ctx context.Context, name string) (bool, error)
	HasInvitationEmail(ctx context.Context, email string) (bool, error)
}

type InvitationEmailRepository interface {
	Store(ctx context.Context, invitation InvitationEmail) error
	Fetch(ctx context.Context, email string) (string, error)
	Exist(ctx context.Context, email string) (bool, error)
}

func (user *User) ComparePassword(password string) (bool, error) {
	err := random.CompareHashAndPassword(user.PasswordHash, password)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (user *User) UpdatePassword(password string) error {
	hash, err := random.GenerateHash(password)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return nil
}

// Email must be verified after changing
func (user *User) UpdateEmail(email string) {
	user.UnverifiedEmail = email
}

// Email must be updated before verified
func (user *User) VerifyEmail(email string) error {
	if user.UnverifiedEmail != email {
		return errors.New("email has changed")
	}

	user.UnverifiedEmail = ""
	user.Email = email
	return nil
}