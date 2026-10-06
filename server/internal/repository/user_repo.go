package repository

import (
	"context"
	"fmt"
	"time"

	"chatix/internal/auth"
	"chatix/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, input models.UserCreate) (*models.User, error) {
	normalized := auth.NormalizeUsername(input.Username)
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("хеширование пароля: %w", err)
	}

	query := `
		INSERT INTO users (username, username_normalized, password_hash, display_name, email, phone, department_id, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	var user models.User
	user.Username = input.Username
	user.UsernameNormalized = normalized
	user.PasswordHash = hash
	user.DisplayName = input.DisplayName
	user.Email = input.Email
	user.Phone = input.Phone
	user.DepartmentID = input.DepartmentID
	user.Status = input.Status

	err = r.pool.QueryRow(ctx, query,
		user.Username, user.UsernameNormalized, user.PasswordHash,
		user.DisplayName, user.Email, user.Phone, user.DepartmentID, user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("создание пользователя: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, username_normalized, password_hash, display_name, email, phone, department_id, status, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.UsernameNormalized, &user.PasswordHash,
		&user.DisplayName, &user.Email, &user.Phone, &user.DepartmentID,
		&user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err != nil {
		return nil, fmt.Errorf("получение пользователя: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	normalized := auth.NormalizeUsername(username)

	query := `
		SELECT id, username, username_normalized, password_hash, display_name, email, phone, department_id, status, created_at, updated_at, last_login_at
		FROM users
		WHERE username_normalized = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, normalized).Scan(
		&user.ID, &user.Username, &user.UsernameNormalized, &user.PasswordHash,
		&user.DisplayName, &user.Email, &user.Phone, &user.DepartmentID,
		&user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err != nil {
		return nil, fmt.Errorf("получение пользователя: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("обновление времени входа: %w", err)
	}
	return nil
}