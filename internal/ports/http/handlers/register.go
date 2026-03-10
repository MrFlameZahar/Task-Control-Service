package handlers

import (
	"TaskControlService/internal/app/auth"
	"TaskControlService/internal/ports/http/messages"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"go.uber.org/zap"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {

	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		messages.BadRequest(w, "invalid_request", "invalid request body")

		h.logger.Error("failed to decode register request",
			zap.Error(err),
			zap.String("code", "invalid_request"),
		)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" {
		messages.BadRequest(w, "invalid_email", "email is required")

		h.logger.Info("register validation failed",
			zap.String("reason", "empty_email"),
		)
		return
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		messages.BadRequest(w, "invalid_email", "invalid email format")

		h.logger.Info("register validation failed",
			zap.String("email", req.Email),
			zap.String("reason", "invalid_email_format"),
		)
		return
	}

	if req.Password == "" {
		messages.BadRequest(w, "invalid_password", "password is required")

		h.logger.Info("register validation failed",
			zap.String("email", req.Email),
			zap.String("reason", "empty_password"),
		)
		return
	}

	err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {

		if errors.Is(err, auth.ErrUserExists) {
			messages.BadRequest(w, "user_exists", "user already exists")

			h.logger.Info("user already exists",
				zap.Error(err),
				zap.String("email", req.Email),
				zap.String("code", "user_exists"),
			)
			return
		}

		messages.InternalError(w)

		h.logger.Error("failed to register user",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.String("code", "internal_error"),
		)

		return
	}

	messages.WriteJSON(w, http.StatusCreated, map[string]string{
		"status": "ok",
	})
}
