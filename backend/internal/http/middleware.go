package http

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/auth"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type contextKey string

const claimsKey contextKey = "claims"

type limiterEntry struct {
	tokens   int
	lastSeen time.Time
}

type ipRateLimiter struct {
	mu      sync.Mutex
	limit   int
	burst   int
	window  time.Duration
	entries map[string]limiterEntry
}

func newIPRateLimiter(limitPerMinute int, burst int) *ipRateLimiter {
	if limitPerMinute <= 0 {
		limitPerMinute = 240
	}
	if burst <= 0 {
		burst = limitPerMinute / 4
	}
	if burst <= 0 {
		burst = 20
	}
	return &ipRateLimiter{
		limit:   limitPerMinute,
		burst:   burst,
		window:  time.Minute,
		entries: map[string]limiterEntry{},
	}
}

func (l *ipRateLimiter) allow(key string) (allowed bool, remaining int) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) > 10000 {
		for k, entry := range l.entries {
			if now.Sub(entry.lastSeen) > 10*time.Minute {
				delete(l.entries, k)
			}
		}
	}

	entry := l.entries[key]
	if entry.lastSeen.IsZero() || now.Sub(entry.lastSeen) >= l.window {
		entry.tokens = l.limit + l.burst
	}
	if entry.tokens <= 0 {
		entry.lastSeen = now
		l.entries[key] = entry
		return false, 0
	}
	entry.tokens--
	entry.lastSeen = now
	l.entries[key] = entry
	return true, entry.tokens
}

func (a *App) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip internal heartbeat and health probes from throttling.
		if r.URL.Path == "/health" || r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		ip := ""
		if a.cfg.TrustProxyHeaders {
			ip = strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
			if ip != "" {
				parts := strings.Split(ip, ",")
				ip = strings.TrimSpace(parts[0])
			}
		}
		if ip == "" {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil && host != "" {
				ip = host
			} else {
				ip = r.RemoteAddr
			}
		}
		limiter := a.rateLimiter
		limit := a.cfg.RateLimitPerMinute
		if strings.HasPrefix(r.URL.Path, "/auth/") {
			limiter = a.authLimiter
			limit = 30
		}
		allowed, remaining := limiter.allow(ip)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if !allowed {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" {
			if _, ok := a.allowedOrigins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			} else if r.Method == http.MethodOptions {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "origin not allowed"})
				return
			}
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Tenant-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

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
	return a.authenticateWithPurpose(auth.TokenPurposeAppAuth, next)
}

func (a *App) authenticateOAuthAccess(next http.Handler) http.Handler {
	return a.authenticateWithPurpose(auth.TokenPurposeOAuthAccess, next)
}

func (a *App) authenticateWithPurpose(expectedPurpose string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := tokenFromRequest(r)
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
			return
		}
		claims, err := a.jwt.ParseForPurpose(token, expectedPurpose)
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

func tokenFromRequest(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token != "" {
			return token
		}
	}
	if c, err := r.Cookie("mcp_access_token"); err == nil {
		if v := strings.TrimSpace(c.Value); v != "" {
			return v
		}
	}
	return ""
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
