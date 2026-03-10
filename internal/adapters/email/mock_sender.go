package email

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

type MockSender struct {
	failRate float64
}

func NewMockSender(failRate float64) *MockSender {
	rand.Seed(time.Now().UnixNano())

	return &MockSender{
		failRate: failRate,
	}
}

func (m *MockSender) SendTeamInvitation(ctx context.Context, email string, teamName string) error {

	time.Sleep(100 * time.Millisecond)

	if rand.Float64() < m.failRate {
		return errors.New("mock email service unavailable")
	}

	return nil
}