package account

import (
	"errors"
	"fmt"

	"github.com/stevealexrs/Go-Libra/random"
)

// Basic account structure
// Business account and user account have the same id space
type Base struct {
	Id              *int
	Username        string
	PasswordHash    []byte
	Email           string
	UnverifiedEmail string
	Deleted			bool
}

func (base *Base) ComparePassword(password string) (bool, error) {
	err := random.CompareHashAndPassword(base.PasswordHash, password)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (base *Base) UpdatePassword(password string) error {
	hash, err := random.GenerateHash(password)
	if err != nil {
		return fmt.Errorf("fail to generate hash: %v", err)
	}
	base.PasswordHash = hash
	return nil
}

// Email must be verified after changing
func (base *Base) UpdateEmail(email string) {
	base.UnverifiedEmail = email
}

// Email must be updated before verified
func (base *Base) VerifyEmail(email string) error {
	if base.UnverifiedEmail != email {
		return errors.New("email has changed")
	}

	base.UnverifiedEmail = ""
	base.Email = email
	return nil
}