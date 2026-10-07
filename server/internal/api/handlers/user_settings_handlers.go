package handlers

import (
	"encoding/json"
	"net/http"

	"chatix/internal/api/middleware"
	"chatix/internal/service"

	"github.com/go-chi/chi/v5"
)

type UserSettingsHandlers struct {
	settingService *service.SettingServiceV2
}

func NewUserSettingsHandlers(settingService *service.SettingServiceV2) *UserSettingsHandlers {
	return &UserSettingsHandlers{settingService: settingService}
}

// GetUserSettings обрабатывает GET /api/v1/settings
// Возвращает все настройки для текущего пользователя с резолвом.
func (h *UserSettingsHandlers) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		http.Error(w, "пользователь не найден", http.StatusUnauthorized)
		return
	}

	settings, err := h.settingService.GetUserSettings(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// GetUserSetting обрабатывает GET /api/v1/settings/{key}
func (h *UserSettingsHandlers) GetUserSetting(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		http.Error(w, "пользователь не найден", http.StatusUnauthorized)
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "ключ настройки не указан", http.StatusBadRequest)
		return
	}

	settings, err := h.settingService.GetUserSettings(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	value, ok := settings[key]
	if !ok {
		http.Error(w, "настройка не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]json.RawMessage{key: value})
}

// SetUserSetting обрабатывает PUT /api/v1/settings/{key}
func (h *UserSettingsHandlers) SetUserSetting(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		http.Error(w, "пользователь не найден", http.StatusUnauthorized)
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "ключ настройки не указан", http.StatusBadRequest)
		return
	}

	var value json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		http.Error(w, "неверный формат значения", http.StatusBadRequest)
		return
	}

	if err := h.settingService.SetUserSetting(r.Context(), user.ID, key, value); err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteUserSetting обрабатывает DELETE /api/v1/settings/{key}
func (h *UserSettingsHandlers) DeleteUserSetting(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		http.Error(w, "пользователь не найден", http.StatusUnauthorized)
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "ключ настройки не указан", http.StatusBadRequest)
		return
	}

	if err := h.settingService.DeleteUserSetting(r.Context(), user.ID, key); err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}