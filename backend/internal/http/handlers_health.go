package http

import (
	"context"
	"net/http"
	"time"
)

type dbHealthReporter interface {
	StoreType() string
	Health(ctx context.Context) error
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	storage := map[string]interface{}{
		"engine":    "unknown",
		"connected": true,
		"required":  a.cfg.MongoRequired,
	}
	status := http.StatusOK
	serviceStatus := "ok"
	if reporter, ok := a.store.(dbHealthReporter); ok {
		storage["engine"] = reporter.StoreType()
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		err := reporter.Health(ctx)
		cancel()
		if err != nil {
			storage["connected"] = false
			storage["error"] = "storage unavailable"
			if a.cfg.MongoRequired || reporter.StoreType() == "mongo" {
				status = http.StatusServiceUnavailable
				serviceStatus = "degraded"
			}
		}
	}
	writeJSON(w, status, map[string]interface{}{
		"status":  serviceStatus,
		"service": "mcp-marketplace-backend",
		"time":    time.Now().UTC(),
		"storage": storage,
	})
}
