package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func (a *App) listTenants(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": a.store.ListTenants()})
}

func (a *App) listSecurityEvents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": a.store.ListSecurityEvents()})
}

func (a *App) listAuditLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": a.store.ListAuditLogs()})
}

type grantEntitlementRequest struct {
	TenantID      string   `json:"tenantId"`
	UserID        string   `json:"userId"`
	ServerID      string   `json:"serverId"`
	AllowedScopes []string `json:"allowedScopes"`
	CloudAllowed  bool     `json:"cloudAllowed"`
	LocalAllowed  bool     `json:"localAllowed"`
}

func (a *App) adminGrantEntitlement(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	var req grantEntitlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TenantID == "" || req.UserID == "" || req.ServerID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if _, ok := a.store.GetTenantByID(req.TenantID); !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown tenant"})
		return
	}
	user, ok := a.store.GetUserByID(req.UserID)
	if !ok || user.TenantID != req.TenantID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown user for tenant"})
		return
	}
	server, ok := a.store.GetServerByID(req.ServerID)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown server"})
		return
	}
	if len(req.AllowedScopes) == 0 {
		req.AllowedScopes = server.RequiredScopes
	}
	filteredScopes := make([]string, 0, len(req.AllowedScopes))
	for _, scope := range req.AllowedScopes {
		scope = strings.TrimSpace(scope)
		if scope == "" || !hasScope(server.RequiredScopes, scope) {
			continue
		}
		filteredScopes = append(filteredScopes, scope)
	}
	if len(filteredScopes) == 0 && len(server.RequiredScopes) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "allowedScopes must include valid server scopes"})
		return
	}
	ent := a.store.GrantEntitlement(models.Entitlement{
		TenantID:      req.TenantID,
		UserID:        req.UserID,
		ServerID:      req.ServerID,
		AllowedScopes: filteredScopes,
		CloudAllowed:  req.CloudAllowed,
		LocalAllowed:  req.LocalAllowed,
	})
	a.store.AddAuditLog(models.AuditLog{TenantID: req.TenantID, ActorID: claims.UserID, Action: "entitlement.grant", TargetType: "entitlement", TargetID: ent.ID, Outcome: "success"})
	writeJSON(w, http.StatusCreated, ent)
}
