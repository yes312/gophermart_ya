package transport

import (
	"context"
	"net/http"
)

type key string

const userIDKey key = "userID"

func (h *handlersData) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			h.logger.Errorf("остутствует токен авторизации")
			http.Error(w, "остутствует токен авторизации", http.StatusUnauthorized)
			return
		}
		user, err := h.AuthToken.GetUserID(authHeader)
		if err != nil {
			h.logger.Errorf("ошибка проверки токена: %w", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// #ВОПРОСМЕНТОРУ  получаем юзера, и передаем его дальше через контекст. Не знаю хороший ли способ. Возможно есть более предпочтительный?
		ctx := context.WithValue(r.Context(), userIDKey, user)
		w.Header().Set("Content-Type", ApplicationJSON)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
