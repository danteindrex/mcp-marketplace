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
	Client          string                 `json:"client"`
	GrantedScopes   []string               `json:"grantedScopes"`
	PaymentMethod   string                 `json:"paymentMethod"`
	ToolName        string                 `json:"toolName"`
	IdempotencyKey  string                 `json:"idempotencyKey"`
	AutoSettle      bool                   `json:"autoSettle"`
	PaymentResponse map[string]interface{} `json:"paymentResponse"`
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

type scopeCheckRequest struct {
	Client        string   `json:"client"`
	GrantedScopes []string `json:"grantedScopes"`
}

type scopeCheckResponse struct {
	Server struct {
		ID          string `json:"id"`
		Slug        string `json:"slug"`
		Name        string `json:"name"`
		PricingType string `json:"pricingType"`
	} `json:"server"`
	Client             string   `json:"client"`
	AllowedScopes      []string `json:"allowedScopes"`
	GrantedScopes      []string `json:"grantedScopes"`
	HasEntitlement     bool     `json:"hasEntitlement"`
	AutoGrantAvailable bool     `json:"autoGrantAvailable"`
	PaymentRequired    bool     `json:"paymentRequired"`
	EntitlementStatus  string   `json:"entitlementStatus"`
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
	if !a.isServerMarketplaceVisible(server) {
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
			} else if server.Status == "published" && server.PricingType == "x402" {
				paid, err := a.ensureInstallPayment(w, r, claims.TenantID, claims.UserID, hub.HubURL, server, req)
				if err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
					return
				}
				if !paid {
					return
				}
				entitlement, hasEntitlement = entitlementForServer(a.store.ListEntitlements(claims.TenantID, claims.UserID), server.ID)
				if !hasEntitlement {
					entitlement = a.store.GrantEntitlement(models.Entitlement{
						TenantID:      claims.TenantID,
						UserID:        claims.UserID,
						ServerID:      server.ID,
						AllowedScopes: server.RequiredScopes,
						CloudAllowed:  true,
						LocalAllowed:  true,
					})
				}
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

func (a *App) scopeCheckMarketplaceServer(w http.ResponseWriter, r *http.Request) {
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
	if !a.isServerMarketplaceVisible(server) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if claims.Role != models.RoleAdmin && claims.TenantID != server.TenantID && server.Status != models.ServerStatusPublished {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if server.DeploymentStatus != "" && server.DeploymentStatus != models.ServerDeploymentDeployed {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}

	var req scopeCheckRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
	}
	if strings.TrimSpace(req.Client) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "client required"})
		return
	}
	client := normalizeClientName(req.Client)
	allowedScopes := server.RequiredScopes
	grantedScopes := allowedScopes
	var hasEntitlement bool
	var entitlementStatus string
	autoGrantAvailable := false
	paymentRequired := false

	if claims.Role == models.RoleAdmin || claims.TenantID == server.TenantID {
		hasEntitlement = true
		entitlementStatus = "owner"
	} else {
		entitlements := a.store.ListEntitlements(claims.TenantID, claims.UserID)
		entitlement, ok := entitlementForServer(entitlements, server.ID)
		if ok {
			hasEntitlement = true
			entitlementStatus = "active_entitlement"
			if len(entitlement.AllowedScopes) > 0 {
				allowedScopes = entitlement.AllowedScopes
			}
		} else {
			if server.Status == models.ServerStatusPublished && server.PricingType == "free" {
				autoGrantAvailable = true
				entitlementStatus = "auto_grant_available"
			} else {
				paymentRequired = true
				entitlementStatus = "payment_required"
			}
		}
	}

	if len(req.GrantedScopes) > 0 {
		filtered := filterGranted(allowedScopes, req.GrantedScopes)
		if len(filtered) > 0 {
			grantedScopes = filtered
		}
	}

	resp := scopeCheckResponse{
		Client:             client,
		AllowedScopes:      allowedScopes,
		GrantedScopes:      grantedScopes,
		HasEntitlement:     hasEntitlement,
		AutoGrantAvailable: autoGrantAvailable,
		PaymentRequired:    paymentRequired,
		EntitlementStatus:  entitlementStatus,
	}
	resp.Server.ID = server.ID
	resp.Server.Slug = server.Slug
	resp.Server.Name = server.Name
	resp.Server.PricingType = server.PricingType

	writeJSON(w, http.StatusOK, resp)
}

