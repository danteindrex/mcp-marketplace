package http

import (
	"net/http"
	"time"
)

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "mcp-marketplace-backend",
		"time":    time.Now().UTC(),
	})
}
