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
	ErrInsufficientStock = errors.New("insufficient stock for a product")
	ErrProductNotFound   = errors.New("product not found")
)

type OrderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	db          *pgxpool.Pool
	logger      logger.Logger
}

func NewOrderService(db *pgxpool.Pool, orderRepo repository.OrderRepository, productRepo repository.ProductRepository, logger logger.Logger) *OrderService {
	return &OrderService{
		db:          db,
		orderRepo:   orderRepo,
		productRepo: productRepo,
		logger:      logger,
	}
}

type OrderItemInput struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, items []OrderItemInput) (*domain.Order, error) {
	const op = "OrderService.CreateOrder"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
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

	for _, item := range items {
		product, err := s.productRepo.FindByIDTx(ctx, tx, item.ProductID)
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				return nil, ErrProductNotFound
			}
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if product.Quantity < item.Quantity {
			return nil, fmt.Errorf("%w: insufficient stock for product %s", ErrInsufficientStock, product.ID)
		}

		product.Quantity -= item.Quantity
		if err = s.productRepo.UpdateTx(ctx, tx, product); err != nil {
			return nil, fmt.Errorf("could not update product quantity: %w", err)
		}

		orderItem := domain.OrderItem{
			ID:              uuid.New(),
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			PriceAtPurchase: product.Price,
		}
		order.Items = append(order.Items, orderItem)
		totalAmount += product.Price * float64(item.Quantity)
	}

	order.TotalAmount = totalAmount

	if err = s.orderRepo.CreateTx(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("could not create order: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}

	return order, nil
}
