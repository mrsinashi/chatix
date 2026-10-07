package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"chatix/internal/audit"
	"chatix/internal/auth"
	"chatix/internal/db"
	"chatix/internal/perms"

	"github.com/google/uuid"
)

type AdminService struct {
	db       *db.DB
	auth     *auth.Service
	audit    *audit.Service
	perms    *perms.Checker
	userSvc  *Service
}

func NewAdminService(database *db.DB, authService *auth.Service, auditSvc *audit.Service, permChecker *perms.Checker, userService *Service) *AdminService {
	return &AdminService{
		db:      database,
		auth:    authService,
		audit:   auditSvc,
		perms:   permChecker,
		userSvc: userService,
	}
}

type CreateUserRequest struct {
	Login        string  `json:"login"`
	DisplayName  string  `json:"display_name"`
	Kind         string  `json:"kind"` // person, room, role
	Email        *string `json:"email,omitempty"`
	Password     string  `json:"password"`
	RoomText     *string `json:"room_text,omitempty"`
	RoleTitle    *string `json:"role_title,omitempty"`
	ExtNumber    *string `json:"ext_number,omitempty"`
}

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	RoomText    *string `json:"room_text,omitempty"`
	RoleTitle   *string `json:"role_title,omitempty"`
	Status      *string `json:"status,omitempty"` // active, blocked, archived
}

type UserListItem struct {
	ID           uuid.UUID `json:"id"`
	Login        string    `json:"login"`
	DisplayName  string    `json:"display_name"`
	Kind         string    `json:"kind"`
	Email        *string   `json:"email,omitempty"`
	RoomText     *string   `json:"room_text,omitempty"`
	RoleTitle    *string   `json:"role_title,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
}

type UserListResponse struct {
	Users []UserListItem `json:"users"`
	Total int            `json:"total"`
}

// ListUsers возвращает список пользователей с фильтрами
func (s *AdminService) ListUsers(ctx context.Context, actorID uuid.UUID, query string, kind *string, status *string, limit, offset int) (*UserListResponse, error) {
	if !s.perms.Can(ctx, actorID, "user.manage", nil) {
		return nil, errors.New("insufficient permissions")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var users []UserListItem
	var total int

	// Базовый запрос
	baseQuery := `
		SELECT id, login, display_name, kind, email, room_text, role_title, status, created_at, last_seen_at
		FROM users
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`

	args := []interface{}{}
	argPos := 1

	if query != "" {
		baseQuery += fmt.Sprintf(" AND (login ILIKE $%d OR display_name ILIKE $%d OR email ILIKE $%d)", argPos, argPos, argPos)
		countQuery += fmt.Sprintf(" AND (login ILIKE $%d OR display_name ILIKE $%d OR email ILIKE $%d)", argPos, argPos, argPos)
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

	// Подсчет общего количества
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Добавляем пагинацию
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := s.db.Pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user UserListItem
		err := rows.Scan(
			&user.ID, &user.Login, &user.DisplayName, &user.Kind,
			&user.Email, &user.RoomText, &user.RoleTitle, &user.Status,
			&user.CreatedAt, &user.LastSeenAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return &UserListResponse{
		Users: users,
		Total: total,
	}, nil
}

// CreateUser создает нового пользователя
func (s *AdminService) CreateUser(ctx context.Context, actorID uuid.UUID, req CreateUserRequest) (*User, error) {
	if !s.perms.Can(ctx, actorID, "user.manage", nil) {
		return nil, errors.New("insufficient permissions")
	}

	// Валидация
	if req.Login == "" {
		return nil, errors.New("login is required")
	}
	if req.DisplayName == "" {
		return nil, errors.New("display name is required")
	}
	if req.Kind == "" {
		req.Kind = "person"
	}
	if req.Kind != "person" && req.Kind != "room" && req.Kind != "role" {
		return nil, errors.New("invalid kind: must be person, room, or role")
	}
	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	// Создаем пользователя через основной сервис
	user, err := s.userSvc.Create(ctx, req.Login, req.DisplayName, req.Kind, req.Email, req.Password, req.RoomText, req.RoleTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Если указан внутренний номер, добавляем его
	if req.ExtNumber != nil && *req.ExtNumber != "" {
		_, err := s.db.Pool.Exec(ctx, `
			INSERT INTO user_contacts (user_id, type, value, visibility, sort)
			VALUES ($1, 'ext', $2, 'all', 0)
		`, user.ID, *req.ExtNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to add ext number: %w", err)
		}
	}

	// Записываем в аудит
	s.audit.Log(ctx, actorID, "user.create", "user", user.ID, map[string]interface{}{
		"login":        user.Login,
		"display_name": user.DisplayName,
		"kind":         user.Kind,
	}, "")

	return user, nil
}

// UpdateUser обновляет данные пользователя
func (s *AdminService) UpdateUser(ctx context.Context, actorID, userID uuid.UUID, req UpdateUserRequest) (*User, error) {
	if !s.perms.Can(ctx, actorID, "user.manage", nil) {
		return nil, errors.New("insufficient permissions")
	}

	// Получаем текущие данные
	user, err := s.userSvc.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Строим UPDATE запрос
	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if req.DisplayName != nil {
		updates = append(updates, fmt.Sprintf("display_name = $%d", argPos))
		args = append(args, *req.DisplayName)
		argPos++
	}
	if req.Email != nil {
		updates = append(updates, fmt.Sprintf("email = $%d", argPos))
		args = append(args, *req.Email)
		argPos++
	}
	if req.RoomText != nil {
		updates = append(updates, fmt.Sprintf("room_text = $%d", argPos))
		args = append(args, *req.RoomText)
		argPos++
	}
	if req.RoleTitle != nil {
		updates = append(updates, fmt.Sprintf("role_title = $%d", argPos))
		args = append(args, *req.RoleTitle)
		argPos++
	}
	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "blocked" && *req.Status != "archived" {
			return nil, errors.New("invalid status")
		}
		updates = append(updates, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *req.Status)
		argPos++
	}

	if len(updates) == 0 {
		return user, nil
	}

	updates = append(updates, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", 
		joinStrings(updates, ", "), argPos)
	args = append(args, userID)

	_, err = s.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Получаем обновленные данные
	user, err = s.userSvc.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	// Записываем в аудит
	changes := map[string]interface{}{}
	if req.DisplayName != nil {
		changes["display_name"] = *req.DisplayName
	}
	if req.Email != nil {
		changes["email"] = *req.Email
	}
	if req.RoomText != nil {
		changes["room_text"] = *req.RoomText
	}
	if req.RoleTitle != nil {
		changes["role_title"] = *req.RoleTitle
	}
	if req.Status != nil {
		changes["status"] = *req.Status
	}

	s.audit.Log(ctx, actorID, "user.update", "user", userID, changes, "")

	return user, nil
}

// SetPassword устанавливает новый пароль пользователю
func (s *AdminService) SetPassword(ctx context.Context, actorID, userID uuid.UUID, newPassword string) error {
	if !s.perms.Can(ctx, actorID, "user.manage", nil) {
		return errors.New("insufficient permissions")
	}

	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	user, err := s.userSvc.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	err = s.userSvc.SetPassword(ctx, userID, newPassword)
	if err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	s.audit.Log(ctx, actorID, "user.set_password", "user", userID, map[string]interface{}{
		"login": user.Login,
	}, "")

	return nil
}

// BlockUser блокирует пользователя
func (s *AdminService) BlockUser(ctx context.Context, actorID, userID uuid.UUID) error {
	if !s.perms.Can(ctx, actorID, "user.manage", nil) {
		return errors.New("insufficient permissions")
	}

	user, err := s.userSvc.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Проверяем, что не блокируем последнего админа
	if user.Kind == "person" {
		var adminCount int
		err := s.db.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM users u
			JOIN user_roles ur ON u.id = ur.user_id
			JOIN roles r ON ur.role_id = r.id
			WHERE r.name = 'Администратор' AND u.status = 'active' AND u.id != $1
		`, userID).Scan(&adminCount)
		if err != nil {
			return fmt.Errorf("failed to check admin count: %w", err)
		}
		if adminCount == 0 {
			return errors.New("cannot block the last active administrator")
		}
	}

	_, err = s.db.Pool.Exec(ctx, `
		UPDATE users SET status = 'blocked', updated_at = NOW() WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	// Отзываем все сессии
	_, err = s.db.Pool.Exec(ctx, `
		UPDATE sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}

	s.audit.Log(ctx, actorID, "user.block", "user", userID, map[string]interface{}{
		"login": user.Login,
	}, "")

	return nil
}

