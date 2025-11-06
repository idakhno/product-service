package postgres

import (
	"context"
	"errors"
	"product-api/internal/domain"
	"product-api/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductRepository implements repository.ProductRepository interface for PostgreSQL.
type ProductRepository struct {
	db *pgxpool.Pool
}

// NewProductRepository creates a new product repository for PostgreSQL.
func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `INSERT INTO products (id, description, tags, quantity, price)
			  VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, product.ID, product.Description, product.Tags, product.Quantity, product.Price)
	return err
}

func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	query := `SELECT id, description, tags, quantity, price FROM products WHERE id = $1`

	p := &domain.Product{}
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.Description, &p.Tags, &p.Quantity, &p.Price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrProductNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *ProductRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Product, error) {
	rows, err := r.db.Query(ctx, "SELECT id, description, tags, quantity, price FROM products WHERE id = ANY($1)", ids)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrProductNotFound
		}
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Description, &p.Tags, &p.Quantity, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, repository.ErrProductNotFound
	}

	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `UPDATE products SET description = $2, tags = $3, quantity = $4, price = $5 WHERE id = $1`

	_, err := r.db.Exec(ctx, query, product.ID, product.Description, product.Tags, product.Quantity, product.Price)
	return err
}

// FindByIDTx finds a product by ID within a transaction with row lock (FOR UPDATE).
// Used to prevent race conditions when updating product quantity.
func (r *ProductRepository) FindByIDTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Product, error) {
	query := `SELECT id, description, tags, quantity, price FROM products WHERE id = $1 FOR UPDATE`

	p := &domain.Product{}
	err := tx.QueryRow(ctx, query, id).Scan(&p.ID, &p.Description, &p.Tags, &p.Quantity, &p.Price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrProductNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *ProductRepository) UpdateTx(ctx context.Context, tx pgx.Tx, product *domain.Product) error {
	query := `UPDATE products SET quantity = $2 WHERE id = $1`

	_, err := tx.Exec(ctx, query, product.ID, product.Quantity)
	return err
}
