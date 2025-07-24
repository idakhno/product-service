package service

import (
	"context"
	"errors"
	"product-api/internal/domain"
	"product-api/internal/repository"

	"github.com/google/uuid"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

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
