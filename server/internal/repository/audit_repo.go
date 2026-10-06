package repository

import (
	"context"
	"fmt"

	"chatix/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) Create(ctx context.Context, input models.AuditLogCreate) error {
	query := `
		INSERT INTO audit_log (user_id, action, entity_type, entity_id, details, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		input.UserID, input.Action, input.EntityType, input.EntityID,
		input.Details, input.IPAddress,
	)

	if err != nil {
		return fmt.Errorf("запись в аудит: %w", err)
	}

	return nil
}