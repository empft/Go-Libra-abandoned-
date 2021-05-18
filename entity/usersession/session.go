package entity

// Session related data should implement this interface
type Session interface {
	UserId() int
}

type UserSession struct {
	Id int
	Token string
}

type UserSessionRepository interface {
	Store(session UserSession, expire int) error
	Fetch(id int) UserSession
}