package entity

type UserAccount struct {
	Id				  int
	InvitationalEmail *string
	Username          string
	DisplayName       string
	PasswordHash      []byte
	Email             *string
}

type UserAccountRepository interface {
	Store(account UserAccount) error
	Fetch(id int) (UserAccount, error)
}
