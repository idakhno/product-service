package repository

import (
	"context"
	"errors"
	"product-api/internal/domain"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserRepository defines the interface for user data persistence.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}