// UnblockUser разблокирует пользователя
func (s *AdminService) UnblockUser(ctx context.Context, actorID, userID uuid.UUID) error {
	if !s.perms.Can(ctx, actorID, "user.manage", nil) {
		return errors.New("insufficient permissions")
	}

	user, err := s.userSvc.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	_, err = s.db.Pool.Exec(ctx, `
		UPDATE users SET status = 'active', updated_at = NOW() WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}

	s.audit.Log(ctx, actorID, "user.unblock", "user", userID, map[string]interface{}{
		"login": user.Login,
	}, "")

	return nil
}

type SessionInfo struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	DeviceInfo *string    `json:"device_info,omitempty"`
	IP         *string    `json:"ip,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt time.Time  `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

// GetUserSessions возвращает список сессий пользователя
func (s *AdminService) GetUserSessions(ctx context.Context, actorID, userID uuid.UUID) ([]SessionInfo, error) {
	if !s.perms.Can(ctx, actorID, "user.manage", nil) {
		return nil, errors.New("insufficient permissions")
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, user_id, device_info, ip, created_at, last_used_at, expires_at, revoked_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY last_used_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []SessionInfo
	for rows.Next() {
		var session SessionInfo
		err := rows.Scan(
			&session.ID, &session.UserID, &session.DeviceInfo, &session.IP,
			&session.CreatedAt, &session.LastUsedAt, &session.ExpiresAt, &session.RevokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// RevokeSession отзывает конкретную сессию
func (s *AdminService) RevokeSession(ctx context.Context, actorID, userID, sessionID uuid.UUID) error {
	if !s.perms.Can(ctx, actorID, "user.manage", nil) {
		return errors.New("insufficient permissions")
	}

	result, err := s.db.Pool.Exec(ctx, `
		UPDATE sessions SET revoked_at = NOW()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("session not found or already revoked")
	}

	s.audit.Log(ctx, actorID, "session.revoke", "session", sessionID, map[string]interface{}{
		"user_id": userID,
	}, "")

	return nil
}

// RevokeAllSessions отзывает все сессии пользователя
func (s *AdminService) RevokeAllSessions(ctx context.Context, actorID, userID uuid.UUID) error {
	if !s.perms.Can(ctx, actorID, "user.manage", nil) {
		return errors.New("insufficient permissions")
	}

	result, err := s.db.Pool.Exec(ctx, `
		UPDATE sessions SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}

	s.audit.Log(ctx, actorID, "session.revoke_all", "user", userID, map[string]interface{}{
		"count": result.RowsAffected(),
	}, "")

	return nil
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