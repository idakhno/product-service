package service_test

import (
	"context"
	"log"
	"os"
	"product-api/internal/domain"
	"product-api/internal/logger"
	"product-api/internal/repository"
	"product-api/internal/repository/postgres"
	"product-api/internal/service"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

type OrderServiceTestSuite struct {
	suite.Suite
	dbpool      *pgxpool.Pool
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	userRepo    repository.UserRepository
	service     *service.OrderService
}

func (s *OrderServiceTestSuite) SetupSuite() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME") + "_test"
	maintenanceDbUrl := "postgres://" + dbUser + ":" + dbPassword + "@localhost:5434/postgres?sslmode=disable"
	testDbUrl := "postgres://" + dbUser + ":" + dbPassword + "@localhost:5434/" + dbName + "?sslmode=disable"

	var err error
	var maintenanceDb *pgxpool.Pool

	for i := 0; i < 10; i++ {
		maintenanceDb, err = pgxpool.New(context.Background(), maintenanceDbUrl)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to maintenance db, retrying in 2 seconds...: %v", err)
		time.Sleep(2 * time.Second)
	}
	s.Require().NoError(err, "Failed to connect to maintenance database after retries")

	_, err = maintenanceDb.Exec(context.Background(), "DROP DATABASE IF EXISTS "+dbName)
	s.Require().NoError(err)
	_, err = maintenanceDb.Exec(context.Background(), "CREATE DATABASE "+dbName)
	s.Require().NoError(err)
	maintenanceDb.Close()

	s.dbpool, err = pgxpool.New(context.Background(), testDbUrl)
	s.Require().NoError(err)

	m, err := migrate.New("file://../../migrations", testDbUrl)
	s.Require().NoError(err)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		s.Require().NoError(err)
	}

	s.orderRepo = postgres.NewOrderRepository(s.dbpool)
	s.productRepo = postgres.NewProductRepository(s.dbpool)
	s.userRepo = postgres.NewUserRepository(s.dbpool)

	testLogger := logger.NewSlogAdapter("local")
	s.service = service.NewOrderService(s.dbpool, s.orderRepo, s.productRepo, testLogger)
}

func (s *OrderServiceTestSuite) TearDownSuite() {
	s.dbpool.Close()
}

func (s *OrderServiceTestSuite) TearDownTest() {
	_, err := s.dbpool.Exec(context.Background(), "TRUNCATE TABLE users, products, orders, order_items RESTART IDENTITY CASCADE")
	s.Require().NoError(err)
}

func (s *OrderServiceTestSuite) TestCreateOrder_Success() {
	ctx := context.Background()

	user := &domain.User{
		ID:        uuid.New(),
		Email:     "test-success@example.com",
		Firstname: "Test", Lastname: "User", Age: 30, IsMarried: false, PasswordHash: "hash",
	}
	s.Require().NoError(s.userRepo.Create(ctx, user))

	product := &domain.Product{
		ID:          uuid.New(),
		Description: "Test Product",
		Quantity:    10,
		Price:       99.99,
	}
	s.Require().NoError(s.productRepo.Create(ctx, product))

	items := []service.OrderItemInput{
		{ProductID: product.ID, Quantity: 3},
	}
	order, err := s.service.CreateOrder(ctx, user.ID, items)

	s.Assert().NoError(err)
	s.Assert().NotNil(order)
	s.Assert().Len(order.Items, 1)
	s.Assert().Equal(product.Price, order.Items[0].PriceAtPurchase)
	s.Assert().Equal(user.ID, order.UserID)

	updatedProduct, err := s.productRepo.FindByID(ctx, product.ID)
	s.Require().NoError(err)
	s.Assert().Equal(7, updatedProduct.Quantity)
}

func (s *OrderServiceTestSuite) TestCreateOrder_InsufficientStock() {
	ctx := context.Background()

	user := &domain.User{
		ID:        uuid.New(),
		Email:     "test-fail@example.com",
		Firstname: "Test", Lastname: "User", Age: 30, IsMarried: false, PasswordHash: "hash",
	}
	s.Require().NoError(s.userRepo.Create(ctx, user))

	product := &domain.Product{
		ID:          uuid.New(),
		Description: "Test Product Fail",
		Quantity:    5,
		Price:       10.00,
	}
	s.Require().NoError(s.productRepo.Create(ctx, product))

	items := []service.OrderItemInput{
		{ProductID: product.ID, Quantity: 10},
	}
	_, err := s.service.CreateOrder(ctx, user.ID, items)

	s.Assert().Error(err)
	s.Assert().ErrorIs(err, service.ErrInsufficientStock)

	updatedProduct, err := s.productRepo.FindByID(ctx, product.ID)
	s.Require().NoError(err)
	s.Assert().Equal(5, updatedProduct.Quantity)
}

func TestOrderServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrderServiceTestSuite))
}
