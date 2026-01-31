package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/proyaai/instantgate/internal/api/handlers"
	"github.com/proyaai/instantgate/internal/security"
)

type contextKey string

const (
	ClaimsContextKey contextKey = "claims"
)

func JWTAuth(jwtManager *security.JWTManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				handlers.SendError(w, r, http.StatusUnauthorized, handlers.ErrUnauthorized, security.ErrNoToken)
				return
			}

			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				handlers.SendError(w, r, http.StatusUnauthorized, handlers.ErrUnauthorized, err)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalJWTAuth(jwtManager *security.JWTManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

func GetClaims(r *http.Request) (*security.Claims, bool) {
	claims, ok := r.Context().Value(ClaimsContextKey).(*security.Claims)
	return claims, ok
}

func RequireRole(roles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r)
			if !ok {
				handlers.SendError(w, r, http.StatusUnauthorized, handlers.ErrUnauthorized, nil)
				return
			}

			if !hasAnyRole(claims.Roles, roles) {
				handlers.SendError(w, r, http.StatusForbidden, handlers.ErrForbidden, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func hasAnyRole(userRoles, requiredRoles []string) bool {
	for _, required := range requiredRoles {
		for _, userRole := range userRoles {
			if userRole == required {
				return true
			}
		}
	}
	return false
}
