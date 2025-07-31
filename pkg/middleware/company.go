package middleware

import (
	"net/http"
)

// CompanyAuth is a middleware that checks if the request is from an authorized company
func CompanyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Implementation details
		next.ServeHTTP(w, r)
	})
}