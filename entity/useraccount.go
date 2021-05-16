package entity

import (
	"errors"

	"github.com/elithrar/simple-scrypt"
	"github.com/stevealexrs/Go-Libra/random"
)

type UserAccount struct {
	Id              *int
	InvitationEmail string
	Username        string
	DisplayName     string
	passwordHash    []byte
	email           string
	emailVerified   bool
}

func NewUserAccount(invitationEmail, username, displayName, password, email string) (*UserAccount, error) {
	hash, err := generateHash(password)
	if err != nil {
		return nil, err
	}
	
	acc := &UserAccount{
		Id: nil,
		InvitationEmail: invitationEmail,
		Username: username,
		DisplayName: displayName,
		passwordHash: hash,
		email: email,
		emailVerified: false,
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
		Email: email,
		Token: token,
	}
	return obj, nil
}

type UserAccountRepository interface {
	Store(account UserAccount) error
	FetchById(id int) (UserAccount, error)
	FetchByUsername(name string) (UserAccount, error)
	Update(account UserAccount) error
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
	Fetch(token string) (RecoveryEmailVerification, error)
}

func (user *UserAccount) Email() (string, error) {
	if !user.emailVerified {
		return "", errors.New("no verified email")
	}

	return user.email, nil
}

func (user *UserAccount) UnverifiedEmail() (string, error) {
	 if user.emailVerified {
		 return "", errors.New("no unverified email")
	 }
	 return user.email, nil
}

func (user *UserAccount) IsEmailVerified() bool {
	return user.emailVerified
}

func (user *UserAccount) ComparePassword(password string) (bool, error) {
	err := scrypt.CompareHashAndPassword(user.passwordHash, []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (user *UserAccount) UpdatePassword(password string) error {
	hash, err := generateHash(password)
	if err != nil {
		return err
	}
	user.passwordHash = hash
	return nil
}

// Email is not retrievable until verified
func (user *UserAccount) UpdateEmail(email string) {
	user.email = email
	user.emailVerified = false
}

func (user *UserAccount) VerifyEmail(email string) error {
	if user.email != email {
		return errors.New("email has changed")
	}

	user.emailVerified = true
	return nil
}

func generateHash(password string) ([]byte, error) {
	return scrypt.GenerateFromPassword([]byte(password), scrypt.DefaultParams)
}




