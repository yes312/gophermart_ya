package transport

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gophermart/internal/models"
	"gophermart/internal/services"
	"net/http"
)

func (h *handlersData) Registration(w http.ResponseWriter, r *http.Request) {

	var data authData

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if data.Login == "" || data.Password == "" {
		http.Error(w, "Пароль или логин не кор ректны", http.StatusBadRequest)
		return
	}

	_, err = h.storage.WithRetry(h.ctx, h.storage.GetUser(h.ctx, data.Login))

	switch {
	case errors.Is(err, sql.ErrNoRows):

		hash := services.GetHash(data.Login, data.Password)

		_, err = h.storage.WithRetry(h.ctx, h.storage.AddUser(h.ctx, data.Login, hash))
		if err != nil {
			h.logger.Errorf("Ошибка добавления пользователя %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jwtString, err := h.AuthToken.BuildJWTString(data.Login)
		if err != nil {
			h.logger.Errorf("ошибка создания токена:  %w")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Authorization", jwtString)
		setResponseHeaders(w, ApplicationJSON, http.StatusOK)

	case err != nil:

		h.logger.Errorf("Ошибка при проверке существования пользователя: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	default:

		h.logger.Errorf("логин уже занят %s", data.Login)
		setResponseHeaders(w, ApplicationJSON, http.StatusConflict)
	}

}

func setResponseHeaders(w http.ResponseWriter, contentType string, statusCode int) {

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)

}

func (h *handlersData) Login(w http.ResponseWriter, r *http.Request) {

	var data authData

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if data.Login == "" || data.Password == "" {
		http.Error(w, "Пароль или логин не корректны", http.StatusBadRequest)
		return
	}

	userInterface, err := h.storage.WithRetry(h.ctx, h.storage.GetUser(h.ctx, data.Login))
	user, ok := userInterface.(models.User)

	switch {
	case errors.Is(err, sql.ErrNoRows):

		h.logger.Errorf("Пользователя %w не существует", data.Login)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return

	case err != nil || !ok:
		h.logger.Errorf("Ошибка при получении пользователя %w", data.Login)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:

		hash := services.GetHash(data.Login, data.Password)
		if hash == user.Hash {

			h.logger.Infof("пользователь %s идентифицирован", data.Login)

			jwtString, err := h.AuthToken.BuildJWTString(data.Login)
			if err != nil {
				h.logger.Error("ошибка создания токена")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Add("Authorization", jwtString)
			setResponseHeaders(w, ApplicationJSON, http.StatusOK)

		} else {
			h.logger.Infof("неверный пароль для пользователя %s", data.Login)
			http.Error(w, fmt.Sprint(data.Login, "-неверный пароль"), http.StatusUnauthorized)
		}

	}

}
