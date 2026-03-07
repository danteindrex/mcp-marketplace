package http

import (
	"net/http"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func (a *App) ensureServerTenantAccess(w http.ResponseWriter, r *http.Request, server models.Server) bool {
	claims, ok := getClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return false
	}
	if claims.Role == models.RoleAdmin {
		return true
	}
	if claims.TenantID != server.TenantID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return false
	}
	return true
}
