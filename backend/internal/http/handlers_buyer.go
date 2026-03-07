package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func entitlementForServer(entitlements []models.Entitlement, serverID string) (models.Entitlement, bool) {
	for _, ent := range entitlements {
		if ent.ServerID == serverID && ent.Status == "active" {
			return ent, true
		}
	}
	return models.Entitlement{}, false
}

func (a *App) listBuyerConnections(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	connections := a.store.ListConnections(claims.TenantID, claims.UserID)
	for i := range connections {
		if connections[i].ServerName == "" && connections[i].ServerID != "" {
			if srv, ok := a.store.GetServerByID(connections[i].ServerID); ok {
				connections[i].ServerName = srv.Name
				connections[i].ServerSlug = srv.Slug
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": connections, "count": len(connections)})
}

type createConnectionRequest struct {
	Client        string   `json:"client"`
	ServerID      string   `json:"serverId"`
	Resource      string   `json:"resource"`
	GrantedScopes []string `json:"grantedScopes"`
}

func (a *App) createBuyerConnection(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	var req createConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Client == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	hub := a.ensureHubProfile(claims.TenantID, claims.UserID)
	resource := req.Resource
	serverID := ""
	serverSlug := ""
	serverName := ""
	grantedScopes := req.GrantedScopes
	if strings.TrimSpace(req.ServerID) != "" {
		server, ok := a.store.GetServerByID(strings.TrimSpace(req.ServerID))
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
			return
		}
		if claims.Role != models.RoleAdmin && claims.TenantID != server.TenantID && server.Status != "published" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
			return
		}
		allowedScopes := server.RequiredScopes
		if claims.Role != models.RoleAdmin && claims.TenantID != server.TenantID {
			entitlements := a.store.ListEntitlements(claims.TenantID, claims.UserID)
			entitlement, hasEntitlement := entitlementForServer(entitlements, server.ID)
			if !hasEntitlement {
				if server.Status == "published" && server.PricingType == "free" {
					entitlement = a.store.GrantEntitlement(models.Entitlement{
						TenantID:      claims.TenantID,
						UserID:        claims.UserID,
						ServerID:      server.ID,
						AllowedScopes: server.RequiredScopes,
						CloudAllowed:  true,
						LocalAllowed:  true,
					})
					hasEntitlement = true
				}
			}
			if !hasEntitlement {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "entitlement required"})
				return
			}
			if len(entitlement.AllowedScopes) > 0 {
				allowedScopes = entitlement.AllowedScopes
			}
		}
		serverID = server.ID
		serverSlug = server.Slug
		serverName = server.Name
		if len(grantedScopes) == 0 {
			grantedScopes = allowedScopes
		} else {
			grantedScopes = filterGranted(allowedScopes, grantedScopes)
		}
	}
	if strings.TrimSpace(resource) == "" {
		resource = hub.HubURL
	}
	conn := a.store.UpsertConnection(models.Connection{
		TenantID:       claims.TenantID,
		UserID:         claims.UserID,
		ServerID:       serverID,
		ServerSlug:     serverSlug,
		ServerName:     serverName,
		HubID:          hub.ID,
		Client:         normalizeClientName(req.Client),
		Status:         "active",
		Resource:       resource,
		GrantedScopes:  grantedScopes,
		TokenExpiresAt: time.Now().UTC().Add(8 * time.Hour),
		CatalogVersion: hub.CatalogVersion,
	})
	a.store.AddAuditLog(models.AuditLog{TenantID: claims.TenantID, ActorID: claims.UserID, Action: "connection.create", TargetType: "connection", TargetID: conn.ID, Outcome: "success", Metadata: map[string]interface{}{"client": req.Client, "serverId": serverID}})
	writeJSON(w, http.StatusCreated, conn)
}

func (a *App) listBuyerEntitlements(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	entitlements := a.store.ListEntitlements(claims.TenantID, claims.UserID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": entitlements, "count": len(entitlements)})
}

func (a *App) getBuyerHub(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	hub := a.ensureHubProfile(claims.TenantID, claims.UserID)
	routes := a.store.ListHubRoutes(hub.ID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"hub":    hub,
		"routes": routes,
		"strategy": map[string]interface{}{
			"singleInstall":        true,
			"autoCatalogSync":      true,
			"localBridgeSupported": true,
		},
	})
}

type upsertAgentRequest struct {
	ID           string `json:"id"`
	DeviceID     string `json:"deviceId"`
	Version      string `json:"version"`
	TunnelStatus string `json:"tunnelStatus"`
}

func (a *App) listLocalAgents(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	agents := a.store.ListLocalAgents(claims.TenantID, claims.UserID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": agents, "count": len(agents)})
}

func (a *App) upsertLocalAgent(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	var req upsertAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	agent := a.store.UpsertLocalAgent(models.LocalAgent{
		ID:           req.ID,
		TenantID:     claims.TenantID,
		UserID:       claims.UserID,
		DeviceID:     req.DeviceID,
		Version:      req.Version,
		TunnelStatus: req.TunnelStatus,
	})
	writeJSON(w, http.StatusCreated, agent)
}

func (a *App) rotateBuyerConnectionToken(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims, _ := getClaims(r.Context())
	connections := a.store.ListConnections(claims.TenantID, claims.UserID)
	for _, conn := range connections {
		if conn.ID == id {
			conn.TokenExpiresAt = time.Now().UTC().Add(365 * 24 * time.Hour)
			a.store.UpsertConnection(conn)
			writeJSON(w, http.StatusOK, map[string]interface{}{"id": id, "status": "rotated", "tokenExpiresAt": conn.TokenExpiresAt})
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
}

func (a *App) revokeBuyerConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims, _ := getClaims(r.Context())
	connections := a.store.ListConnections(claims.TenantID, claims.UserID)
	for _, conn := range connections {
		if conn.ID == id {
			conn.Status = "revoked"
			a.store.UpsertConnection(conn)
			writeJSON(w, http.StatusOK, map[string]interface{}{"id": id, "status": "revoked"})
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
}
