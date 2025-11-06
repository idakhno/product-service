package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"product-api/internal/domain"
	"product-api/internal/repository"
)

var (
	// ErrUserAlreadyExists is returned when attempting to register a user with an existing email.
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	// ErrUserNotFound is returned when user is not found in the database.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidCredentials is returned when authentication credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UsersService provides business logic for user operations.
type UsersService struct {
	repo      repository.UserRepository
	jwtSecret []byte
	jwtTTL    time.Duration
}

// NewUsersService creates a new users service.
func NewUsersService(repo repository.UserRepository, jwtSecret []byte, jwtTTL time.Duration) *UsersService {
	return &UsersService{repo: repo, jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

// Register registers a new user.
// Checks that a user with this email does not already exist,
// hashes the password and saves the user to the database.
func (s *UsersService) Register(ctx context.Context, email, password, firstname, lastname string, age int, isMarried bool) (*domain.User, error) {
	// Check if user with this email already exists
	_, err := s.repo.FindByEmail(ctx, email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	// Hash password before saving
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create new user
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(passwordHash),
		Firstname:    firstname,
		Lastname:     lastname,
		Age:          age,
		IsMarried:    isMarried,
	}

	// Save user to database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token.
// Validates email and password, generates JWT token on successful validation.
func (s *UsersService) Login(ctx context.Context, email, password string) (string, error) {
	const op = "UsersService.Login"

	// Find user by email
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(s.jwtTTL).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("%s: failed to sign token: %w", op, err)
	}

	return tokenString, nil
}
