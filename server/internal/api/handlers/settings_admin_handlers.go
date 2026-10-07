package handlers

import (
	"encoding/json"
	"net/http"

	"chatix/internal/api/middleware"
	"chatix/internal/service"

	"github.com/go-chi/chi/v5"
)

type SettingsAdminHandlers struct {
	settingService *service.SettingServiceV2
}

func NewSettingsAdminHandlers(settingService *service.SettingServiceV2) *SettingsAdminHandlers {
	return &SettingsAdminHandlers{settingService: settingService}
}

// ListSettings обрабатывает GET /api/v1/admin/settings
func (h *SettingsAdminHandlers) ListSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingService.GetSystemSettings(r.Context())
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// GetSetting обрабатывает GET /api/v1/admin/settings/{key}
func (h *SettingsAdminHandlers) GetSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "ключ настройки не указан", http.StatusBadRequest)
		return
	}

	settings, err := h.settingService.GetSystemSettings(r.Context())
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

// SetSetting обрабатывает PUT /api/v1/admin/settings/{key}
func (h *SettingsAdminHandlers) SetSetting(w http.ResponseWriter, r *http.Request) {
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

	user := middleware.GetUser(r)
	var changedBy *string
	if user != nil {
		changedBy = &user.ID
	}

	if err := h.settingService.SetSystemSetting(r.Context(), key, value, changedBy); err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteSetting обрабатывает DELETE /api/v1/admin/settings/{key}
func (h *SettingsAdminHandlers) DeleteSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "ключ настройки не указан", http.StatusBadRequest)
		return
	}

	if err := h.settingService.DeleteSystemSetting(r.Context(), key); err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}