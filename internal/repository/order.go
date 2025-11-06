package repository

import (
	"context"
	"product-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// OrderRepository defines the interface for order database operations.
// CreateTx works within a transaction to ensure operation atomicity.
type OrderRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, order *domain.Order) error // Create order within transaction
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
}
