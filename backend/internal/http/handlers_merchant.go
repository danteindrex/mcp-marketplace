package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func (a *App) listMerchantServers(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	servers := a.store.ListMerchantServers(claims.TenantID)
	for i := range servers {
		a.normalizeServerLifecycleForView(&servers[i])
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": servers, "count": len(servers)})
}

func normalizeServerStatus(input, fallback string) (string, bool) {
	value := strings.ToLower(strings.TrimSpace(input))
	if value == "" {
		return fallback, true
	}
	switch value {
	case models.ServerStatusDraft, "build_draft", "marketplace_draft", "deployed_private":
		return models.ServerStatusDraft, true
	case models.ServerStatusPublished:
		return models.ServerStatusPublished, true
	case models.ServerStatusArchived:
		return models.ServerStatusArchived, true
	default:
		return "", false
	}
}

type serverBlockingReason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Stage   string `json:"stage"`
}

type serverPublishability struct {
	CanPublish      bool                   `json:"canPublish"`
	BlockingReasons []serverBlockingReason `json:"blockingReasons"`
}

func canonicalResourceBlockingReasons(server models.Server) []serverBlockingReason {
	reasons := make([]serverBlockingReason, 0, 3)
	if !server.SupportsCloud && !server.SupportsLocal {
		reasons = append(reasons, serverBlockingReason{
			Code:    "install_surface_required",
			Message: "Enable at least one buyer install surface before publishing.",
			Stage:   "distribution",
			Field:   "supportsCloud",
		})
	}

	canonical := strings.TrimSpace(server.CanonicalResourceURI)
	if canonical == "" {
		if server.SupportsCloud {
			reasons = append(reasons, serverBlockingReason{
				Code:    "upstream_runtime_required",
				Message: "Set the upstream runtime URL before publishing cloud installs.",
				Stage:   "runtime",
				Field:   "canonicalResourceUri",
			})
		}
		return reasons
	}

	parsed, err := url.Parse(canonical)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		reasons = append(reasons, serverBlockingReason{
			Code:    "upstream_runtime_invalid",
			Message: "Canonical resource URI must be an absolute URL.",
			Stage:   "runtime",
			Field:   "canonicalResourceUri",
		})
		return reasons
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "example.com" {
		reasons = append(reasons, serverBlockingReason{
			Code:    "upstream_runtime_placeholder",
			Message: "Replace placeholder upstream URLs before publishing.",
			Stage:   "runtime",
			Field:   "canonicalResourceUri",
		})
	}

	if server.SupportsCloud && (host == "localhost" || host == "127.0.0.1") {
		reasons = append(reasons, serverBlockingReason{
			Code:    "upstream_runtime_local_only",
			Message: "Cloud-published servers cannot point at localhost upstream runtimes.",
			Stage:   "runtime",
			Field:   "canonicalResourceUri",
		})
	}

	return reasons
}

func (a *App) getMerchantServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if !a.ensureServerTenantAccess(w, r, server) {
		return
	}
	a.normalizeServerLifecycleForView(&server)
	publishability := a.serverPublishability(server)
	response := map[string]interface{}{
		"server": server,
		"lifecycle": map[string]interface{}{
			"marketplaceStatus": server.Status,
			"deploymentStatus":  server.DeploymentStatus,
			"canPublish":        publishability.CanPublish,
			"blockingReasons":   publishability.BlockingReasons,
		},
	}
	if task, exists := a.store.GetDeployTaskByServer(server.ID); exists {
		response["queue"] = task
	}
	writeJSON(w, http.StatusOK, response)
}

type createServerRequest struct {
	Name                 string   `json:"name"`
	Slug                 string   `json:"slug"`
	Description          string   `json:"description"`
	Category             string   `json:"category"`
	DockerImage          string            `json:"dockerImage"`
	ContainerPort        int               `json:"containerPort"`
	CanonicalResourceURI string            `json:"canonicalResourceUri"`
	UpstreamAuthType     string            `json:"upstreamAuthType"`
	UpstreamAuthToken    string            `json:"upstreamAuthToken"`
	UpstreamHeaders      map[string]string `json:"upstreamHeaders"`
	RequiredScopes       []string          `json:"requiredScopes"`
	PricingType          string   `json:"pricingType"`
	PricingAmount        float64  `json:"pricingAmount"`
	SupportsLocal        bool     `json:"supportsLocal"`
	SupportsCloud        bool     `json:"supportsCloud"`
	SupportsChatGPTApp   bool     `json:"supportsChatGptApp"`
	ChatGPTAppURL        string   `json:"chatGptAppUrl"`
	PaymentMethods       []string `json:"paymentMethods"`
	PaymentAddress       string   `json:"paymentAddress"`
	PerCallCapUSDC       float64  `json:"perCallCapUsdc"`
	DailyCapUSDC         float64  `json:"dailyCapUsdc"`
	MonthlyCapUSDC       float64  `json:"monthlyCapUsdc"`
	Status               string   `json:"status"`
}

