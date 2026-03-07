package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *App) listMarketplaceServers(w http.ResponseWriter, r *http.Request) {
	servers := a.store.ListMarketplaceServers()
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": servers, "count": len(servers)})
}

func (a *App) getMarketplaceServer(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	server, ok := a.store.GetServerBySlug(slug)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if server.Status != "published" {
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
