package account

// Basic account structure
type Base struct {
	Id              *int
	Username        string
	PasswordHash    []byte
	Email           string
	UnverifiedEmail string
}