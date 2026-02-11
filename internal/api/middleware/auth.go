package middleware

import (
	"net/http"
	"strings"
)

// AuthMiddleware validates Bearer token (mock implementation for documentation)
// In production, this should validate actual JWT tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// For now, just check if Bearer token exists
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// TODO: Validate actual JWT token here
		// token := strings.TrimPrefix(authHeader, "Bearer ")
		// if !validateJWT(token) {
		//     http.Error(w, `{"error":"Invalid token"}`, http.StatusUnauthorized)
		//     return
		// }

		next.ServeHTTP(w, r)
	})
}
