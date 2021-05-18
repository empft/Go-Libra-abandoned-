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
	Email           string
	UnverifiedEmail string
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
		Email: "",
		UnverifiedEmail: email,
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

type AccountRecovery struct {
	UserId int
	Token string
}

func NewAccountRecovery(userId int) (*AccountRecovery, error) {
	token, err := random.Token20Byte()
	if err != nil {
		return nil, err
	}

	obj := &AccountRecovery{
		UserId: userId,
		Token: token,
	}
	return obj, nil
}

type UserAccountRepository interface {
	Store(account UserAccount) (int, error)
	FetchById(id int) (UserAccount, error)
	FetchByUsername(name string) (UserAccount, error)
	FetchByEmail(email string) ([]UserAccount, error)
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
	Fetch(userId int, email string) (RecoveryEmailVerification, error)
	Exist(userId int, email string) (bool, error)
}

type AccountRecoveryRepository interface {
	Store(recovery AccountRecovery) error
	Fetch(userId int) (AccountRecovery, error)
	Exist(userId int) (bool, error)
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

// Email must be verified after changing
func (user *UserAccount) UpdateEmail(email string) {
	user.UnverifiedEmail = email
}

// Email must be updated before verified
func (user *UserAccount) VerifyEmail(email string) error {
	if user.UnverifiedEmail != email {
		return errors.New("email has changed")
	}

	user.UnverifiedEmail = ""
	user.Email = email
	return nil
}

func generateHash(password string) ([]byte, error) {
	return scrypt.GenerateFromPassword([]byte(password), scrypt.DefaultParams)
}