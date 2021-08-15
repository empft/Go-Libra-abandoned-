package account

import (
	"errors"

	"github.com/stevealexrs/Go-Libra/random"
)

// Basic account structure
// Business account and user account have the same id space
type base struct {
	Id              *int
	Username        string
	PasswordHash    []byte
	Email           string
	UnverifiedEmail string
}

func (base *base) ComparePassword(password string) (bool, error) {
	err := random.CompareHashAndPassword(base.PasswordHash, password)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (base *base) UpdatePassword(password string) error {
	hash, err := random.GenerateHash(password)
	if err != nil {
		return err
	}
	base.PasswordHash = hash
	return nil
}

// Email must be verified after changing
func (base *base) UpdateEmail(email string) {
	base.UnverifiedEmail = email
}

// Email must be updated before verified
func (base *base) VerifyEmail(email string) error {
	if base.UnverifiedEmail != email {
		return errors.New("email has changed")
	}

	base.UnverifiedEmail = ""
	base.Email = email
	return nil
}