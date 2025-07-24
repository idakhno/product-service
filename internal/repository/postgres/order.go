package postgres

import (
	"context"
	"product-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateTx(ctx context.Context, tx pgx.Tx, order *domain.Order) error {
	orderQuery := `INSERT INTO orders (id, user_id, created_at, total_amount) VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(ctx, orderQuery, order.ID, order.UserID, order.CreatedAt, order.TotalAmount)
	if err != nil {
		return err
	}

	itemQuery := `INSERT INTO order_items (id, order_id, product_id, quantity, price_at_purchase)
				  VALUES ($1, $2, $3, $4, $5)`
	for _, item := range order.Items {
		_, err := tx.Exec(ctx, itemQuery, item.ID, order.ID, item.ProductID, item.Quantity, item.PriceAtPurchase)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `
        SELECT id, user_id, created_at, total_amount
        FROM orders
        WHERE id = $1
    `
	order := &domain.Order{}
	err := r.db.QueryRow(ctx, query, id).Scan(&order.ID, &order.UserID, &order.CreatedAt, &order.TotalAmount)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
        SELECT id, product_id, quantity, price_at_purchase
        FROM order_items
        WHERE order_id = $1
    `
	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := domain.OrderItem{}
		err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.PriceAtPurchase)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}
