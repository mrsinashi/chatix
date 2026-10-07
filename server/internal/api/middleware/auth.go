package middleware

import (
	"context"
	"net/http"

	"chatix/internal/models"
	"chatix/internal/service"
)

type contextKey string

const UserContextKey contextKey = "user"

// Auth проверяет токен в заголовке и добавляет пользователя в контекст
func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, "токен не указан", http.StatusUnauthorized)
				return
			}

			user, err := authService.ValidateSession(r.Context(), token)
			if err != nil {
				http.Error(w, "сессия недействительна", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission проверяет, что пользователь имеет указанное право
func RequirePermission(userService *service.UserService, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r)
			if user == nil {
				http.Error(w, "пользователь не найден в контексте", http.StatusUnauthorized)
				return
			}

			has, err := userService.HasPermission(r.Context(), user.ID, permission)
			if err != nil {
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}

			if !has {
				http.Error(w, "недостаточно прав", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}