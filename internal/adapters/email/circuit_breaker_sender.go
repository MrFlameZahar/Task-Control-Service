package email

import (
	"context"
	"fmt"
	"time"

	"TaskControlService/internal/domain/notification"

	"github.com/sony/gobreaker"
)

type CircuitBreakerSender struct {
	sender  notification.EmailSender
	breaker *gobreaker.CircuitBreaker
}

func NewCircuitBreakerSender(sender notification.EmailSender) *CircuitBreakerSender {
	settings := gobreaker.Settings{
		Name:        "email_sender",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}

	return &CircuitBreakerSender{
		sender:  sender,
		breaker: gobreaker.NewCircuitBreaker(settings),
	}
}

func (c *CircuitBreakerSender) SendTeamInvitation(ctx context.Context, email string, teamName string) error {
	_, err := c.breaker.Execute(func() (interface{}, error) {
		return nil, c.sender.SendTeamInvitation(ctx, email, teamName)
	})
	if err != nil {
		return fmt.Errorf("email sender failed: %w", err)
	}

	return nil
}