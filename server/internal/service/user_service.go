package service

import (
	"context"
	"fmt"

	"chatix/internal/models"
	"chatix/internal/repository"
)

type UserService struct {
	userRepo    *repository.UserRepository
	roleRepo    *repository.RoleRepository
	auditRepo   *repository.AuditRepository
	sessionRepo *repository.SessionRepository
}

func NewUserService(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	auditRepo *repository.AuditRepository,
	sessionRepo *repository.SessionRepository,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		auditRepo:   auditRepo,
		sessionRepo: sessionRepo,
	}
}

// CreateUser создаёт пользователя и назначает роли
func (s *UserService) CreateUser(ctx context.Context, input models.UserCreate, roles []string, actorUserID *string, ipAddress *string) (*models.User, error) {
	if input.Status == "" {
		input.Status = models.UserStatusActive
	}

	user, err := s.userRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("создание пользователя: %w", err)
	}

	// Назначаем роли
	for _, roleName := range roles {
		if err := s.roleRepo.AssignRole(ctx, user.ID, roleName); err != nil {
			return nil, fmt.Errorf("назначение роли %s: %w", roleName, err)
		}
	}

	// Пишем в аудит
	s.auditRepo.Create(ctx, models.AuditLogCreate{
		UserID:     actorUserID,
		Action:     "user.created",
		EntityType: "user",
		EntityID:   &user.ID,
		IPAddress:  ipAddress,
	})

	return user, nil
}

// GetUser возвращает пользователя по ID
func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("получение пользователя: %w", err)
	}
	return user, nil
}

// GetUserPermissions возвращает права пользователя
func (s *UserService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	permissions, err := s.roleRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("получение прав: %w", err)
	}
	return permissions, nil
}

// ListUsers возвращает список пользователей с фильтрами
func (s *UserService) ListUsers(ctx context.Context, query string, kind *string, status *string, limit, offset int) ([]*models.User, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	users, total, err := s.userRepo.List(ctx, query, kind, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("получение списка пользователей: %w", err)
	}

	return users, total, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(ctx context.Context, id string, input models.UserUpdate, actorUserID *string, ipAddress *string) (*models.User, error) {
	user, err := s.userRepo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("обновление пользователя: %w", err)
	}

	s.auditRepo.Create(ctx, models.AuditLogCreate{
		UserID:     actorUserID,
		Action:     "user.updated",
		EntityType: "user",
		EntityID:   &id,
		IPAddress:  ipAddress,
	})

	return user, nil
}

// SetPassword устанавливает новый пароль пользователю
func (s *UserService) SetPassword(ctx context.Context, userID, newPassword string, actorUserID *string, ipAddress *string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("пароль должен быть не короче 8 символов")
	}

	err := s.userRepo.SetPassword(ctx, userID, newPassword)
	if err != nil {
		return fmt.Errorf("установка пароля: %w", err)
	}

	s.auditRepo.Create(ctx, models.AuditLogCreate{
		UserID:     actorUserID,
		Action:     "user.password_changed",
		EntityType: "user",
		EntityID:   &userID,
		IPAddress:  ipAddress,
	})

	return nil
}

// BlockUser блокирует пользователя
func (s *UserService) BlockUser(ctx context.Context, userID string, actorUserID *string, ipAddress *string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}

	// Проверяем, что не блокируем последнего админа
	if user.Kind == models.UserKindPerson {
		adminCount, err := s.userRepo.CountActiveAdmins(ctx, userID)
		if err != nil {
			return fmt.Errorf("проверка администраторов: %w", err)
		}
		if adminCount == 0 {
			return fmt.Errorf("нельзя заблокировать последнего активного администратора")
		}
	}

	status := models.UserStatusBlocked
	_, err = s.userRepo.Update(ctx, userID, models.UserUpdate{Status: &status})
	if err != nil {
		return fmt.Errorf("блокировка пользователя: %w", err)
	}

	// Отзываем все сессии
	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("отзыв сессий: %w", err)
	}

	s.auditRepo.Create(ctx, models.AuditLogCreate{
		UserID:     actorUserID,
		Action:     "user.blocked",
		EntityType: "user",
		EntityID:   &userID,
		IPAddress:  ipAddress,
	})

	return nil
}

// UnblockUser разблокирует пользователя
func (s *UserService) UnblockUser(ctx context.Context, userID string, actorUserID *string, ipAddress *string) error {
	status := models.UserStatusActive
	_, err := s.userRepo.Update(ctx, userID, models.UserUpdate{Status: &status})
	if err != nil {
		return fmt.Errorf("разблокировка пользователя: %w", err)
	}

	s.auditRepo.Create(ctx, models.AuditLogCreate{
		UserID:     actorUserID,
		Action:     "user.unblocked",
		EntityType: "user",
		EntityID:   &userID,
		IPAddress:  ipAddress,
	})

	return nil
}

// HasPermission проверяет наличие права у пользователя
func (s *UserService) HasPermission(ctx context.Context, userID, permissionCode string) (bool, error) {
	has, err := s.roleRepo.HasPermission(ctx, userID, permissionCode)
	if err != nil {
		return false, fmt.Errorf("проверка права: %w", err)
	}
	return has, nil
}


// GetUserSessions возвращает список сессий пользователя
func (s *UserService) GetUserSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	sessions, err := s.sessionRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("получение сессий: %w", err)
	}
	return sessions, nil
}

// RevokeSession отзывает конкретную сессию
func (s *UserService) RevokeSession(ctx context.Context, userID, sessionID string, actorUserID *string, ipAddress *string) error {
	// Проверяем, что сессия принадлежит пользователю
	sessions, err := s.sessionRepo.ListByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("получение сессий: %w", err)
	}

	found := false
	for _, sess := range sessions {
		if sess.ID == sessionID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("сессия не найдена")
	}

	err = s.sessionRepo.Delete(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("отзыв сессии: %w", err)
	}

	s.auditRepo.Create(ctx, models.AuditLogCreate{
		UserID:     actorUserID,
		Action:     "session.revoked",
		EntityType: "session",
		EntityID:   &sessionID,
		IPAddress:  ipAddress,
	})

	return nil
}

// RevokeAllSessions отзывает все сессии пользователя
func (s *UserService) RevokeAllSessions(ctx context.Context, userID string, actorUserID *string, ipAddress *string) error {
	err := s.sessionRepo.DeleteByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("отзыв сессий: %w", err)
	}

	s.auditRepo.Create(ctx, models.AuditLogCreate{
		UserID:     actorUserID,
		Action:     "session.revoke_all",
		EntityType: "user",
		EntityID:   &userID,
		IPAddress:  ipAddress,
	})

	return nil
}