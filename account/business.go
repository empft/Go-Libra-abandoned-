package account

import (
	"context"
	"github.com/stevealexrs/Go-Libra/random"
)

type Business struct {
	base
	BusinessName            BusinessName
}

type BusinessName struct {
	Name     string
	Verified bool
}

type BusinessIdentity struct {
	Id                 *int
	Name               string
	RegistrationNumber string
	Address            string
	// URL to the document
	Document string
	Verified bool
}

type BusinessAccountRepository interface {
	Store(ctx context.Context, account *Business) (int, error)
	StoreAsChild(ctx context.Context, account *Business, parentId int) (int, error)
	StoreWithIdentity(ctx context.Context, account *Business, identity *BusinessIdentity) (int, error)
	FetchById(ctx context.Context, id int) (*Business, error)
	FetchByUsername(ctx context.Context, name string) (*Business, error)
	FetchByEmail(ctx context.Context, email string) ([]Business, error)
	Update(ctx context.Context, account *Business) error
	HasUsername(ctx context.Context, name string) (bool, error)
}

func NewBusinessAccountWithPassword(username, businessName, password, email string) (*Business, error) {
	hash, err := random.GenerateHash(password)
	if err != nil {
		return nil, err
	}

	acc := &Business{
		base: base{
			Id:              nil,
			Username:        username,
			PasswordHash:    hash,
			Email:           "",
			UnverifiedEmail: email,
		},
		BusinessName: BusinessName{
			Name:     businessName,
			Verified: false,
		},
	}
	return acc, nil
}

func NewBusinessIdentity(name, registrationNumber, address, document string) (*BusinessIdentity, error) {
	identity := &BusinessIdentity{
		Id:                 nil,
		Name:               name,
		RegistrationNumber: registrationNumber,
		Address:            address,
		Document:           document,
		Verified:           false,
	}
	return identity, nil
}