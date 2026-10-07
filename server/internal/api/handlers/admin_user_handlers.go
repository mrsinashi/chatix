package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chatix/internal/api/middleware"
	"chatix/internal/models"
	"chatix/internal/service"

	"github.com/go-chi/chi/v5"
)

type AdminUserHandlers struct {
	userService *service.UserService
}

func NewAdminUserHandlers(userService *service.UserService) *AdminUserHandlers {
	return &AdminUserHandlers{userService: userService}
}

func (h *AdminUserHandlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	query := r.URL.Query().Get("q")
	kind := r.URL.Query().Get("kind")
	status := r.URL.Query().Get("status")

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	var kindPtr *string
	if kind != "" {
		kindPtr = &kind
	}
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	users, total, err := h.userService.ListUsers(r.Context(), query, kindPtr, statusPtr, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_users_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"total": total,
	})
}

func (h *AdminUserHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	ipAddress := r.RemoteAddr

	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "Не указан ID пользователя")
		return
	}

	var req models.UserUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Некорректный формат запроса")
		return
	}

	updatedUser, err := h.userService.UpdateUser(r.Context(), userID, req, &user.ID, &ipAddress)
	if err != nil {
		writeError(w, http.StatusBadRequest, "update_user_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updatedUser)
}

func (h *AdminUserHandlers) SetPassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	ipAddress := r.RemoteAddr

	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "Не указан ID пользователя")
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Некорректный формат запроса")
		return
	}

	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password_required", "Пароль обязателен")
		return
	}

	err := h.userService.SetPassword(r.Context(), userID, req.Password, &user.ID, &ipAddress)
	if err != nil {
		writeError(w, http.StatusBadRequest, "set_password_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminUserHandlers) BlockUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	ipAddress := r.RemoteAddr

	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "Не указан ID пользователя")
		return
	}

	err := h.userService.BlockUser(r.Context(), userID, &user.ID, &ipAddress)
	if err != nil {
		writeError(w, http.StatusBadRequest, "block_user_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminUserHandlers) UnblockUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	ipAddress := r.RemoteAddr

	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "Не указан ID пользователя")
		return
	}

	err := h.userService.UnblockUser(r.Context(), userID, &user.ID, &ipAddress)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unblock_user_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminUserHandlers) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "Не указан ID пользователя")
		return
	}

	sessions, err := h.userService.GetUserSessions(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_sessions_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"sessions": sessions})
}

func (h *AdminUserHandlers) RevokeSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	ipAddress := r.RemoteAddr

	userID := chi.URLParam(r, "userID")
	sessionID := chi.URLParam(r, "sessionID")

	if userID == "" || sessionID == "" {
		writeError(w, http.StatusBadRequest, "invalid_params", "Не указаны параметры")
		return
	}

	err := h.userService.RevokeSession(r.Context(), userID, sessionID, &user.ID, &ipAddress)
	if err != nil {
		writeError(w, http.StatusBadRequest, "revoke_session_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminUserHandlers) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	ipAddress := r.RemoteAddr

	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "Не указан ID пользователя")
		return
	}

	err := h.userService.RevokeAllSessions(r.Context(), userID, &user.ID, &ipAddress)
	if err != nil {
		writeError(w, http.StatusBadRequest, "revoke_sessions_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}