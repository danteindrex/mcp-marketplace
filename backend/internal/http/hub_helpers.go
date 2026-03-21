package http

import (
	"strings"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func (a *App) ensureHubProfile(tenantID, userID string) models.HubProfile {
	hub, ok := a.store.GetHubProfile(tenantID, userID)
	if ok {
		return hub
	}
	return a.store.UpsertHubProfile(models.HubProfile{
		TenantID: tenantID,
		UserID:   userID,
		HubURL:   a.cfg.BaseURL + "/mcp/hub/" + tenantID + "/" + userID,
		Status:   "active",
	})
}

func normalizeClientName(v string) string {
	client := strings.ToLower(strings.TrimSpace(v))
	switch client {
	case "vscode", "cursor", "claude", "codex", "chatgpt", "chatgpt_app":
		return client
	default:
		if client == "" {
			return "vscode"
		}
		return client
	}
}
