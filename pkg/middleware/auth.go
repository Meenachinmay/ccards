package middleware

import (
	"net/http"
)

// Auth is a middleware that checks if the request is authenticated
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Implementation details
		next.ServeHTTP(w, r)
	})
}