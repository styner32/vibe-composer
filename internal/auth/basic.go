package auth

import (
	"context"
	"net/http"

	"github.com/vibe-composer/internal/config"
)

type contextKey string

const usernameKey contextKey = "username"

// UsernameFromContext extracts the authenticated username from the request context.
func UsernameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(usernameKey).(string); ok {
		return v
	}
	return ""
}

// BasicAuth returns middleware that validates Basic Auth credentials
// against the hardcoded allowed users list.
func BasicAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="Vibe Composer"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !cfg.IsAllowedUser(username) {
				http.Error(w, "Forbidden: user not invited", http.StatusForbidden)
				return
			}

			if cfg.AuthPassword != "" && password != cfg.AuthPassword {
				w.Header().Set("WWW-Authenticate", `Basic realm="Vibe Composer"`)
				http.Error(w, "Unauthorized: invalid password", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), usernameKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
