package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"chatix/internal/service"
)

type AuthHandlers struct {
	authService *service.AuthService
	userService *service.UserService
}

func NewAuthHandlers(authService *service.AuthService, userService *service.UserService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		userService: userService,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

type UserInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email,omitempty"`
}

type MeResponse struct {
	User        UserInfo `json:"user"`
	Permissions []string `json:"permissions"`
}

// Login обрабатывает POST /api/v1/auth/login
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "логин и пароль обязательны", http.StatusBadRequest)
		return
	}

	session, user, err := h.authService.Login(r.Context(), req.Username, req.Password, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			http.Error(w, "неверный логин или пароль", http.StatusUnauthorized)
		case service.ErrUserLocked:
			http.Error(w, "пользователь заблокирован", http.StatusForbidden)
		case service.ErrTooManyAttempts:
			http.Error(w, "слишком много попыток входа, попробуйте позже", http.StatusTooManyRequests)
		default:
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	resp := LoginResponse{
		Token:     session.TokenHash,
		ExpiresAt: session.ExpiresAt,
		User: UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
		},
	}
	if user.Email != nil {
		resp.User.Email = *user.Email
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout обрабатывает POST /api/v1/auth/logout
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		http.Error(w, "токен не указан", http.StatusUnauthorized)
		return
	}

	if err := h.authService.Logout(r.Context(), token); err != nil {
		http.Error(w, "сессия не найдена", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Me обрабатывает GET /api/v1/auth/me
func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		http.Error(w, "токен не указан", http.StatusUnauthorized)
		return
	}

	user, err := h.authService.ValidateSession(r.Context(), token)
	if err != nil {
		http.Error(w, "сессия недействительна", http.StatusUnauthorized)
		return
	}

	permissions, err := h.userService.GetUserPermissions(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	resp := MeResponse{
		User: UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
		},
		Permissions: permissions,
	}
	if user.Email != nil {
		resp.User.Email = *user.Email
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}