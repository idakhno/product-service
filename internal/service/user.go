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
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UsersService struct {
	repo      repository.UserRepository
	jwtSecret []byte
	jwtTTL    time.Duration
}

func NewUsersService(repo repository.UserRepository, jwtSecret []byte, jwtTTL time.Duration) *UsersService {
	return &UsersService{repo: repo, jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

func (s *UsersService) Register(ctx context.Context, email, password, firstname, lastname string, age int, isMarried bool) (*domain.User, error) {
	_, err := s.repo.FindByEmail(ctx, email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(passwordHash),
		Firstname:    firstname,
		Lastname:     lastname,
		Age:          age,
		IsMarried:    isMarried,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UsersService) Login(ctx context.Context, email, password string) (string, error) {
	const op = "UsersService.Login"

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

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
