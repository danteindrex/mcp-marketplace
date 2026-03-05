package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func (a *App) listBuyerConnections(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	connections := a.store.ListConnections(claims.TenantID, claims.UserID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": connections, "count": len(connections)})
}

type createConnectionRequest struct {
	Client        string   `json:"client"`
	Resource      string   `json:"resource"`
	GrantedScopes []string `json:"grantedScopes"`
}

func (a *App) createBuyerConnection(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	var req createConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Client == "" || req.Resource == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	hub, ok := a.store.GetHubProfile(claims.TenantID, claims.UserID)
	if !ok {
		hub = a.store.UpsertHubProfile(models.HubProfile{
			TenantID: claims.TenantID,
			UserID:   claims.UserID,
			HubURL:   a.cfg.BaseURL + "/mcp/hub/" + claims.TenantID + "/" + claims.UserID,
			Status:   "active",
		})
	}
	conn := a.store.UpsertConnection(models.Connection{
		TenantID:       claims.TenantID,
		UserID:         claims.UserID,
		HubID:          hub.ID,
		Client:         req.Client,
		Status:         "active",
		Resource:       req.Resource,
		GrantedScopes:  req.GrantedScopes,
		TokenExpiresAt: time.Now().UTC().Add(8 * time.Hour),
		CatalogVersion: hub.CatalogVersion,
	})
	a.store.AddAuditLog(models.AuditLog{TenantID: claims.TenantID, ActorID: claims.UserID, Action: "connection.create", TargetType: "connection", TargetID: conn.ID, Outcome: "success", Metadata: map[string]interface{}{"client": req.Client}})
	writeJSON(w, http.StatusCreated, conn)
}

func (a *App) listBuyerEntitlements(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	entitlements := a.store.ListEntitlements(claims.TenantID, claims.UserID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": entitlements, "count": len(entitlements)})
}

func (a *App) getBuyerHub(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	hub, ok := a.store.GetHubProfile(claims.TenantID, claims.UserID)
	if !ok {
		hub = a.store.UpsertHubProfile(models.HubProfile{
			TenantID: claims.TenantID,
			UserID:   claims.UserID,
			HubURL:   a.cfg.BaseURL + "/mcp/hub/" + claims.TenantID + "/" + claims.UserID,
			Status:   "active",
		})
	}
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
