package http

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func (a *App) listTenants(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": a.store.ListTenants()})
}

func (a *App) listUsers(w http.ResponseWriter, r *http.Request) {
	tenants := a.store.ListTenants()
	tenantByID := make(map[string]models.Tenant, len(tenants))
	for _, tenant := range tenants {
		tenantByID[tenant.ID] = tenant
	}

	items := make([]map[string]interface{}, 0)
	for _, user := range a.store.ListUsers() {
		tenant := tenantByID[user.TenantID]
		items = append(items, map[string]interface{}{
			"id":         user.ID,
			"tenantId":   user.TenantID,
			"tenantName": tenant.Name,
			"tenantSlug": tenant.Slug,
			"email":      user.Email,
			"name":       user.Name,
			"role":       user.Role,
			"mfaEnabled": user.MFAEnabled,
			"createdAt":  user.CreatedAt,
			"updatedAt":  user.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

func (a *App) adminMarketplaceInsights(w http.ResponseWriter, r *http.Request) {
	tenants := a.store.ListTenants()
	serverByID := map[string]models.Server{}
	for _, tenant := range tenants {
		for _, server := range a.store.ListMerchantServers(tenant.ID) {
			serverByID[server.ID] = server
		}
	}

	allIntents := a.store.ListAllX402Intents()
	intentCountByServer := map[string]int{}
	settledUsdcByServer := map[string]float64{}
	for _, intent := range allIntents {
		if strings.TrimSpace(intent.ServerID) == "" {
			continue
		}
		intentCountByServer[intent.ServerID]++
		if intent.Status == "settled" {
			settledUsdcByServer[intent.ServerID] += intent.AmountUSDC
		}
	}

	type serverInsight struct {
		ID              string  `json:"id"`
		Name            string  `json:"name"`
		Slug            string  `json:"slug"`
		TenantID        string  `json:"tenantId"`
		Status          string  `json:"status"`
		Deployment      string  `json:"deploymentStatus"`
		PricingType     string  `json:"pricingType"`
		InstallCount    int     `json:"installCount"`
		X402IntentCount int     `json:"x402IntentCount"`
		SettledVolume   float64 `json:"settledVolumeUsdc"`
	}

	items := make([]serverInsight, 0, len(serverByID))
	for _, server := range serverByID {
		items = append(items, serverInsight{
			ID:              server.ID,
			Name:            server.Name,
			Slug:            server.Slug,
			TenantID:        server.TenantID,
			Status:          server.Status,
			Deployment:      server.DeploymentStatus,
			PricingType:     server.PricingType,
			InstallCount:    server.InstallCount,
			X402IntentCount: intentCountByServer[server.ID],
			SettledVolume:   settledUsdcByServer[server.ID],
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].InstallCount == items[j].InstallCount {
			return items[i].SettledVolume > items[j].SettledVolume
		}
		return items[i].InstallCount > items[j].InstallCount
	})

	totalInstalls := 0
	for _, item := range items {
		totalInstalls += item.InstallCount
	}

	limit := 5
	popular := items
	if len(popular) > limit {
		popular = popular[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":         items,
		"count":         len(items),
		"totalInstalls": totalInstalls,
		"popular":       popular,
	})
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
