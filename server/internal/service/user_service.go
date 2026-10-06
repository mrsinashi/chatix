package service

import (
	"context"
	"fmt"

	"chatix/internal/models"
	"chatix/internal/repository"
)

type UserService struct {
	userRepo  *repository.UserRepository
	roleRepo  *repository.RoleRepository
	auditRepo *repository.AuditRepository
}

func NewUserService(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	auditRepo *repository.AuditRepository,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		auditRepo: auditRepo,
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

// HasPermission проверяет наличие права у пользователя
func (s *UserService) HasPermission(ctx context.Context, userID, permissionCode string) (bool, error) {
	has, err := s.roleRepo.HasPermission(ctx, userID, permissionCode)
	if err != nil {
		return false, fmt.Errorf("проверка права: %w", err)
	}
	return has, nil
}