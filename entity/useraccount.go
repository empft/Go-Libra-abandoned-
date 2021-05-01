package entity

type UserAccount struct {
	Id				  int
	VerificationEmail string
	Username          string
	DisplayName       string
	PasswordHash      []byte
	Email             string
}

type UserAccountRepository interface {
	Store(account UserAccount) error
	Fetch(id int) (UserAccount, error)
}
