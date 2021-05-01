package entity

type InvitationalEmail struct {
	Email string
	Code string
}

type InvitationalEmailRepository interface {
	Store(invitation InvitationalEmail) error
	Fetch(email string) (string, error)
}