func (a *App) createMerchantServer(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	var req createServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.Slug == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if strings.TrimSpace(req.DockerImage) == "" && strings.TrimSpace(req.CanonicalResourceURI) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "either dockerImage or canonicalResourceUri is required"})
		return
	}
	status, ok := normalizeServerStatus(req.Status, models.ServerStatusDraft)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "status must be draft, published, or archived"})
		return
	}
	upstreamAuthType := strings.ToLower(strings.TrimSpace(req.UpstreamAuthType))
	if upstreamAuthType != "" && upstreamAuthType != "bearer" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "upstreamAuthType must be empty or 'bearer'"})
		return
	}
	if status != models.ServerStatusDraft {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "new servers must start as draft and not_deployed"})
		return
	}
	now := time.Now().UTC()
	containerPort := req.ContainerPort
	if containerPort <= 0 {
		containerPort = 3000
	}
	server := models.Server{
		TenantID:             claims.TenantID,
		Author:               claims.TenantID,
		Name:                 req.Name,
		Slug:                 req.Slug,
		Description:          req.Description,
		Category:             req.Category,
		Version:              "1.0.0",
		DockerImage:          req.DockerImage,
		ContainerPort:        containerPort,
		CanonicalResourceURI: req.CanonicalResourceURI,
		UpstreamAuthType:     upstreamAuthType,
		UpstreamAuthToken:    strings.TrimSpace(req.UpstreamAuthToken),
		UpstreamHeaders:      sanitizeUpstreamHeaders(req.UpstreamHeaders),
		RequiredScopes:       req.RequiredScopes,
		PricingType:          req.PricingType,
		PricingAmount:        req.PricingAmount,
		Status:               status,
		DeploymentStatus:     models.ServerDeploymentPending,
		SupportsCloud:        req.SupportsCloud,
		SupportsLocal:        req.SupportsLocal,
		SupportsChatGPTApp:   req.SupportsChatGPTApp,
		ChatGPTAppURL:        strings.TrimSpace(req.ChatGPTAppURL),
		PaymentMethods:       normalizePaymentMethods(req.PaymentMethods),
		PaymentAddress:       strings.TrimSpace(req.PaymentAddress),
		PerCallCapUSDC:       req.PerCallCapUSDC,
		DailyCapUSDC:         req.DailyCapUSDC,
		MonthlyCapUSDC:       req.MonthlyCapUSDC,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if len(server.PaymentMethods) == 0 {
		server.PaymentMethods = a.defaultAllowedPaymentMethods()
	}
	if server.SupportsChatGPTApp && server.ChatGPTAppURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "chatGptAppUrl is required when supportsChatGptApp is enabled"})
		return
	}
	if strings.TrimSpace(server.PaymentAddress) == "" {
		if wallet, err := a.ensureSellerManagedWallet(r.Context(), claims.TenantID); err == nil {
			server.PaymentAddress = wallet.Address
		}
	}
	if err := a.validateEnabledPaymentMethods(server.PaymentMethods); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if tenant, ok := a.store.GetTenantByID(claims.TenantID); ok {
		server.Author = tenant.Name
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
	if req.ContainerPort > 0 {
		server.ContainerPort = req.ContainerPort
	}
	if req.CanonicalResourceURI != "" {
		server.CanonicalResourceURI = req.CanonicalResourceURI
	}
	if strings.TrimSpace(req.UpstreamAuthType) != "" {
		authType := strings.ToLower(strings.TrimSpace(req.UpstreamAuthType))
		if authType != "bearer" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "upstreamAuthType must be 'bearer' when provided"})
			return
		}
		server.UpstreamAuthType = authType
	}
	if strings.TrimSpace(req.UpstreamAuthToken) != "" {
		server.UpstreamAuthToken = strings.TrimSpace(req.UpstreamAuthToken)
	}
	if req.UpstreamHeaders != nil {
		server.UpstreamHeaders = sanitizeUpstreamHeaders(req.UpstreamHeaders)
	}
	if len(req.RequiredScopes) > 0 {
		server.RequiredScopes = req.RequiredScopes
	}
	if req.PricingType != "" {
		server.PricingType = req.PricingType
		server.PricingAmount = req.PricingAmount
	}
	if req.SupportsChatGPTApp || strings.TrimSpace(req.ChatGPTAppURL) != "" {
		server.SupportsChatGPTApp = req.SupportsChatGPTApp
		server.ChatGPTAppURL = strings.TrimSpace(req.ChatGPTAppURL)
	}
	if strings.TrimSpace(req.PaymentAddress) != "" {
		server.PaymentAddress = strings.TrimSpace(req.PaymentAddress)
	} else if strings.TrimSpace(server.PaymentAddress) == "" {
		if wallet, err := a.ensureSellerManagedWallet(r.Context(), server.TenantID); err == nil {
			server.PaymentAddress = wallet.Address
		}
	}
	if len(req.PaymentMethods) > 0 {
		server.PaymentMethods = normalizePaymentMethods(req.PaymentMethods)
		if err := a.validateEnabledPaymentMethods(server.PaymentMethods); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}
	if server.SupportsChatGPTApp && server.ChatGPTAppURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "chatGptAppUrl is required when supportsChatGptApp is enabled"})
		return
	}
	if strings.TrimSpace(req.PaymentAddress) != "" {
		server.PaymentAddress = strings.TrimSpace(req.PaymentAddress)
	}
	if req.PerCallCapUSDC > 0 {
		server.PerCallCapUSDC = req.PerCallCapUSDC
	}
	if req.DailyCapUSDC > 0 {
		server.DailyCapUSDC = req.DailyCapUSDC
	}
	if req.MonthlyCapUSDC > 0 {
		server.MonthlyCapUSDC = req.MonthlyCapUSDC
	}
	if req.Status != "" {
		normalizedStatus, ok := normalizeServerStatus(req.Status, server.Status)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "status must be draft, published, or archived"})
			return
		}
		if normalizedStatus == models.ServerStatusPublished {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "use the publish endpoint to publish a server"})
			return
		}
		server.Status = normalizedStatus
	}
	if server.Status == models.ServerStatusPublished {
		publishability := a.serverPublishability(server)
		if !publishability.CanPublish {
			writeJSON(w, publishabilityStatusCode(publishability), map[string]interface{}{
				"error":           "published servers must remain publishable",
				"blockingReasons": publishability.BlockingReasons,
			})
			return
		}
	}
	if server.Status != models.ServerStatusPublished {
		server.PublishedAt = time.Time{}
	}
	server.UpdatedAt = time.Now().UTC()
	a.store.UpdateServer(server)
	a.store.AddAuditLog(models.AuditLog{TenantID: claims.TenantID, ActorID: claims.UserID, Action: "server.update", TargetType: "server", TargetID: server.ID, Outcome: "success"})
	writeJSON(w, http.StatusOK, server)
}

