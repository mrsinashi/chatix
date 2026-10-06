package handlers

import (
	"encoding/json"
	"net/http"

	"chatix/internal/service"
)

type SettingHandlers struct {
	settingService *service.SettingService
}

func NewSettingHandlers(settingService *service.SettingService) *SettingHandlers {
	return &SettingHandlers{settingService: settingService}
}

// GetPublicSettings обрабатывает GET /api/v1/settings/system
// Публичный эндпоинт, не требует авторизации.
func (h *SettingHandlers) GetPublicSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingService.GetPublicSettings(r.Context())
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}