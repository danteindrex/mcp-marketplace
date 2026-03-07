package http

import (
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"
)

func (a *App) getBuyerBilling(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	intents := a.store.ListX402Intents(claims.TenantID, claims.UserID)
	monthlySpend := 0.0
	for _, intent := range intents {
		if intent.Status == "settled" && intent.SettledAt.Month() == time.Now().UTC().Month() && intent.SettledAt.Year() == time.Now().UTC().Year() {
			monthlySpend += intent.AmountUSDC
		}
	}
	entitlements := a.store.ListEntitlements(claims.TenantID, claims.UserID)
	plan := "free"
	if len(entitlements) > 0 {
		plan = "pro"
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":              "bill_" + claims.UserID,
		"userId":          claims.UserID,
		"plan":            plan,
		"monthlySpend":    monthlySpend,
		"currentBalance":  0.0,
		"nextBillingDate": time.Now().UTC().AddDate(0, 1, 0),
		"paymentMethod":   "",
		"status":          "active",
	})
}

func (a *App) listBuyerInvoices(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	intents := a.store.ListX402Intents(claims.TenantID, claims.UserID)
	items := make([]map[string]interface{}, 0)
	for _, intent := range intents {
		if intent.Status == "settled" {
			items = append(items, map[string]interface{}{
				"id":     "inv_" + intent.ID,
				"date":   intent.SettledAt,
				"amount": intent.AmountUSDC,
				"status": "paid",
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		ti, _ := items[i]["date"].(time.Time)
		tj, _ := items[j]["date"].(time.Time)
		return ti.After(tj)
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items, "count": len(items)})
}

func (a *App) merchantRevenue(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	servers := a.store.ListMerchantServers(claims.TenantID)
	serverRows := make([]map[string]interface{}, 0, len(servers))
	totalRevenue := 0.0
	totalCustomers := 0
	for _, server := range servers {
		revenue := float64(server.InstallCount) * server.PricingAmount * 0.3
		if server.PricingType == "free" {
			revenue = 0
		}
		totalRevenue += revenue
		totalCustomers += server.InstallCount
		serverRows = append(serverRows, map[string]interface{}{
			"id":        server.ID,
			"name":      server.Name,
			"revenue":   revenue,
			"customers": server.InstallCount,
			"trend":     "+8%",
		})
	}
	trend := []map[string]interface{}{
		{"month": "Jan", "revenue": totalRevenue * 0.65, "subscriptions": totalRevenue * 0.45, "perCall": totalRevenue * 0.2},
		{"month": "Feb", "revenue": totalRevenue * 0.75, "subscriptions": totalRevenue * 0.5, "perCall": totalRevenue * 0.25},
		{"month": "Mar", "revenue": totalRevenue * 0.85, "subscriptions": totalRevenue * 0.55, "perCall": totalRevenue * 0.3},
		{"month": "Apr", "revenue": totalRevenue * 0.9, "subscriptions": totalRevenue * 0.6, "perCall": totalRevenue * 0.3},
		{"month": "May", "revenue": totalRevenue * 0.95, "subscriptions": totalRevenue * 0.62, "perCall": totalRevenue * 0.33},
		{"month": "Jun", "revenue": totalRevenue, "subscriptions": totalRevenue * 0.64, "perCall": totalRevenue * 0.36},
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"totalRevenue":   totalRevenue,
		"totalCustomers": totalCustomers,
		"servers":        serverRows,
		"trend":          trend,
	})
}

func (a *App) serverDeployments(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	items := []map[string]interface{}{
		{"id": "dep_prod_" + id, "environment": "production", "region": "us-west-1", "replicas": 3, "status": "healthy", "transport": "sse", "version": server.Version, "updatedAt": time.Now().UTC()},
		{"id": "dep_stg_" + id, "environment": "staging", "region": "us-west-2", "replicas": 1, "status": "healthy", "transport": "sse", "version": server.Version, "updatedAt": time.Now().UTC().Add(-24 * time.Hour)},
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items, "count": len(items)})
}

func (a *App) serverBuilder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"serverId": server.ID,
		"toolCatalog": []map[string]interface{}{
			{"name": "query", "inputSchema": map[string]string{"query": "string"}, "outputSchema": map[string]string{"rows": "array"}},
			{"name": "analyze", "inputSchema": map[string]string{"data": "array"}, "outputSchema": map[string]string{"summary": "object"}},
		},
		"scopeMappings": server.RequiredScopes,
		"framework":     "FastMCP",
	})
}

func (a *App) clientCompatibility(w http.ResponseWriter, r *http.Request) {
	items := []map[string]interface{}{
		{"client": "codex", "supportsDCR": true, "supportsCIMD": true, "supportsInteractive": true, "notes": "Hub-based one-click supported"},
		{"client": "vscode", "supportsDCR": true, "supportsCIMD": true, "supportsInteractive": true, "notes": "Requires localhost and vscode.dev redirects"},
		{"client": "cursor", "supportsDCR": true, "supportsCIMD": true, "supportsInteractive": true, "notes": "OAuth public client flow"},
		{"client": "claude", "supportsDCR": true, "supportsCIMD": true, "supportsInteractive": true, "notes": "Remote MCP with OAuth"},
	}
	sort.Slice(items, func(i, j int) bool { return items[i]["client"].(string) < items[j]["client"].(string) })
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items, "count": len(items)})
}