type deployServerRequest struct {
	DeploymentTarget string `json:"deploymentTarget"`
	N8nWorkflowID    string `json:"n8nWorkflowId"`
	N8nWorkflowURL   string `json:"n8nWorkflowUrl"`
}

func (a *App) deployMerchantServer(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok || (server.TenantID != claims.TenantID && claims.Role != models.RoleAdmin) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if server.Status == models.ServerStatusArchived {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "archived servers cannot be deployed"})
		return
	}
	req := deployServerRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	a.normalizeServerLifecycleForView(&server)

	if strings.TrimSpace(req.DeploymentTarget) != "" {
		server.DeploymentTarget = strings.TrimSpace(req.DeploymentTarget)
	}
	if strings.TrimSpace(req.N8nWorkflowID) != "" {
		server.N8nWorkflowID = strings.TrimSpace(req.N8nWorkflowID)
	}
	if strings.TrimSpace(req.N8nWorkflowURL) != "" {
		server.N8nWorkflowURL = strings.TrimSpace(req.N8nWorkflowURL)
	}

	now := time.Now().UTC()
	target := strings.TrimSpace(server.DeploymentTarget)
	if target == "" {
		target = "local-docker"
		server.DeploymentTarget = target
	}

	if target == "external" {
		server.DeploymentStatus = models.ServerDeploymentDeployed
		server.DeployedAt = now
		if server.Status != models.ServerStatusPublished {
			server.Status = models.ServerStatusDraft
			server.PublishedAt = time.Time{}
		}
		server.UpdatedAt = now
		a.store.UpdateServer(server)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"server":    server,
			"lifecycle": a.serverLifecyclePayload(server),
			"n8n": map[string]interface{}{
				"configured": false,
			},
		})
		return
	}

	n8nConfigured := a.currentN8NService() != nil && a.currentN8NService().configured()
	if n8nConfigured && strings.TrimSpace(req.DeploymentTarget) == "" && target == "local-docker" {
		server.DeploymentTarget = "managed-cloud"
		target = server.DeploymentTarget
	}
	if n8nConfigured {
		preferredWorkflowID := strings.TrimSpace(req.N8nWorkflowID)
		if preferredWorkflowID == "" {
			preferredWorkflowID = strings.TrimSpace(server.N8nWorkflowID)
		}
		task := a.store.UpsertDeployTaskByServer(models.DeployTask{
			TenantID:            claims.TenantID,
			ServerID:            server.ID,
			RequestedBy:         claims.UserID,
			PreferredWorkflowID: preferredWorkflowID,
			DeploymentTarget:    server.DeploymentTarget,
			Status:              models.DeployTaskStatusPending,
			AttemptCount:        0,
			MaxAttempts:         6,
			NextAttemptAt:       now,
			LastError:           "",
		})
		server.DeploymentStatus = models.ServerDeploymentQueued
		if server.Status != models.ServerStatusPublished {
			server.Status = models.ServerStatusDraft
			server.PublishedAt = time.Time{}
		}
		server.UpdatedAt = now
		a.store.UpdateServer(server)
		a.store.AddAuditLog(models.AuditLog{
			TenantID:   claims.TenantID,
			ActorID:    claims.UserID,
			Action:     "server.deploy.queued",
			TargetType: "server",
			TargetID:   server.ID,
			Outcome:    "success",
			Metadata: map[string]interface{}{
				"taskId":            task.ID,
				"deploymentStatus":  server.DeploymentStatus,
				"marketplaceStatus": server.Status,
				"workflowId":        server.N8nWorkflowID,
				"workflowUrl":       server.N8nWorkflowURL,
			},
		})
		a.triggerDeployWorker()
		writeJSON(w, http.StatusAccepted, map[string]interface{}{
			"server":    server,
			"lifecycle": a.serverLifecyclePayload(server),
			"queue": map[string]interface{}{
				"taskId":        task.ID,
				"status":        task.Status,
				"attemptCount":  task.AttemptCount,
				"maxAttempts":   task.MaxAttempts,
				"nextAttemptAt": task.NextAttemptAt,
			},
			"n8n": map[string]interface{}{
				"configured":  n8nConfigured,
				"workflowId":  server.N8nWorkflowID,
				"workflowUrl": server.N8nWorkflowURL,
			},
		})
		return
	}

	if a.cfg.AllowInsecureDefaults {
		server.DeploymentStatus = models.ServerDeploymentDeployed
		server.DeployedAt = now
		if server.Status != models.ServerStatusPublished {
			server.Status = models.ServerStatusDraft
			server.PublishedAt = time.Time{}
		}
		server.UpdatedAt = now
		a.store.UpdateServer(server)
		a.store.AddAuditLog(models.AuditLog{
			TenantID:   claims.TenantID,
			ActorID:    claims.UserID,
			Action:     "server.deploy",
			TargetType: "server",
			TargetID:   server.ID,
			Outcome:    "success",
			Metadata: map[string]interface{}{
				"deploymentStatus":  server.DeploymentStatus,
				"marketplaceStatus": server.Status,
				"target":            target,
				"mode":              "insecure-defaults",
			},
		})
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"server":    server,
			"lifecycle": a.serverLifecyclePayload(server),
			"n8n": map[string]interface{}{
				"configured":  false,
				"workflowId":  server.N8nWorkflowID,
				"workflowUrl": server.N8nWorkflowURL,
			},
		})
		return
	}

	writeJSON(w, http.StatusConflict, map[string]string{"error": "no deployment provider is configured for this target"})
}

