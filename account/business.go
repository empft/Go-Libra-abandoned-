package account

import (
	"context"

	"github.com/stevealexrs/Go-Libra/random"
)

type Business struct {
	Base
	BusinessName
	BusinessIdentity
	ChildOf 		 *int
}

type BusinessName struct {
	DisplayName string
	Verified 	bool
}

type BusinessIdentity struct {	
	Name               string
	RegistrationNumber string
	Address            string
	// URL of documents
	Documents 		   []string
	Verified 		   bool
}

type BusinessAccountRepository interface {
	// returns id
	Store(ctx context.Context, account *Business, documents [][]byte) (int, error)
	FetchById(ctx context.Context, id int) (*Business, error)
	FetchByUsername(ctx context.Context, name string) (*Business, error)
	FetchByEmail(ctx context.Context, email string) ([]Business, error)
	Update(ctx context.Context, account *Business, documents [][]byte) error
	HasUsername(ctx context.Context, name string) (bool, error)
}

func NewBusinessAccountWithPassword(username, displayName, password, email string, identity *BusinessIdentity) (*Business, error) {
	hash, err := random.GenerateHash(password)
	if err != nil {
		return nil, err
	}

	businessIdentity := BusinessIdentity{}
	if identity != nil {
		businessIdentity = *identity
	}

	acc := &Business{
		Base: Base{
			Id:              nil,
			Username:        username,
			PasswordHash:    hash,
			Email:           "",
			UnverifiedEmail: email,
		},
		BusinessName: BusinessName{
			DisplayName:     displayName,
			Verified: false,
		},
		BusinessIdentity: businessIdentity,
	}
	return acc, nil
}

func NewBusinessIdentity(name, registrationNumber, address string) (*BusinessIdentity, error) {
	identity := &BusinessIdentity{
		Name:               name,
		RegistrationNumber: registrationNumber,
		Address:            address,
		Verified:           false,
	}
	return identity, nil
}