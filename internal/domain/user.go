package domain

import (
	"time"
)

type User struct {
	id        UserID
	email     string
	password  string
	subsc     bool
	createdAt time.Time
}

func NewUser(id UserID, email string, password string, subsc bool, createdAt time.Time) User {
	return User{
		id:        id,
		email:     email,
		password:  password,
		subsc:     subsc,
		createdAt: createdAt,
	}
}

func (u User) ID() UserID {
	return u.id
}

func (u User) Email() string {
	return u.email
}

func (u User) Password() string {
	return u.password
}

func (u User) Subsc() bool {
	return u.subsc
}

func (u User) CreatedAt() time.Time {
	return u.createdAt
}

type UserRepository interface {
	Create(user User) (User, error)
	FindByEmail(email string) (User, error)
	FindByID(id UserID) (User, error)
}