func (a *App) ensureInstallPayment(
	w http.ResponseWriter,
	r *http.Request,
	tenantID string,
	userID string,
	resource string,
	server models.Server,
	req installServerRequest,
) (bool, error) {
	amount := server.PricingAmount
	if amount <= 0 {
		return false, nil
	}
	policy := a.effectivePaymentPolicy(tenantID, userID)
	method := strings.ToLower(strings.TrimSpace(req.PaymentMethod))
	if method == "" {
		method = firstMethod(policy.AllowedMethods)
	}
	if !a.isPaymentMethodEnabled(method) {
		writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
			"error":  a.describePaymentMethod(method).Notes,
			"method": method,
		})
		return false, nil
	}
	if !hasScope(policy.AllowedMethods, method) {
		writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
			"error":  "payment method is not allowed by buyer policy",
			"policy": policy,
		})
		return false, nil
	}
	if len(server.PaymentMethods) > 0 && !hasScope(server.PaymentMethods, method) {
		writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
			"error":  "payment method is not enabled by server",
			"server": server.ID,
		})
		return false, nil
	}
	intents := a.store.ListX402Intents(tenantID, userID)
	if err := a.validateCaps(policy, server, intents, amount); err != nil {
		writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
			"error": err.Error(),
			"caps":  policy,
		})
		return false, nil
	}

	toolName := strings.TrimSpace(req.ToolName)
	if toolName == "" {
		toolName = "install_" + strings.ReplaceAll(server.Slug, "-", "_")
	}
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	}
	if idempotencyKey == "" {
		idempotencyKey = "install_" + server.ID + "_" + time.Now().UTC().Format(time.RFC3339Nano)
	}
	if strings.TrimSpace(resource) == "" {
		resource = server.CanonicalResourceURI
	}
	requirement := buildX402Requirement(server, toolName, amount, resource, method, idempotencyKey)
	challengeBytes, _ := json.Marshal([]map[string]interface{}{requirement})
	intent := a.store.CreateX402Intent(models.X402Intent{
		TenantID:           tenantID,
		UserID:             userID,
		ServerID:           server.ID,
		ToolName:           toolName,
		AmountUSDC:         amount,
		Network:            "base",
		Asset:              "USDC",
		Challenge:          string(challengeBytes),
		PaymentMethod:      method,
		IdempotencyKey:     idempotencyKey,
		X402Version:        "2",
		Resource:           resource,
		VerificationStatus: "pending",
		Quantity:           1,
		RemainingQuantity:  0,
		RequestFingerprint: hashAny(map[string]interface{}{"serverId": server.ID, "toolName": toolName, "idempotencyKey": idempotencyKey}),
	})
	if !req.AutoSettle && len(req.PaymentResponse) == 0 {
		w.Header().Set("PAYMENT-REQUIRED", string(challengeBytes))
		w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope"`)
		writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
			"error":       "payment required",
			"intent":      intent,
			"requirement": requirement,
		})
		return false, nil
	}

	if method == "wallet_balance" && len(req.PaymentResponse) == 0 {
		_, paymentID, err := a.applyWalletDebit(tenantID, userID, amount, idempotencyKey)
		if err != nil {
			writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
				"error":       err.Error(),
				"intent":      intent,
				"requirement": requirement,
				"wallet": map[string]interface{}{
					"balanceUsdc":        policy.WalletBalanceUSDC,
					"minimumBalanceUsdc": policy.MinimumBalanceUSDC,
				},
			})
			return false, nil
		}
		settled, ok := a.store.SettleX402Intent(intent.ID)
		if !ok {
			return false, errInstallPaymentState
		}
		settled.PaymentIdentifier = paymentID
		settled.PaymentMethod = "wallet_balance"
		settled.Network = "base"
		settled.Asset = "USDC"
		settled.VerificationStatus = "verified"
		settled.VerificationNote = "debited from prepaid wallet balance"
		if settled.Quantity <= 0 {
			settled.Quantity = 1
		}
		if settled.RemainingQuantity <= 0 {
			settled.RemainingQuantity = settled.Quantity
		}
		_ = a.store.UpdateX402Intent(settled)
		if _, err := a.postIntentAccounting(settled); err != nil {
			return false, err
		}
		a.ensurePaidEntitlement(settled)
		return true, nil
	}

	verifyRes, err := a.x402.verifyAndSettle(r.Context(), requirement, req.PaymentResponse)
	if err != nil {
		writeJSON(w, http.StatusPaymentRequired, map[string]string{"error": err.Error()})
		return false, nil
	}
	if !verifyRes.Valid {
		writeJSON(w, http.StatusPaymentRequired, map[string]string{"error": verifyRes.Note})
		return false, nil
	}
	settled, ok := a.store.SettleX402Intent(intent.ID)
	if !ok {
		return false, errInstallPaymentState
	}
	settled.PaymentIdentifier = verifyRes.PaymentIdentifier
	settled.PaymentMethod = verifyRes.Method
	settled.Network = verifyRes.Network
	settled.Asset = verifyRes.Asset
	settled.FacilitatorTx = verifyRes.TxHash
	settled.VerificationStatus = "verified"
	settled.VerificationNote = verifyRes.Note
	if settled.Quantity <= 0 {
		settled.Quantity = 1
	}
	if settled.RemainingQuantity <= 0 {
		settled.RemainingQuantity = settled.Quantity
	}
	_ = a.store.UpdateX402Intent(settled)
	if _, err := a.postIntentAccounting(settled); err != nil {
		return false, err
	}
	a.ensurePaidEntitlement(settled)
	return true, nil
}

var errInstallPaymentState = &installPaymentError{message: "failed to persist install payment state"}

type installPaymentError struct {
	message string
}

func (e *installPaymentError) Error() string {
	return e.message
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
