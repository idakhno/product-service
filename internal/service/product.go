package service

import (
	"context"
	"errors"
	"product-api/internal/domain"
	"product-api/internal/repository"

	"github.com/google/uuid"
)

var (
	// ErrProductNotFound is returned when product is not found in the database.
	ErrProductNotFound = errors.New("product not found")
)

// ProductService provides business logic for product operations.
type ProductService struct {
	repo repository.ProductRepository
}

// NewProductService creates a new product service.
func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// CreateProduct creates a new product in the database.
func (s *ProductService) CreateProduct(ctx context.Context, description string, tags []string, quantity int, price float64) (*domain.Product, error) {
	product := &domain.Product{
		ID:          uuid.New(),
		Description: description,
		Tags:        tags,
		Quantity:    quantity,
		Price:       price,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// GetProductByID retrieves a product by its ID.
// Returns ErrProductNotFound if product is not found.
func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return product, nil
}
