package entity

type BusinessAccount struct {
	Id 						*int
	Username 				string
	BusinessName 			BusinessName
	PasswordHash    		[]byte
	RecoveryEmail			string
	UnverifiedRecoveryEmail string
}

type BusinessName struct {
	Name 	 string
	Verified bool
}

type BusinessIdentity struct {
	Id				   *int
	Name 			   string
	RegistrationNumber string
	Address 		   string
	// URL to the document
	Document 		   string
	Verified 		   bool
}

type BusinessAccountRepository interface {
	Store(account *BusinessAccount) (int, error)
	StoreAsChild(account *BusinessAccount, parentId int) (int, error)
	StoreWithIdentity(account *BusinessAccount, identity *BusinessIdentity) (int, error)
	FetchById(id int) (*BusinessAccount, error)
	FetchByUsername(name string) (*BusinessAccount, error)
	FetchByEmail(email string) ([]BusinessAccount, error)
	Update(account *BusinessAccount) error
	HasUsername(name string) (bool, error)
}

func NewBusinessAccountWithPassword(username, businessName, password, email string) (*BusinessAccount, error) {
	hash, err := generateHash(password)
	if err != nil {
		return nil, err
	}

	acc := &BusinessAccount{
		Id:       nil,
		Username: username,
		BusinessName: BusinessName{
			Name:     businessName,
			Verified: false,
		},
		PasswordHash:            hash,
		RecoveryEmail:           "",
		UnverifiedRecoveryEmail: email,
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