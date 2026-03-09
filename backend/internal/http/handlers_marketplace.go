package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func (a *App) listMarketplaceServers(w http.ResponseWriter, r *http.Request) {
	all := a.store.ListMarketplaceServers()
	servers := make([]models.Server, 0, len(all))
	for i := range all {
		a.normalizeServerLifecycleForView(&all[i])
		if a.isServerMarketplaceVisible(all[i]) {
			servers = append(servers, all[i])
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": servers, "count": len(servers)})
}

func (a *App) getMarketplaceServer(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	server, ok := a.store.GetServerBySlug(slug)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if !a.isServerMarketplaceVisible(server) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"server": server,
		"install": map[string]interface{}{
			"oneClick":         true,
			"hubStrategy":      "single-personal-hub",
			"clients":          []string{"vscode", "codex", "claude", "cursor", "chatgpt"},
			"installEndpoint":  "/v1/marketplace/servers/" + server.Slug + "/install",
			"supportsCommands": true,
		},
	})
}
