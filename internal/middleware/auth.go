package middleware

import (
	"Confetki/pkg/auth"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"
const RoleKey contextKey = "role"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid token", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[7:]
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := r.Context().Value(RoleKey).(string)
		if !ok || userRole != "admin" {
			http.Error(w, "access denied: admin only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
