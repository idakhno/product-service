package repository

import (
	"context"
	"errors"
	"product-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	UpdateTx(ctx context.Context, tx pgx.Tx, product *domain.Product) error
	FindByIDTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Product, error)
}
