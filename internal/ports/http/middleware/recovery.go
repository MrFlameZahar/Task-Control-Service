package middleware

import (
	"fmt"
	"net/http"
)

func Recoverer(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				fmt.Println("PANIC:", err)
				http.Error(w, "internal error:", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}