package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func (r *RoleRepository) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT DISTINCT p.code
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("получение прав: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("чтение права: %w", err)
		}
		permissions = append(permissions, code)
	}

	return permissions, nil
}

func (r *RoleRepository) HasPermission(ctx context.Context, userID, permissionCode string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM permissions p
			JOIN role_permissions rp ON rp.permission_id = p.id
			JOIN user_roles ur ON ur.role_id = rp.role_id
			WHERE ur.user_id = $1 AND p.code = $2
		)
	`

	var has bool
	err := r.pool.QueryRow(ctx, query, userID, permissionCode).Scan(&has)
	if err != nil {
		return false, fmt.Errorf("проверка права: %w", err)
	}

	return has, nil
}

func (r *RoleRepository) AssignRole(ctx context.Context, userID, roleName string) error {
	query := `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = $2
		ON CONFLICT DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, roleName)
	if err != nil {
		return fmt.Errorf("назначение роли: %w", err)
	}

	return nil
}