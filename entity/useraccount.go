package entity

import (
	"errors"

	"github.com/elithrar/simple-scrypt"
)

type UserAccount struct {
	Id              *int
	InvitationEmail *string
	Username        string
	DisplayName     string
	passwordHash    []byte
	email           string
	emailVerified   bool
}

type InvitationEmail struct {
	Email string
	Code string
}

type RecoveryEmailVerification struct {
	Email string
	Token string // for verificating email
}

type UserAccountRepository interface {
	Store(account UserAccount) error
	Fetch(id int) (UserAccount, error)
	HasUsername(name string) (bool, error)
	HasInvitationEmail(email string) (bool, error)
}

type InvitationEmailRepository interface {
	Store(invitation InvitationEmail) error
	Fetch(email string) (string, error)
	Exist(email string) (bool, error)
}

type RecoveryEmailVerificationRepository interface {
	Store(verification RecoveryEmailVerification) error
	Fetch(email string) (RecoveryEmailVerification, error)
}

func (user *UserAccount) Email() (string, error) {
	if !user.emailVerified {
		return "", errors.New("Email Not Verified")
	}

	return user.email, nil
}

func (user *UserAccount) ComparePassword(password string) (bool, error) {
	err := scrypt.CompareHashAndPassword(user.passwordHash, []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (UserAccount) GenerateHash(password string) ([]byte, error) {
	return scrypt.GenerateFromPassword([]byte(password), scrypt.DefaultParams)
}