type publishServerRequest struct {
	PricingType   string   `json:"pricingType"`
	PricingAmount *float64 `json:"pricingAmount"`
}

func (a *App) publishMerchantServer(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok || (server.TenantID != claims.TenantID && claims.Role != models.RoleAdmin) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if server.Status == models.ServerStatusArchived {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "archived servers cannot be published"})
		return
	}
	req := publishServerRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if strings.TrimSpace(req.PricingType) != "" {
		server.PricingType = strings.TrimSpace(req.PricingType)
	}
	if req.PricingAmount != nil {
		server.PricingAmount = *req.PricingAmount
	}
	publishability := a.serverPublishability(server)
	if !publishability.CanPublish {
		writeJSON(w, publishabilityStatusCode(publishability), map[string]interface{}{
			"error":           "server is not publishable",
			"blockingReasons": publishability.BlockingReasons,
		})
		return
	}
	now := time.Now().UTC()
	server.Status = models.ServerStatusPublished
	server.PublishedAt = now
	server.UpdatedAt = now
	a.store.UpdateServer(server)
	a.store.AddAuditLog(models.AuditLog{
		TenantID:   claims.TenantID,
		ActorID:    claims.UserID,
		Action:     "server.publish",
		TargetType: "server",
		TargetID:   server.ID,
		Outcome:    "success",
		Metadata: map[string]interface{}{
			"marketplaceStatus": server.Status,
			"pricingAmount":     server.PricingAmount,
		},
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"server":    server,
		"lifecycle": a.serverLifecyclePayload(server),
	})
}

