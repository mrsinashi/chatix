package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"chatix/internal/auth"
	"chatix/internal/models"
	"chatix/internal/repository"

	"github.com/valkey-io/valkey-go"
)

var (
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	ErrUserLocked         = errors.New("пользователь заблокирован")
	ErrTooManyAttempts    = errors.New("слишком много попыток входа, попробуйте позже")
	ErrSessionNotFound    = errors.New("сессия не найдена")
)

type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	roleRepo    *repository.RoleRepository
	auditRepo   *repository.AuditRepository
	valkey      valkey.Client

	maxAttempts    int
	lockDuration   time.Duration
	sessionTTL     time.Duration
}

func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	roleRepo *repository.RoleRepository,
	auditRepo *repository.AuditRepository,
	valkey valkey.Client,
	maxAttempts int,
	lockDuration time.Duration,
	sessionTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		roleRepo:     roleRepo,
		auditRepo:    auditRepo,
		valkey:       valkey,
		maxAttempts:  maxAttempts,
		lockDuration: lockDuration,
		sessionTTL:   sessionTTL,
	}
}

// Login проверяет логин/пароль и создаёт сессию
func (s *AuthService) Login(ctx context.Context, username, password, userAgent, ipAddress string) (*models.Session, *models.User, error) {
	normalized := auth.NormalizeUsername(username)

	// Проверяем, не заблокирован ли пользователь из-за попыток
	if s.isLocked(ctx, normalized) {
		return nil, nil, ErrTooManyAttempts
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.recordFailedAttempt(ctx, normalized)
		return nil, nil, ErrInvalidCredentials
	}

	if user.Status == models.UserStatusBlocked {
		return nil, nil, ErrUserLocked
	}

	if user.Status == models.UserStatusArchived {
		return nil, nil, ErrInvalidCredentials
	}

	ok, err := auth.VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		s.recordFailedAttempt(ctx, normalized)
		s.auditRepo.Create(ctx, models.AuditLogCreate{
			UserID:     &user.ID,
			Action:     "login.failed",
			EntityType: "user",
			EntityID:   &user.ID,
			IPAddress:  &ipAddress,
		})
		return nil, nil, ErrInvalidCredentials
	}

	// Сбрасываем счётчик неудачных попыток
	s.clearAttempts(ctx, normalized)

	// Генерируем токен сессии
	token, err := generateToken()
	if err != nil {
		return nil, nil, fmt.Errorf("генерация токена: %w", err)
	}

	expiresAt := time.Now().Add(s.sessionTTL)
	session, err := s.sessionRepo.Create(ctx, user.ID, token, userAgent, ipAddress, expiresAt)
	if err != nil {
		return nil, nil, fmt.Errorf("создание сессии: %w", err)
	}

	// Обновляем время последнего входа
	s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Пишем в аудит
	s.auditRepo.Create(ctx, models.AuditLogCreate{
		UserID:     &user.ID,
		Action:     "login.success",
		EntityType: "user",
		EntityID:   &user.ID,
		IPAddress:  &ipAddress,
	})

	// Сохраняем токен в сессии (для возврата клиенту)
	session.TokenHash = token

	return session, user, nil
}

// ValidateSession проверяет токен сессии и возвращает пользователя
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*models.User, error) {
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		s.sessionRepo.Delete(ctx, session.ID)
		return nil, ErrSessionNotFound
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if user.Status != models.UserStatusActive {
		return nil, ErrUserLocked
	}

	return user, nil
}

// Logout удаляет сессию
func (s *AuthService) Logout(ctx context.Context, token string) error {
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return ErrSessionNotFound
	}

	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		return fmt.Errorf("удаление сессии: %w", err)
	}

	s.auditRepo.Create(ctx, models.AuditLogCreate{
		UserID:     &session.UserID,
		Action:     "logout",
		EntityType: "session",
		EntityID:   &session.ID,
	})

	return nil
}

// RevokeAllSessions отзывает все сессии пользователя
func (s *AuthService) RevokeAllSessions(ctx context.Context, userID string) error {
	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("отзыв сессий: %w", err)
	}
	return nil
}

// isLocked проверяет, заблокирован ли вход из-за неудачных попыток
func (s *AuthService) isLocked(ctx context.Context, username string) bool {
	key := fmt.Sprintf("login:attempts:%s", username)
	resp := s.valkey.Do(ctx, s.valkey.B().Get().Key(key).Build())
	if resp.Error() != nil {
		return false
	}
	attempts, err := resp.AsInt64()
	if err != nil {
		return false
	}
	return attempts >= int64(s.maxAttempts)
}

// recordFailedAttempt увеличивает счётчик неудачных попыток
func (s *AuthService) recordFailedAttempt(ctx context.Context, username string) {
	key := fmt.Sprintf("login:attempts:%s", username)
	s.valkey.Do(ctx, s.valkey.B().Incr().Key(key).Build())
	s.valkey.Do(ctx, s.valkey.B().Expire().Key(key).Seconds(int64(s.lockDuration.Seconds())).Build())
}

// clearAttempts сбрасывает счётчик
func (s *AuthService) clearAttempts(ctx context.Context, username string) {
	key := fmt.Sprintf("login:attempts:%s", username)
	s.valkey.Do(ctx, s.valkey.B().Del().Key(key).Build())
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}