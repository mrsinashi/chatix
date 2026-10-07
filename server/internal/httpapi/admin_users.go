package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chatix/internal/users"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AdminUsersHandler struct {
	adminSvc *users.AdminService
}

func NewAdminUsersHandler(adminSvc *users.AdminService) *AdminUsersHandler {
	return &AdminUsersHandler{adminSvc: adminSvc}
}

func (h *AdminUsersHandler) RegisterRoutes(r chi.Router) {
	r.Route("/admin/users", func(r chi.Router) {
		r.Get("/", h.ListUsers)
		r.Post("/", h.CreateUser)
		r.Route("/{userID}", func(r chi.Router) {
			r.Patch("/", h.UpdateUser)
			r.Post("/set-password", h.SetPassword)
			r.Post("/block", h.BlockUser)
			r.Post("/unblock", h.UnblockUser)
			r.Get("/sessions", h.GetUserSessions)
			r.Delete("/sessions/{sessionID}", h.RevokeSession)
			r.Delete("/sessions", h.RevokeAllSessions)
		})
	})
}

func (h *AdminUsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	actorID := GetActorID(r.Context())
	if actorID == uuid.Nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
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

	result, err := h.adminSvc.ListUsers(r.Context(), actorID, query, kindPtr, statusPtr, limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_users_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *AdminUsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	actorID := GetActorID(r.Context())
	if actorID == uuid.Nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	var req users.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Некорректный формат запроса")
		return
	}

	user, err := h.adminSvc.CreateUser(r.Context(), actorID, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "create_user_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, user)
}

func (h *AdminUsersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	actorID := GetActorID(r.Context())
	if actorID == uuid.Nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Некорректный ID пользователя")
		return
	}

	var req users.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Некорректный формат запроса")
		return
	}

	user, err := h.adminSvc.UpdateUser(r.Context(), actorID, userID, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "update_user_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

func (h *AdminUsersHandler) SetPassword(w http.ResponseWriter, r *http.Request) {
	actorID := GetActorID(r.Context())
	if actorID == uuid.Nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Некорректный ID пользователя")
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Некорректный формат запроса")
		return
	}

	if req.Password == "" {
		WriteError(w, http.StatusBadRequest, "password_required", "Пароль обязателен")
		return
	}

	err = h.adminSvc.SetPassword(r.Context(), actorID, userID, req.Password)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "set_password_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminUsersHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	actorID := GetActorID(r.Context())
	if actorID == uuid.Nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Некорректный ID пользователя")
		return
	}

	err = h.adminSvc.BlockUser(r.Context(), actorID, userID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "block_user_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminUsersHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	actorID := GetActorID(r.Context())
	if actorID == uuid.Nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Некорректный ID пользователя")
		return
	}

	err = h.adminSvc.UnblockUser(r.Context(), actorID, userID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "unblock_user_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminUsersHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	actorID := GetActorID(r.Context())
	if actorID == uuid.Nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Некорректный ID пользователя")
		return
	}

	sessions, err := h.adminSvc.GetUserSessions(r.Context(), actorID, userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "get_sessions_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{"sessions": sessions})
}

func (h *AdminUsersHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	actorID := GetActorID(r.Context())
	if actorID == uuid.Nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Некорректный ID пользователя")
		return
	}

	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_session_id", "Некорректный ID сессии")
		return
	}

	err = h.adminSvc.RevokeSession(r.Context(), actorID, userID, sessionID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "revoke_session_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminUsersHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	actorID := GetActorID(r.Context())
	if actorID == uuid.Nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Некорректный ID пользователя")
		return
	}

	err = h.adminSvc.RevokeAllSessions(r.Context(), actorID, userID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "revoke_sessions_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}