package entity

import "github.com/stevealexrs/Go-Libra/random"

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
	Store(verification *RecoveryEmailVerification) error
	Fetch(accountId int, email string) (*RecoveryEmailVerification, error)
	Delete(accountId int, email string) error
	Exist(accountId int, email string) (bool, error)
}

// Reset account password
type AccountRecovery struct {
	UserId int
	Token string
}

func (r *AccountRecovery) Verify(token string) bool {
	return r.Token == token
}

type AccountRecoveryRepository interface {
	Store(recovery *AccountRecovery) error
	Fetch(accountId int) (*AccountRecovery, error)
	Delete(accountId int) error
	Exist(accountId int) (bool, error)
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