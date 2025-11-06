package repository

import (
	"context"
	"errors"
	"product-api/internal/domain"

	"github.com/google/uuid"
)

var (
	// ErrUserNotFound is returned when user is not found in the database.
	ErrUserNotFound = errors.New("user not found")
)

// UserRepository defines the interface for user database operations.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}
