package repository

import (
	"context"
	"product-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OrderRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, order *domain.Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
}
