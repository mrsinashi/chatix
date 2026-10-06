package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"chatix/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, userID, token, userAgent, ipAddress string, expiresAt time.Time) (*models.Session, error) {
	tokenHash := hashToken(token)

	query := `
		INSERT INTO sessions (user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	var session models.Session
	session.UserID = userID
	session.TokenHash = tokenHash
	if userAgent != "" {
		session.UserAgent = &userAgent
	}
	if ipAddress != "" {
		session.IPAddress = &ipAddress
	}
	session.ExpiresAt = expiresAt

	err := r.pool.QueryRow(ctx, query, userID, tokenHash, session.UserAgent, session.IPAddress, expiresAt).
		Scan(&session.ID, &session.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("создание сессии: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	tokenHash := hashToken(token)

	query := `
		SELECT id, user_id, token_hash, user_agent, ip_address, created_at, expires_at
		FROM sessions
		WHERE token_hash = $1
	`

	var session models.Session
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID, &session.UserID, &session.TokenHash,
		&session.UserAgent, &session.IPAddress, &session.CreatedAt, &session.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("получение сессии: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("удаление сессии: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("удаление сессий пользователя: %w", err)
	}
	return nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}