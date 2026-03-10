package task

import "errors"

var (
	ErrTaskNotFound        = errors.New("task not found")
	ErrTeamNotFound        = errors.New("team not found")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidTitle        = errors.New("invalid title")
	ErrInvalidStatus       = errors.New("invalid status")
	ErrAssigneeNotInTeam   = errors.New("assignee is not a team member")
	ErrCreatorNotInTeam    = errors.New("creator is not a team member")
)