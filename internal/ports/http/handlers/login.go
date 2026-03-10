package handlers

import (
	"TaskControlService/internal/app/auth"
	"TaskControlService/internal/ports/http/messages"
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		messages.BadRequest(w, "invalid_request", "invalid request body")

		h.logger.Error("failed to decode login request", zap.Error(err), zap.String("code", "invalid_request"))

		return
	}

	token, err := h.authService.Login(
		r.Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			messages.Unauthorized(w, "invalid_credentials", "indalid email or password")

			h.logger.Info("invalid login attempt", zap.String("email", req.Email), zap.String("code", "invalid_credentials"))
			
			return
		}

		messages.InternalError(w)

		h.logger.Error("failed to login user", zap.Error(err), zap.String("code", "internal_error"))
		
		return
	}

	resp := map[string]string{
		"token": token,
	}
	
	messages.WriteJSON(w, http.StatusOK, resp)
}
