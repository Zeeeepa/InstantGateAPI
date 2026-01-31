package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/proyaai/instantgate/internal/api/handlers"
	"github.com/proyaai/instantgate/internal/security"
)

func TableAccessControl(ac *security.AccessControl) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tableName := chi.URLParam(r, "table")
			if tableName == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !ac.IsTableAllowed(tableName) {
				handlers.SendError(w, r, http.StatusForbidden, handlers.ErrForbidden, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth(requireAuth bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !requireAuth {
				next.ServeHTTP(w, r)
				return
			}

			_, ok := GetClaims(r)
			if !ok {
				handlers.SendError(w, r, http.StatusUnauthorized, handlers.ErrUnauthorized, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
