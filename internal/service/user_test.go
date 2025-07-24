package service_test

import (
	"context"
	"log"
	"os"
	"product-api/internal/domain"
	"product-api/internal/repository"
	"product-api/internal/repository/postgres"
	"product-api/internal/service"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceTestSuite struct {
	suite.Suite
	dbpool    *pgxpool.Pool
	userRepo  repository.UserRepository
	service   *service.UsersService
	jwtSecret []byte
}

func (s *UserServiceTestSuite) SetupSuite() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME") + "_test_user" // Use a dedicated test database
	maintenanceDbUrl := "postgres://" + dbUser + ":" + dbPassword + "@localhost:5434/postgres?sslmode=disable"
	testDbUrl := "postgres://" + dbUser + ":" + dbPassword + "@localhost:5434/" + dbName + "?sslmode=disable"

	var err error
	var maintenanceDb *pgxpool.Pool

	// Retry logic for connecting to the maintenance database
	for i := 0; i < 10; i++ {
		maintenanceDb, err = pgxpool.New(context.Background(), maintenanceDbUrl)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to maintenance db, retrying in 2 seconds...: %v", err)
		time.Sleep(2 * time.Second)
	}
	s.Require().NoError(err, "Failed to connect to maintenance database after retries")

	// Drop and create test database
	_, err = maintenanceDb.Exec(context.Background(), "DROP DATABASE IF EXISTS "+dbName)
	s.Require().NoError(err)
	_, err = maintenanceDb.Exec(context.Background(), "CREATE DATABASE "+dbName)
	s.Require().NoError(err)
	maintenanceDb.Close()

	// Connect to the new test database
	s.dbpool, err = pgxpool.New(context.Background(), testDbUrl)
	s.Require().NoError(err)

	// Run migrations on the test database
	m, err := migrate.New("file://../../migrations", testDbUrl)
	s.Require().NoError(err)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		s.Require().NoError(err)
	}

	s.userRepo = postgres.NewUserRepository(s.dbpool)
	s.jwtSecret = []byte("test-secret")
	s.service = service.NewUsersService(s.userRepo, s.jwtSecret, time.Hour)
}

func (s *UserServiceTestSuite) TearDownSuite() {
	s.dbpool.Close()
}

func (s *UserServiceTestSuite) TearDownTest() {
	_, err := s.dbpool.Exec(context.Background(), "TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	s.Require().NoError(err)
}

func (s *UserServiceTestSuite) TestRegister_Success() {
	ctx := context.Background()
	user, err := s.service.Register(ctx, "test@example.com", "password123", "John", "Doe", 25, false)
	s.NoError(err)
	s.NotNil(user)
	dbUser, err := s.userRepo.FindByEmail(ctx, "test@example.com")
	s.NoError(err)
	s.Equal(user.ID, dbUser.ID)
}

func (s *UserServiceTestSuite) TestRegister_UserAlreadyExists() {
	ctx := context.Background()
	existingUser := &domain.User{
		ID:           uuid.New(),
		Email:        "exists@example.com",
		PasswordHash: "somehash",
	}
	s.Require().NoError(s.userRepo.Create(ctx, existingUser))
	_, err := s.service.Register(ctx, "exists@example.com", "password123", "John", "Doe", 25, false)
	s.ErrorIs(err, service.ErrUserAlreadyExists)
}

func (s *UserServiceTestSuite) TestLogin_Success() {
	ctx := context.Background()
	email := "login@example.com"
	password := "password123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	s.Require().NoError(err)
	user := &domain.User{ID: uuid.New(), Email: email, PasswordHash: string(passwordHash)}
	s.Require().NoError(s.userRepo.Create(ctx, user))
	token, err := s.service.Login(ctx, email, password)
	s.NoError(err)
	s.NotEmpty(token)

	// Verify JWT claims
	tokenClaims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, tokenClaims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	s.Require().NoError(err)
	s.Require().True(parsedToken.Valid)
	s.Equal(user.ID.String(), tokenClaims["sub"])
	s.InDelta(time.Now().Add(time.Hour).Unix(), tokenClaims["exp"], 10) // Check exp is roughly correct
}

func (s *UserServiceTestSuite) TestLogin_UserNotFound() {
	ctx := context.Background()
	_, err := s.service.Login(ctx, "nonexistent@example.com", "password123")
	s.ErrorIs(err, service.ErrInvalidCredentials)
}

func (s *UserServiceTestSuite) TestLogin_InvalidPassword() {
	ctx := context.Background()
	email := "invalidpass@example.com"
	password := "password123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	s.Require().NoError(err)
	user := &domain.User{ID: uuid.New(), Email: email, PasswordHash: string(passwordHash)}
	s.Require().NoError(s.userRepo.Create(ctx, user))
	_, err = s.service.Login(ctx, email, "wrongpassword")
	s.ErrorIs(err, service.ErrInvalidCredentials)
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
