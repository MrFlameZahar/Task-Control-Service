package handlers

import (
	"TaskControlService/internal/app/team"
	teamDomain "TaskControlService/internal/domain/team"
	"TaskControlService/internal/ports/http/messages"
	"TaskControlService/internal/ports/http/middleware"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type InviteMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type TeamMemberResponse struct {
	UserID string `json:"user_id"`
	TeamID string `json:"team_id"`
	Role   string `json:"role"`
}

type TeamResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedBy   string `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Error("invalid user id in context", zap.String("code", "missing_user_id"))
		messages.Unauthorized(w, "invalid_token", "missing user id in token")
		return
	}

	var req CreateTeamRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("failed to decode create team request", zap.Error(err), zap.String("code", "invalid_request"))
		messages.BadRequest(w, "invalid_request", "invalid request body")
		return
	}

	err = h.teamService.CreateTeam(r.Context(), req.Name, userID)

	if err != nil {
		if errors.Is(err, team.ErrTeamExists) {
			h.logger.Info("team already exists", zap.Error(err), zap.String("name", req.Name), zap.String("code", "team_exists"))
			messages.BadRequest(w, "team_exists", "team already exists")
			return
		}
		h.logger.Error("failed to create team", zap.Error(err), zap.String("name", req.Name), zap.Any("userID", userID), zap.String("code", "internal_error"))
		messages.InternalError(w)
		return
	}

	messages.WriteJSON(w, http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) GetTeams(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Error("invalid user id in context", zap.String("code", "missing_user_id"))
		messages.Unauthorized(w, "unauthorized", "user not authenticated")
		return
	}

	teams, err := h.teamService.GetUserTeams(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user teams", zap.Error(err), zap.Any("userID", userID), zap.String("code", "internal_error"))
		messages.InternalError(w)
		return
	}

	resp := make([]TeamResponse, 0, len(teams))
	for _, t := range teams {
		resp = append(resp, TeamResponse{
			ID:        t.ID.String(),
			Name:      t.Name,
			CreatedBy:   t.CreatedBy.String(),
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		})
	}

	messages.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"teams": resp,
	})
}

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Warn("user not authenticated", zap.String("code", "unauthorized"))
		messages.Unauthorized(w, "unauthorized", "user not authenticated")
		return
	}

	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		h.logger.Warn("invalid team id", zap.String("team_id", teamIDStr), zap.String("code", "invalid_team_id"))
		messages.BadRequest(w, "invalid_team_id", "invalid team id")
		return
	}

	var req InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body", zap.Error(err), zap.String("code", "invalid_request"))
		messages.BadRequest(w, "invalid_request", "invalid request body")
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.logger.Warn("invalid target user id", zap.String("user_id", req.UserID), zap.String("code", "invalid_user_id"))
		messages.BadRequest(w, "invalid_user_id", "invalid user id")
		return
	}

	err = h.teamService.InviteMember(
		r.Context(),
		actorID,
		teamID,
		targetUserID,
		teamDomain.Role(req.Role),
	)
	if err != nil {
		switch {
		case errors.Is(err, team.ErrInvalidInviteRole):
			h.logger.Warn("invalid role for team member", zap.String("role", req.Role), zap.String("code", "invalid_role"))
			messages.BadRequest(w, "invalid_role", "role must be admin or member")
			return

		case errors.Is(err, team.ErrTeamNotFound):
			h.logger.Warn("team not found", zap.String("team_id", teamIDStr), zap.String("code", "team_not_found"))
			messages.WriteError(w, http.StatusNotFound, messages.Error{
				Code:    "team_not_found",
				Message: "team not found",
			})
			return

		case errors.Is(err, team.ErrUserNotFound):
			h.logger.Warn("target user not found", zap.String("user_id", req.UserID), zap.String("code", "user_not_found"))
			messages.WriteError(w, http.StatusNotFound, messages.Error{
				Code:    "user_not_found",
				Message: "user not found",
			})
			return

		case errors.Is(err, team.ErrForbidden):
			h.logger.Warn("insufficient permissions for invite", zap.Any("actor_id", actorID), zap.String("team_id", teamIDStr), zap.String("code", "forbidden"))
			messages.WriteError(w, http.StatusForbidden, messages.Error{
				Code:    "forbidden",
				Message: "insufficient permissions",
			})
			return

		case errors.Is(err, team.ErrAlreadyTeamMember):
			h.logger.Info("user is already a team member", zap.Any("target_user_id", targetUserID), zap.String("team_id", teamIDStr), zap.String("code", "already_team_member"))
			messages.WriteError(w, http.StatusConflict, messages.Error{
				Code:    "already_team_member",
				Message: "user is already a team member",
			})
			return

		default:
			h.logger.Error("failed to invite member", zap.Error(err))
			messages.InternalError(w)
			return
		}
	}

	messages.WriteJSON(w, http.StatusCreated, map[string]string{
		"status": "created",
	})
}

func (h *Handler) GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Warn("user not authenticated", zap.String("code", "unauthorized"))
		messages.Unauthorized(w, "unauthorized", "user not authenticated")
		return
	}

	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		h.logger.Warn("invalid team id", zap.String("team_id", teamIDStr), zap.String("code", "invalid_team_id"))
		messages.BadRequest(w, "invalid_team_id", "invalid team id")
		return
	}

	members, err := h.teamService.GetTeamMembers(r.Context(), actorID, teamID)
	if err != nil {
		switch {
		case errors.Is(err, team.ErrTeamNotFound):
			messages.WriteError(w, http.StatusNotFound, messages.Error{
				Code:    "team_not_found",
				Message: "team not found",
			})
			return

		case errors.Is(err, team.ErrForbidden):
			messages.WriteError(w, http.StatusForbidden, messages.Error{
				Code:    "forbidden",
				Message: "you are not a member of this team",
			})
			return

		default:
			h.logger.Error("failed to get team members", zap.Error(err))
			messages.InternalError(w)
			return
		}
	}

	resp := make([]TeamMemberResponse, 0, len(members))
	for _, m := range members {
		resp = append(resp, TeamMemberResponse{
			UserID: m.UserID.String(),
			TeamID: m.TeamID.String(),
			Role:   string(m.Role),
		})
	}

	messages.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"members": resp,
	})
}
