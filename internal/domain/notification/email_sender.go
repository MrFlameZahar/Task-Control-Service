package notification

import "context"

type EmailSender interface {
	SendTeamInvitation(ctx context.Context, email string, teamName string) error
}