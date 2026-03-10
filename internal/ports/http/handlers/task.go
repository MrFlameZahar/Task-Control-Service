package handlers

import (
	"TaskControlService/internal/app/task"
	taskDomain "TaskControlService/internal/domain/task"
	"TaskControlService/internal/ports/http/messages"
	"TaskControlService/internal/ports/http/middleware"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type HistoryResponse struct {
	ID        string  `json:"id"`
	TaskID    string  `json:"task_id"`
	Field     string  `json:"field"`
	OldValue  *string `json:"old_value,omitempty"`
	NewValue  *string `json:"new_value,omitempty"`
	ChangedBy string  `json:"changed_by"`
	ChangedAt string  `json:"changed_at"`
}

type TaskResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	TeamID      string  `json:"team_id"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      *string `json:"status,omitempty"`
	TeamID      string  `json:"team_id"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	AssigneeID  *string `json:"assignee_id"`
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		messages.Unauthorized(w, "unauthorized", "user not authenticated")
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		messages.BadRequest(w, "invalid_request", "invalid request body")
		return
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		messages.BadRequest(w, "invalid_team_id", "invalid team id")
		return
	}

	var status *taskDomain.Status
	if req.Status != nil {
		parsedStatus := taskDomain.Status(*req.Status)
		status = &parsedStatus
	}

	var assigneeID *uuid.UUID
	if req.AssigneeID != nil {
		parsedAssigneeID, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			messages.BadRequest(w, "invalid_assignee_id", "invalid assignee id")
			return
		}
		assigneeID = &parsedAssigneeID
	}

	err = h.taskService.CreateTask(r.Context(), task.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		TeamID:      teamID,
		AssigneeID:  assigneeID,
		CreatedBy:   actorID,
	})
	if err != nil {
		switch {
		case errors.Is(err, task.ErrInvalidTitle):
			messages.BadRequest(w, "invalid_title", "title cannot be empty")
			return

		case errors.Is(err, task.ErrInvalidStatus):
			messages.BadRequest(w, "invalid_status", "invalid task status")
			return

		case errors.Is(err, task.ErrAssigneeNotInTeam):
			messages.BadRequest(w, "invalid_assignee", "assignee is not a member of the team")
			return

		case errors.Is(err, task.ErrTeamNotFound):
			messages.WriteError(w, http.StatusNotFound, messages.Error{
				Code:    "team_not_found",
				Message: "team not found",
			})
			return

		case errors.Is(err, task.ErrCreatorNotInTeam):
			messages.WriteError(w, http.StatusForbidden, messages.Error{
				Code:    "forbidden",
				Message: "you are not a member of this team",
			})
			return

		default:
			h.logger.Error("failed to create task", zap.Error(err))
			messages.InternalError(w)
			return
		}
	}

	messages.WriteJSON(w, http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		messages.Unauthorized(w, "unauthorized", "user not authenticated")
		return
	}

	query := r.URL.Query()

	teamIDStr := query.Get("team_id")
	if teamIDStr == "" {
		messages.BadRequest(w, "missing_team_id", "team_id is required")
		return
	}

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		messages.BadRequest(w, "invalid_team_id", "invalid team id")
		return
	}

	var status *taskDomain.Status
	if statusStr := query.Get("status"); statusStr != "" {
		parsedStatus := taskDomain.Status(statusStr)
		status = &parsedStatus
	}

	var assigneeID *uuid.UUID
	if assigneeIDStr := query.Get("assignee_id"); assigneeIDStr != "" {
		parsedAssigneeID, err := uuid.Parse(assigneeIDStr)
		if err != nil {
			messages.BadRequest(w, "invalid_assignee_id", "invalid assignee id")
			return
		}
		assigneeID = &parsedAssigneeID
	}

	limit := 20
	if limitStr := query.Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 0 {
			messages.BadRequest(w, "invalid_limit", "limit must be a non-negative integer")
			return
		}
		limit = parsedLimit
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			messages.BadRequest(w, "invalid_offset", "offset must be a non-negative integer")
			return
		}
		offset = parsedOffset
	}

	tasks, err := h.taskService.GetTasks(r.Context(), actorID, taskDomain.ListFilter{
		TeamID:     teamID,
		Status:     status,
		AssigneeID: assigneeID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		switch {
		case errors.Is(err, task.ErrInvalidStatus):
			messages.BadRequest(w, "invalid_status", "invalid task status")
			return

		case errors.Is(err, task.ErrTeamNotFound):
			messages.WriteError(w, http.StatusNotFound, messages.Error{
				Code:    "team_not_found",
				Message: "team not found",
			})
			return

		case errors.Is(err, task.ErrForbidden):
			messages.WriteError(w, http.StatusForbidden, messages.Error{
				Code:    "forbidden",
				Message: "you are not a member of this team",
			})
			return

		default:
			h.logger.Error("failed to get tasks", zap.Error(err))
			messages.InternalError(w)
			return
		}
	}

	resp := make([]TaskResponse, 0, len(tasks))
	for _, taskEntity := range tasks {
		resp = append(resp, toTaskResponse(taskEntity))
	}

	messages.WriteJSON(w, http.StatusOK, map[string]any{
		"tasks": resp,
	})
}

