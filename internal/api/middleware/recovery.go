package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/proyaai/instantgate/internal/api/handlers"
)

func Recovery() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					println("PANIC:", err)
					println(string(stack))

					handlers.SendError(w, r, http.StatusInternalServerError, "Internal server error", nil)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
