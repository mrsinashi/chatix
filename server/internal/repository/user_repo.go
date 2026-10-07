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
		INSERT INTO users (username, username_normalized, kind, password_hash, display_name, email, phone, department_id, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	var user models.User
	user.Username = input.Username
	user.UsernameNormalized = normalized
	user.Kind = models.UserKindPerson
	user.PasswordHash = hash
	user.DisplayName = input.DisplayName
	user.Email = input.Email
	user.Phone = input.Phone
	user.DepartmentID = input.DepartmentID
	user.Status = input.Status

	err = r.pool.QueryRow(ctx, query,
		user.Username, user.UsernameNormalized, user.Kind, user.PasswordHash,
		user.DisplayName, user.Email, user.Phone, user.DepartmentID, user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("создание пользователя: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, username_normalized, kind, password_hash, display_name, email, phone, department_id, status, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.UsernameNormalized, &user.Kind, &user.PasswordHash,
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
		SELECT id, username, username_normalized, kind, password_hash, display_name, email, phone, department_id, status, created_at, updated_at, last_login_at
		FROM users
		WHERE username_normalized = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, normalized).Scan(
		&user.ID, &user.Username, &user.UsernameNormalized, &user.Kind, &user.PasswordHash,
		&user.DisplayName, &user.Email, &user.Phone, &user.DepartmentID,
		&user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err != nil {
		return nil, fmt.Errorf("получение пользователя: %w", err)
	}

	return &user, nil
}

// List возвращает список пользователей с фильтрами
func (r *UserRepository) List(ctx context.Context, query string, kind *string, status *string, limit, offset int) ([]*models.User, int, error) {
	baseQuery := `
		SELECT id, username, username_normalized, kind, password_hash, display_name, 
		       email, phone, department_id, status, created_at, updated_at, last_login_at
		FROM users
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`

	args := []interface{}{}
	argPos := 1

	if query != "" {
		baseQuery += fmt.Sprintf(" AND (username ILIKE $%d OR display_name ILIKE $%d OR email ILIKE $%d)", argPos, argPos, argPos)
		countQuery += fmt.Sprintf(" AND (username ILIKE $%d OR display_name ILIKE $%d OR email ILIKE $%d)", argPos, argPos, argPos)
		args = append(args, "%"+query+"%")
		argPos++
	}

	if kind != nil && *kind != "" {
		baseQuery += fmt.Sprintf(" AND kind = $%d", argPos)
		countQuery += fmt.Sprintf(" AND kind = $%d", argPos)
		args = append(args, *kind)
		argPos++
	}

	if status != nil && *status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argPos)
		countQuery += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *status)
		argPos++
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчет пользователей: %w", err)
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("запрос списка пользователей: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.UsernameNormalized, &user.Kind, &user.PasswordHash,
			&user.DisplayName, &user.Email, &user.Phone, &user.DepartmentID,
			&user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("чтение пользователя: %w", err)
		}
		users = append(users, &user)
	}

	return users, total, nil
}

// Update обновляет данные пользователя
func (r *UserRepository) Update(ctx context.Context, id string, input models.UserUpdate) (*models.User, error) {
	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if input.DisplayName != nil {
		updates = append(updates, fmt.Sprintf("display_name = $%d", argPos))
		args = append(args, *input.DisplayName)
		argPos++
	}
	if input.Email != nil {
		updates = append(updates, fmt.Sprintf("email = $%d", argPos))
		args = append(args, *input.Email)
		argPos++
	}
	if input.Phone != nil {
		updates = append(updates, fmt.Sprintf("phone = $%d", argPos))
		args = append(args, *input.Phone)
		argPos++
	}
	if input.DepartmentID != nil {
		updates = append(updates, fmt.Sprintf("department_id = $%d", argPos))
		args = append(args, *input.DepartmentID)
		argPos++
	}
	if input.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *input.Status)
		argPos++
	}

	if len(updates) == 0 {
		return r.GetByID(ctx, id)
	}

	updates = append(updates, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", joinStrings(updates, ", "), argPos)
	args = append(args, id)

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("обновление пользователя: %w", err)
	}

	return r.GetByID(ctx, id)
}

// SetPassword устанавливает новый пароль
func (r *UserRepository) SetPassword(ctx context.Context, userID, newPassword string) error {
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("хеширование пароля: %w", err)
	}

	query := `UPDATE users SET password_hash = $1, must_change_password = false, updated_at = NOW() WHERE id = $2`
	_, err = r.pool.Exec(ctx, query, hash, userID)
	if err != nil {
		return fmt.Errorf("установка пароля: %w", err)
	}

	return nil
}

// CountActiveAdmins возвращает количество активных администраторов
func (r *UserRepository) CountActiveAdmins(ctx context.Context, excludeUserID string) (int, error) {
	query := `
		SELECT COUNT(*) FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		WHERE r.name = 'Администратор' AND u.status = 'active' AND u.id != $1
	`

	var count int
	err := r.pool.QueryRow(ctx, query, excludeUserID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("подсчет администраторов: %w", err)
	}

	return count, nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("обновление времени входа: %w", err)
	}
	return nil
}