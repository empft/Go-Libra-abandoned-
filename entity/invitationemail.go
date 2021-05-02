package entity

type InvitationEmail struct {
	Email string
	Code string
}

type InvitationEmailRepository interface {
	Store(invitation InvitationEmail) error
	Fetch(email string) (string, error)
}