package middleware

import (
	"net/http"
	"sync"
	"time"

	"TaskControlService/internal/ports/http/messages"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	users    map[uuid.UUID]*userLimiter
	rate     rate.Limit
	burst    int
	lifetime time.Duration
}

func NewRateLimiter(requestsPerMinute int, burst int, lifetime time.Duration) *RateLimiter {
	return &RateLimiter{
		users:    make(map[uuid.UUID]*userLimiter),
		rate:     rate.Every(time.Minute / time.Duration(requestsPerMinute)),
		burst:    burst,
		lifetime: lifetime,
	}
}

func (rl *RateLimiter) getLimiter(userID uuid.UUID) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.users[userID]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.users[userID] = &userLimiter{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	entry.lastSeen = time.Now()
	return entry.limiter
}

func (rl *RateLimiter) CleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for userID, entry := range rl.users {
			if time.Since(entry.lastSeen) > rl.lifetime {
				delete(rl.users, userID)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		limiter := rl.getLimiter(userID)
		if !limiter.Allow() {
			messages.WriteError(w, http.StatusTooManyRequests, messages.Error{
				Code:    "rate_limit_exceeded",
				Message: "too many requests",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}