func (a *App) normalizeServerLifecycleForView(server *models.Server) {
	if strings.TrimSpace(server.Status) == "" {
		server.Status = models.ServerStatusDraft
	}
	if strings.TrimSpace(server.DeploymentStatus) == "" {
		server.DeploymentStatus = models.ServerDeploymentPending
	}
}

func (a *App) serverPublishability(server models.Server) serverPublishability {
	reasons := make([]serverBlockingReason, 0, 4)
	if server.Status == models.ServerStatusArchived {
		reasons = append(reasons, serverBlockingReason{Code: "server_archived", Message: "Archived servers cannot be published.", Stage: "marketplace", Field: "status"})
	}
	if server.DeploymentStatus != models.ServerDeploymentDeployed {
		reasons = append(reasons, serverBlockingReason{Code: "server_not_deployed", Message: "Deploy the server before publishing.", Stage: "deployment", Field: "deploymentStatus"})
	}
	if server.PricingType == "x402" && server.PricingAmount <= 0 {
		reasons = append(reasons, serverBlockingReason{Code: "pricing_amount_required", Message: "Set a positive price before publishing x402 servers.", Stage: "pricing", Field: "pricingAmount"})
	}
	if server.PricingType == "x402" {
		if len(server.PaymentMethods) == 0 {
			reasons = append(reasons, serverBlockingReason{Code: "payment_methods_required", Message: "x402 servers need at least one enabled payment method.", Stage: "payments", Field: "paymentMethods"})
		} else if err := a.validateEnabledPaymentMethods(server.PaymentMethods); err != nil {
			reasons = append(reasons, serverBlockingReason{Code: "payment_methods_disabled", Message: err.Error(), Stage: "payments", Field: "paymentMethods"})
		}
	}
	reasons = append(reasons, canonicalResourceBlockingReasons(server)...)
	if a.cfg.MCPSDKEnabled {
		filtered := make([]serverBlockingReason, 0, len(reasons))
		for _, reason := range reasons {
			if reason.Code == "upstream_runtime_local_only" {
				continue
			}
			filtered = append(filtered, reason)
		}
		reasons = filtered
	}
	return serverPublishability{CanPublish: len(reasons) == 0, BlockingReasons: reasons}
}

func (a *App) serverLifecyclePayload(server models.Server) map[string]interface{} {
	publishability := a.serverPublishability(server)
	return map[string]interface{}{
		"marketplaceStatus": server.Status,
		"deploymentStatus":  server.DeploymentStatus,
		"canPublish":        publishability.CanPublish,
		"blockingReasons":   publishability.BlockingReasons,
	}
}

func (a *App) isServerMarketplaceVisible(server models.Server) bool {
	a.normalizeServerLifecycleForView(&server)
	if server.Status != models.ServerStatusPublished || server.DeploymentStatus != models.ServerDeploymentDeployed {
		return false
	}
	return a.serverPublishability(server).CanPublish
}

