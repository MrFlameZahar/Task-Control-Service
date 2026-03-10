package middleware

import (
	"TaskControlService/internal/ports/http/messages"
	auth "TaskControlService/pkg/auth"
	"context"

	"net/http"
	"strings"

	"github.com/google/uuid"
)

func Auth(secret string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			header := r.Header.Get("Authorization")

			if header == "" {
				messages.Unauthorized(w, "missing_token", "missing authorization token")
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")

			userID, err := auth.ParseToken(token, secret)
			if err != nil {
				messages.WriteError(w, http.StatusUnauthorized, messages.Error{Code: "invalid token", Message: "invalid authorization token"})
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}
