package http

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type installServerRequest struct {
	Client        string   `json:"client"`
	GrantedScopes []string `json:"grantedScopes"`
}

type installAction struct {
	Client            string `json:"client"`
	Label             string `json:"label"`
	LaunchURL         string `json:"launchUrl,omitempty"`
	OpenURL           string `json:"openUrl,omitempty"`
	Command           string `json:"command,omitempty"`
	FallbackCopy      string `json:"fallbackCopy,omitempty"`
	Description       string `json:"description,omitempty"`
	RequiresLocalExec bool   `json:"requiresLocalExec"`
}

func (a *App) installMarketplaceServer(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	slug := strings.TrimSpace(chi.URLParam(r, "slug"))
	server, found := a.store.GetServerBySlug(slug)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if claims.Role != models.RoleAdmin && claims.TenantID != server.TenantID && server.Status != "published" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}

	var req installServerRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
	}
	client := normalizeClientName(req.Client)
	hub := a.ensureHubProfile(claims.TenantID, claims.UserID)

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

	grantedScopes := allowedScopes
	if len(req.GrantedScopes) > 0 {
		grantedScopes = filterGranted(allowedScopes, req.GrantedScopes)
		if len(grantedScopes) == 0 {
			grantedScopes = allowedScopes
		}
	}

	conn := a.store.UpsertConnection(models.Connection{
		TenantID:       claims.TenantID,
		UserID:         claims.UserID,
		ServerID:       server.ID,
		ServerSlug:     server.Slug,
		ServerName:     server.Name,
		HubID:          hub.ID,
		Client:         client,
		Status:         "active",
		Resource:       hub.HubURL,
		GrantedScopes:  grantedScopes,
		TokenExpiresAt: time.Now().UTC().Add(8 * time.Hour),
		CatalogVersion: hub.CatalogVersion,
	})

	actions := buildInstallActions(server.Name, server.Slug, hub.HubURL)
	selected := actions[0]
	for _, action := range actions {
		if action.Client == client {
			selected = action
			break
		}
	}

	a.store.AddAuditLog(models.AuditLog{
		TenantID:   claims.TenantID,
		ActorID:    claims.UserID,
		Action:     "marketplace.install",
		TargetType: "server",
		TargetID:   server.ID,
		Outcome:    "success",
		Metadata: map[string]interface{}{
			"client":     client,
			"connection": conn.ID,
		},
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"server":     server,
		"hub":        hub,
		"resource":   hub.HubURL,
		"connection": conn,
		"install": map[string]interface{}{
			"selected": selected,
			"actions":  actions,
		},
	})
}

func buildInstallActions(serverName string, serverSlug string, resourceURL string) []installAction {
	vscodePayload, _ := json.Marshal(map[string]string{
		"name":      serverName,
		"serverUrl": resourceURL,
	})
	encodedVSCode := url.QueryEscape(string(vscodePayload))
	cursorConfig, _ := json.Marshal(map[string]string{
		"url": resourceURL,
	})
	cursorValues := url.Values{}
	cursorValues.Set("name", serverName)
	cursorValues.Set("config", string(cursorConfig))
	cursorLaunchURL := "cursor://anysphere.cursor-deeplink/mcp/install?" + cursorValues.Encode()
	codexLaunchURL := buildLocalBridgeInstallURL("codex", serverSlug, resourceURL)
	claudeLaunchURL := buildLocalBridgeInstallURL("claude", serverSlug, resourceURL)
	codexCommand := "codex mcp add " + serverSlug + " --url " + resourceURL
	claudeCommand := "claude mcp add --transport http " + serverSlug + " " + resourceURL

	return []installAction{
		{
			Client:       "vscode",
			Label:        "Install in VS Code",
			LaunchURL:    "vscode:mcp/install?" + encodedVSCode,
			FallbackCopy: resourceURL,
			Description:  "One-click deep link into VS Code MCP installer.",
		},
		{
			Client:            "codex",
			Label:             "Install in Codex",
			LaunchURL:         codexLaunchURL,
			Command:           codexCommand,
			FallbackCopy:      codexCommand,
			Description:       "One-click install via MCP Local Bridge, with Codex CLI fallback.",
			RequiresLocalExec: true,
		},
		{
			Client:            "claude",
			Label:             "Install in Claude",
			LaunchURL:         claudeLaunchURL,
			Command:           claudeCommand,
			FallbackCopy:      claudeCommand,
			Description:       "One-click install via MCP Local Bridge, with Claude CLI fallback.",
			RequiresLocalExec: true,
		},
		{
			Client:       "cursor",
			Label:        "Install in Cursor",
			LaunchURL:    cursorLaunchURL,
			FallbackCopy: resourceURL,
			Description:  "One-click deep link into Cursor MCP installer.",
		},
		{
			Client:       "chatgpt",
			Label:        "Connect in ChatGPT",
			OpenURL:      "https://chatgpt.com/#settings/connectors",
			FallbackCopy: resourceURL,
			Description:  "Open connector settings and paste the remote MCP URL.",
		},
	}
}

func buildLocalBridgeInstallURL(client string, serverSlug string, resourceURL string) string {
	values := url.Values{}
	values.Set("client", client)
	values.Set("slug", serverSlug)
	values.Set("resource", resourceURL)
	return "mcp-marketplace://install?" + values.Encode()
}