func sanitizeUpstreamHeaders(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		if strings.EqualFold(key, "authorization") {
			continue
		}
		out[key] = strings.TrimSpace(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func publishabilityStatusCode(publishability serverPublishability) int {
	for _, reason := range publishability.BlockingReasons {
		if reason.Stage == "deployment" || reason.Stage == "marketplace" {
			return http.StatusConflict
		}
	}
	if slices.ContainsFunc(publishability.BlockingReasons, func(reason serverBlockingReason) bool {
		return reason.Stage == "pricing" || reason.Stage == "payments" || reason.Stage == "runtime" || reason.Stage == "distribution"
	}) {
		return http.StatusUnprocessableEntity
	}
	return http.StatusConflict
}

func (a *App) serverObservability(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if !a.ensureServerTenantAccess(w, r, server) {
		return
	}

	// Get real metrics from X402 intents if available
	allIntents := a.store.ListAllX402Intents()
	var totalRequests int
	var settledAmount float64
	var pendingCount int

	for _, intent := range allIntents {
		if intent.ServerID == server.ID {
			totalRequests++
			if intent.Status == "settled" {
				settledAmount += intent.SellerNetUSDC
			} else if intent.Status == "pending" || intent.Status == "created" {
				pendingCount++
			}
		}
	}

	// Return real data or explicit "no data" state
	hasData := totalRequests > 0
	metrics := map[string]interface{}{
		"hasData": hasData,
	}

	if hasData {
		// Calculate real metrics from actual data
		errorCount := 0
		for _, intent := range allIntents {
			if intent.ServerID == server.ID && intent.Status == "settled" {
				if intent.VerificationStatus == "failed" || intent.VerificationStatus == "rejected" {
					errorCount++
				}
			}
		}

		errorRate := 0.0
		if totalRequests > 0 {
			errorRate = float64(errorCount) / float64(totalRequests)
		}

		metrics["p50LatencyMs"] = nil // No latency tracking available
		metrics["p95LatencyMs"] = nil // No latency tracking available
		metrics["errorRate"] = errorRate
		metrics["totalRequests"] = totalRequests
		metrics["settledRevenue"] = settledAmount
		metrics["pendingCount"] = pendingCount
		metrics["insufficientScopeCount"] = errorCount
	} else {
		metrics["p50LatencyMs"] = nil
		metrics["p95LatencyMs"] = nil
		metrics["errorRate"] = nil
		metrics["totalRequests"] = 0
		metrics["settledRevenue"] = 0.0
		metrics["pendingCount"] = 0
		metrics["insufficientScopeCount"] = 0
		metrics["message"] = "No observability data available. Metrics will appear once the server receives traffic."
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"serverId": server.ID,
		"metrics":  metrics,
	})
}

func (a *App) serverAuthConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if !a.ensureServerTenantAccess(w, r, server) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"serverId": server.ID,
		"oauth": map[string]interface{}{
			"pkceRequired":               true,
			"resourceIndicatorRequired":  true,
			"canonicalResourceUri":       server.CanonicalResourceURI,
			"upstreamResourceUrl":        server.CanonicalResourceURI,
			"marketplaceMetadataUrl":     strings.TrimRight(a.cfg.BaseURL, "/") + "/.well-known/mcp.json",
			"marketplaceResourceTemplate": strings.TrimRight(a.cfg.BaseURL, "/") + "/mcp/hub/{tenantID}/{userID}",
			"registrationModes":          []string{"pre_registered", "cimd", "dcr"},
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
	if !a.ensureServerTenantAccess(w, r, server) {
		return
	}
	a.normalizeServerLifecycleForView(&server)
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
			"methods":        server.PaymentMethods,
			"paymentAddress": server.PaymentAddress,
			"caps": map[string]float64{
				"perCallCapUsdc": server.PerCallCapUSDC,
				"dailyCapUsdc":   server.DailyCapUSDC,
				"monthlyCapUsdc": server.MonthlyCapUSDC,
			},
		},
		"supportedMethods": a.paymentMethodsCatalog(),
		"lifecycle": map[string]interface{}{
			"marketplaceStatus": server.Status,
			"deploymentStatus":  server.DeploymentStatus,
			"canPublish":        a.serverPublishability(server).CanPublish,
			"blockingReasons":   a.serverPublishability(server).BlockingReasons,
		},
	})
}
