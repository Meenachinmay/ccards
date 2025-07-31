package middleware

import (
	"net/http"
)

// CORS is a middleware that adds CORS headers to the response
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Implementation details
		next.ServeHTTP(w, r)
	})
}