func (h *Handler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		messages.Unauthorized(w, "unauthorized", "user not authenticated")
		return
	}

	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		messages.BadRequest(w, "invalid_task_id", "invalid task id")
		return
	}

	history, err := h.taskService.GetTaskHistory(r.Context(), actorID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, task.ErrTaskNotFound):
			messages.WriteError(w, http.StatusNotFound, messages.Error{
				Code:    "task_not_found",
				Message: "task not found",
			})
			return

		case errors.Is(err, task.ErrForbidden):
			messages.WriteError(w, http.StatusForbidden, messages.Error{
				Code:    "forbidden",
				Message: "you are not a member of this team",
			})
			return

		default:
			h.logger.Error("failed to get task history", zap.Error(err))
			messages.InternalError(w)
			return
		}
	}

	resp := make([]HistoryResponse, 0, len(history))
	for _, entry := range history {
		resp = append(resp, HistoryResponse{
			ID:        entry.ID.String(),
			TaskID:    entry.TaskID.String(),
			Field:     entry.Field,
			OldValue:  entry.OldValue,
			NewValue:  entry.NewValue,
			ChangedBy: entry.ChangedBy.String(),
			ChangedAt: entry.ChangedAt.Format(time.RFC3339),
		})
	}

	messages.WriteJSON(w, http.StatusOK, map[string]any{
		"history": resp,
	})
}

func toTaskResponse(taskEntity taskDomain.Task) TaskResponse {
	var assigneeID *string
	if taskEntity.AssigneeID != nil {
		value := taskEntity.AssigneeID.String()
		assigneeID = &value
	}

	return TaskResponse{
		ID:          taskEntity.ID.String(),
		Title:       taskEntity.Title,
		Description: taskEntity.Description,
		Status:      string(taskEntity.Status),
		TeamID:      taskEntity.TeamID.String(),
		AssigneeID:  assigneeID,
		CreatedBy:   taskEntity.CreatedBy.String(),
		CreatedAt:   taskEntity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   taskEntity.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		messages.Unauthorized(w, "unauthorized", "user not authenticated")
		return
	}

	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		messages.BadRequest(w, "invalid_task_id", "invalid task id")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		messages.BadRequest(w, "invalid_request", "invalid request body")
		return
	}

	var input task.UpdateTaskInput
	input.TaskID = taskID
	input.ActorID = actorID

	if value, exists := raw["title"]; exists {
		var title string
		if err := json.Unmarshal(value, &title); err != nil {
			messages.BadRequest(w, "invalid_title", "invalid title")
			return
		}
		input.Title = &title
	}

	if value, exists := raw["description"]; exists {
		var description string
		if err := json.Unmarshal(value, &description); err != nil {
			messages.BadRequest(w, "invalid_description", "invalid description")
			return
		}
		input.Description = &description
	}

	if value, exists := raw["status"]; exists {
		var statusStr string
		if err := json.Unmarshal(value, &statusStr); err != nil {
			messages.BadRequest(w, "invalid_status", "invalid status")
			return
		}
		status := taskDomain.Status(statusStr)
		input.Status = &status
	}

	if value, exists := raw["assignee_id"]; exists {
		if string(value) == "null" {
			input.ClearAssign = true
		} else {
			var assigneeIDStr string
			if err := json.Unmarshal(value, &assigneeIDStr); err != nil {
				messages.BadRequest(w, "invalid_assignee_id", "invalid assignee id")
				return
			}

			assigneeID, err := uuid.Parse(assigneeIDStr)
			if err != nil {
				messages.BadRequest(w, "invalid_assignee_id", "invalid assignee id")
				return
			}

			input.AssigneeID = &assigneeID
		}
	}

	err = h.taskService.UpdateTask(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, task.ErrTaskNotFound):
			messages.WriteError(w, http.StatusNotFound, messages.Error{
				Code:    "task_not_found",
				Message: "task not found",
			})
			return

		case errors.Is(err, task.ErrForbidden):
			messages.WriteError(w, http.StatusForbidden, messages.Error{
				Code:    "forbidden",
				Message: "you are not a member of this team",
			})
			return

		case errors.Is(err, task.ErrInvalidTitle):
			messages.BadRequest(w, "invalid_title", "title cannot be empty")
			return

		case errors.Is(err, task.ErrInvalidStatus):
			messages.BadRequest(w, "invalid_status", "invalid task status")
			return

		case errors.Is(err, task.ErrAssigneeNotInTeam):
			messages.BadRequest(w, "invalid_assignee", "assignee is not a member of the team")
			return

		default:
			h.logger.Error("failed to update task", zap.Error(err))
			messages.InternalError(w)
			return
		}
	}

	messages.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}