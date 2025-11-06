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

// UserRepository implements repository.UserRepository interface for PostgreSQL.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository for PostgreSQL.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, firstname, lastname, email, age, is_married, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, user.ID, user.Firstname, user.Lastname, user.Email, user.Age, user.IsMarried, user.PasswordHash)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, firstname, lastname, email, age, is_married, password_hash
			  FROM users WHERE id = $1`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Email, &user.Age, &user.IsMarried, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, firstname, lastname, email, age, is_married, password_hash
		FROM users
		WHERE email = $1
	`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Firstname,
		&user.Lastname,
		&user.Email,
		&user.Age,
		&user.IsMarried,
		&user.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
