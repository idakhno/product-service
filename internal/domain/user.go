package domain

import "github.com/google/uuid"

// User represents a user in the system.
type User struct {
	ID           uuid.UUID
	Firstname    string
	Lastname     string
	Email        string
	Age          int
	IsMarried    bool
	PasswordHash string // Password hash (bcrypt)
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.Firstname + " " + u.Lastname
}
