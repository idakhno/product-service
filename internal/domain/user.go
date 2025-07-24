package domain

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID
	Firstname    string
	Lastname     string
	Email        string
	Age          int
	IsMarried    bool
	PasswordHash string
}

func (u *User) FullName() string {
	return u.Firstname + " " + u.Lastname
}
