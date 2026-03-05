package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/yourorg/mcp-marketplace/backend/internal/auth"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type contextKey string

const claimsKey contextKey = "claims"

func (a *App) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func (a *App) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := a.jwt.Parse(token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		if requestTenant := r.Header.Get("X-Tenant-ID"); requestTenant != "" && requestTenant != claims.TenantID {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "tenant mismatch"})
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) requireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := getClaims(r.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			if _, ok := allowed[string(claims.Role)]; !ok {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient role"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func getClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*auth.Claims)
	return claims, ok
}

func hasScope(scopes []string, want string) bool {
	for _, scope := range scopes {
		if scope == want {
			return true
		}
	}
	return false
}

func filterGranted(required []string, granted []string) []string {
	out := make([]string, 0)
	for _, req := range required {
		if hasScope(granted, req) {
			out = append(out, req)
		}
	}
	return out
}

func roleForUser(user models.User) string {
	return string(user.Role)
}
