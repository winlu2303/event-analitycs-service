package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/event-analytics-service/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.Role == "" {
		user.Role = "user"
	}

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, full_name, role, created_at, updated_at
        FROM users
        WHERE id = $1 AND deleted_at IS NULL
    `

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, full_name, role, created_at, updated_at
        FROM users
        WHERE email = $1 AND deleted_at IS NULL
    `

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
        UPDATE users
        SET email = $1, full_name = $2, role = $3, updated_at = $4
        WHERE id = $5 AND deleted_at IS NULL
    `

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.FullName,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}
