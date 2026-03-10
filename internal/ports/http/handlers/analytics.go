package handlers

import (
	"TaskControlService/internal/ports/http/messages"
	"TaskControlService/internal/ports/http/middleware"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) GetTeamStats(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		messages.Unauthorized(w, "unauthorized", "user not authenticated")
		return
	}

	stats, err := h.analyticsService.GetTeamStats(r.Context())
	if err != nil {
		h.logger.Error("failed to get team stats", zap.Error(err))
		messages.InternalError(w)
		return
	}

	messages.WriteJSON(w, http.StatusOK, map[string]any{
		"data": stats,
	})
}

func (h *Handler) GetTopCreators(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		messages.Unauthorized(w, "unauthorized", "user not authenticated")
		return
	}

	data, err := h.analyticsService.GetTopCreators(r.Context())
	if err != nil {
		h.logger.Error("failed to get top creators", zap.Error(err))
		messages.InternalError(w)
		return
	}

	messages.WriteJSON(w, http.StatusOK, map[string]any{
		"data": data,
	})
}

func (h *Handler) GetIntegrityIssues(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		messages.Unauthorized(w, "unauthorized", "user not authenticated")
		return
	}

	data, err := h.analyticsService.GetIntegrityIssues(r.Context())
	if err != nil {
		h.logger.Error("failed to get integrity issues", zap.Error(err))
		messages.InternalError(w)
		return
	}

	messages.WriteJSON(w, http.StatusOK, map[string]any{
		"data": data,
	})
}