package team

import "errors"

var (
	ErrTeamExists         = errors.New("team exists")
	ErrTeamNotFound	   = errors.New("team not found")
	ErrAlreadyTeamMember	   = errors.New("user is already a member of the team")
	ErrUserNotFound	   = errors.New("user not found")
	ErrInvalidInviteRole   = errors.New("invalid invite role")
	ErrForbidden           = errors.New("forbidden")
)