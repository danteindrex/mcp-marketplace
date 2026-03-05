package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func (a *App) listMerchantServers(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	servers := a.store.ListMerchantServers(claims.TenantID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": servers, "count": len(servers)})
}

type createServerRequest struct {
	Name                 string   `json:"name"`
	Slug                 string   `json:"slug"`
	Description          string   `json:"description"`
	Category             string   `json:"category"`
	DockerImage          string   `json:"dockerImage"`
	CanonicalResourceURI string   `json:"canonicalResourceUri"`
	RequiredScopes       []string `json:"requiredScopes"`
	PricingType          string   `json:"pricingType"`
	PricingAmount        float64  `json:"pricingAmount"`
	SupportsLocal        bool     `json:"supportsLocal"`
	SupportsCloud        bool     `json:"supportsCloud"`
}

func (a *App) createMerchantServer(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	var req createServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.Slug == "" || req.DockerImage == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	server := models.Server{
		TenantID:             claims.TenantID,
		Name:                 req.Name,
		Slug:                 req.Slug,
		Description:          req.Description,
		Category:             req.Category,
		Version:              "1.0.0",
		DockerImage:          req.DockerImage,
		CanonicalResourceURI: req.CanonicalResourceURI,
		RequiredScopes:       req.RequiredScopes,
		PricingType:          req.PricingType,
		PricingAmount:        req.PricingAmount,
		Status:               "draft",
		SupportsCloud:        req.SupportsCloud,
		SupportsLocal:        req.SupportsLocal,
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}
	server = a.store.CreateServer(server)
	a.store.AddAuditLog(models.AuditLog{TenantID: claims.TenantID, ActorID: claims.UserID, Action: "server.create", TargetType: "server", TargetID: server.ID, Outcome: "success"})
	writeJSON(w, http.StatusCreated, server)
}

func (a *App) updateMerchantServer(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok || (server.TenantID != claims.TenantID && claims.Role != models.RoleAdmin) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	var req createServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.Name != "" {
		server.Name = req.Name
	}
	if req.Description != "" {
		server.Description = req.Description
	}
	if req.DockerImage != "" {
		server.DockerImage = req.DockerImage
	}
	if req.CanonicalResourceURI != "" {
		server.CanonicalResourceURI = req.CanonicalResourceURI
	}
	if len(req.RequiredScopes) > 0 {
		server.RequiredScopes = req.RequiredScopes
	}
	if req.PricingType != "" {
		server.PricingType = req.PricingType
		server.PricingAmount = req.PricingAmount
	}
	server.UpdatedAt = time.Now().UTC()
	a.store.UpdateServer(server)
	a.store.AddAuditLog(models.AuditLog{TenantID: claims.TenantID, ActorID: claims.UserID, Action: "server.update", TargetType: "server", TargetID: server.ID, Outcome: "success"})
	writeJSON(w, http.StatusOK, server)
}

func (a *App) serverObservability(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"serverId": server.ID,
		"metrics": map[string]interface{}{
			"p50LatencyMs":           82,
			"p95LatencyMs":           250,
			"errorRate":              0.014,
			"insufficientScopeCount": 5,
		},
	})
}

func (a *App) serverAuthConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"serverId": server.ID,
		"oauth": map[string]interface{}{
			"pkceRequired":              true,
			"resourceIndicatorRequired": true,
			"canonicalResourceUri":      server.CanonicalResourceURI,
			"registrationModes":         []string{"pre_registered", "cimd", "dcr"},
		},
	})
}

func (a *App) serverPricing(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"serverId": server.ID,
		"pricing": map[string]interface{}{
			"type":   server.PricingType,
			"amount": server.PricingAmount,
			"x402": map[string]interface{}{
				"version": "v2",
				"network": "base",
				"asset":   "USDC",
				"caip2":   "eip155:8453",
			},
		},
	})
}
