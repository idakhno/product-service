package service

import (
	"context"
	"errors"
	"fmt"
	"product-api/internal/domain"
	"product-api/internal/logger"
	"product-api/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrInsufficientStock is returned when there is insufficient stock to create an order.
	ErrInsufficientStock = errors.New("insufficient stock for a product")
)

// OrderService provides business logic for order operations.
// Uses transactions to ensure data integrity when creating orders.
type OrderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	db          *pgxpool.Pool
	logger      logger.Logger
}

// NewOrderService creates a new order service.
func NewOrderService(db *pgxpool.Pool, orderRepo repository.OrderRepository, productRepo repository.ProductRepository, logger logger.Logger) *OrderService {
	return &OrderService{
		db:          db,
		orderRepo:   orderRepo,
		productRepo: productRepo,
		logger:      logger,
	}
}

// OrderItemInput contains information about a single item in an order.
type OrderItemInput struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

// CreateOrder creates a new order for a user.
// Uses a transaction to ensure atomicity of operations:
// - Check product availability in stock
// - Update product quantities
// - Create order and order items
// On any error, the transaction is rolled back.
func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, items []OrderItemInput) (*domain.Order, error) {
	const op = "OrderService.CreateOrder"

	// Begin transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		// Rollback transaction on error
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.logger.Error("error rolling back transaction", "rollback_error", rbErr, "original_error", err)
			}
		}
	}()

	var totalAmount float64
	order := &domain.Order{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	// Process each item in the order
	for _, item := range items {
		// Get product with row lock (FOR UPDATE) to prevent race condition
		product, err := s.productRepo.FindByIDTx(ctx, tx, item.ProductID)
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				return nil, ErrProductNotFound
			}
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		// Check if sufficient quantity is available
		if product.Quantity < item.Quantity {
			return nil, fmt.Errorf("%w: insufficient stock for product %s", ErrInsufficientStock, product.ID)
		}

		// Decrease product quantity in stock
		product.Quantity -= item.Quantity
		if err = s.productRepo.UpdateTx(ctx, tx, product); err != nil {
			return nil, fmt.Errorf("could not update product quantity: %w", err)
		}

		// Add item to order
		orderItem := domain.OrderItem{
			ID:              uuid.New(),
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			PriceAtPurchase: product.Price, // Save price at time of purchase
		}
		order.Items = append(order.Items, orderItem)
		totalAmount += product.Price * float64(item.Quantity)
	}

	order.TotalAmount = totalAmount

	// Create order in database
	if err = s.orderRepo.CreateTx(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("could not create order: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}

	return order, nil
}
