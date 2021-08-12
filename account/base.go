package account

// Basic account structure
type base struct {
	Id              *int
	Username        string
	PasswordHash    []byte
	Email           string
	UnverifiedEmail string
}