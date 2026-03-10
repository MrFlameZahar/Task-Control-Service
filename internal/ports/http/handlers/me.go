package handlers

import (
	"TaskControlService/internal/app/auth"
	"TaskControlService/internal/domain/user"
	"TaskControlService/internal/ports/http/messages"
	"TaskControlService/internal/ports/http/middleware"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Error("invalid user id in context", zap.String("code", "missing_user_id"))

		messages.Unauthorized(w, "invalid_token", "missing user id in token")

		return
	}

	userResult, err := h.authService.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			messages.BadRequest(w, "user_not_found", "user not found")

			h.logger.Warn("failed to get user info", zap.Error(err), zap.String("code", "user_not_found"))

			return
		}
		messages.BadRequest(w, "user_not_found", "user not found")

		h.logger.Warn("failed to get user info", zap.Error(err), zap.String("code", "user_not_found"))

		return
	}

	messages.WriteJSON(w, http.StatusOK, map[string]user.User{
		"user": *userResult,
	})
}
