package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
	response := map[string]interface{}{
		"server": server,
		"lifecycle": map[string]interface{}{
			"marketplaceStatus": server.Status,
			"deploymentStatus":  server.DeploymentStatus,
			"canPublish":        server.DeploymentStatus == models.ServerDeploymentDeployed && server.PricingAmount > 0,
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
	DockerImage          string   `json:"dockerImage"`
	CanonicalResourceURI string   `json:"canonicalResourceUri"`
	RequiredScopes       []string `json:"requiredScopes"`
	PricingType          string   `json:"pricingType"`
	PricingAmount        float64  `json:"pricingAmount"`
	SupportsLocal        bool     `json:"supportsLocal"`
	SupportsCloud        bool     `json:"supportsCloud"`
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.Slug == "" || req.DockerImage == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	status, ok := normalizeServerStatus(req.Status, models.ServerStatusDraft)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "status must be draft, published, or archived"})
		return
	}
	now := time.Now().UTC()
	server := models.Server{
		TenantID:             claims.TenantID,
		Author:               claims.TenantID,
		Name:                 req.Name,
		Slug:                 req.Slug,
		Description:          req.Description,
		Category:             req.Category,
		Version:              "1.0.0",
		DockerImage:          req.DockerImage,
		CanonicalResourceURI: req.CanonicalResourceURI,
		RequiredScopes:       req.RequiredScopes,
		PricingType:          req.PricingType,
		PricingAmount:        req.PricingAmount,
		Status:               status,
		DeploymentStatus:     models.ServerDeploymentPending,
		SupportsCloud:        req.SupportsCloud,
		SupportsLocal:        req.SupportsLocal,
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
	if err := a.validateEnabledPaymentMethods(server.PaymentMethods); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if status == models.ServerStatusPublished {
		server.DeploymentStatus = models.ServerDeploymentDeployed
		server.DeployedAt = now
		server.DeployedBy = claims.UserID
		server.PublishedAt = now
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
	if req.CanonicalResourceURI != "" {
		server.CanonicalResourceURI = req.CanonicalResourceURI
	}
	if len(req.RequiredScopes) > 0 {
		server.RequiredScopes = req.RequiredScopes
	}
	if req.PricingType != "" {
		server.PricingType = req.PricingType
		server.PricingAmount = req.PricingAmount
	}
	if len(req.PaymentMethods) > 0 {
		server.PaymentMethods = normalizePaymentMethods(req.PaymentMethods)
		if err := a.validateEnabledPaymentMethods(server.PaymentMethods); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
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
		server.Status = normalizedStatus
	}
	if server.Status == models.ServerStatusPublished && server.PricingAmount <= 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "price must be set before publish"})
		return
	}
	if server.Status == models.ServerStatusPublished && server.DeploymentStatus != models.ServerDeploymentDeployed {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "server must be deployed before publish"})
		return
	}
	if server.Status == models.ServerStatusPublished && server.PublishedAt.IsZero() {
		server.PublishedAt = time.Now().UTC()
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
	if a.n8n != nil && a.n8n.configured() {
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
			"server": server,
			"lifecycle": map[string]interface{}{
				"marketplaceStatus": server.Status,
				"deploymentStatus":  server.DeploymentStatus,
				"canPublish":        false,
			},
			"queue": map[string]interface{}{
				"taskId":        task.ID,
				"status":        task.Status,
				"attemptCount":  task.AttemptCount,
				"maxAttempts":   task.MaxAttempts,
				"nextAttemptAt": task.NextAttemptAt,
			},
			"n8n": map[string]interface{}{
				"configured":  true,
				"workflowId":  server.N8nWorkflowID,
				"workflowUrl": server.N8nWorkflowURL,
			},
		})
		return
	}

	server.DeploymentStatus = models.ServerDeploymentDeployed
	server.DeployedAt = now
	server.DeployedBy = claims.UserID
	// Deploying exposes the agent endpoint while keeping marketplace listing as draft.
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
			"workflowId":        server.N8nWorkflowID,
			"workflowUrl":       server.N8nWorkflowURL,
		},
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"server": server,
		"lifecycle": map[string]interface{}{
			"marketplaceStatus": server.Status,
			"deploymentStatus":  server.DeploymentStatus,
			"canPublish":        server.PricingAmount > 0,
		},
		"n8n": map[string]interface{}{
			"configured":  false,
			"workflowId":  server.N8nWorkflowID,
			"workflowUrl": server.N8nWorkflowURL,
		},
	})
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
	if server.DeploymentStatus != models.ServerDeploymentDeployed {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "deploy server before publishing to marketplace"})
		return
	}
	if server.PricingAmount <= 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "price must be set before publish"})
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
		"server": server,
		"lifecycle": map[string]interface{}{
			"marketplaceStatus": server.Status,
			"deploymentStatus":  server.DeploymentStatus,
			"canPublish":        true,
		},
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
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"serverId": server.ID,
		"metrics": map[string]interface{}{
			"p50LatencyMs":           82,
			"p95LatencyMs":           250,
			"errorRate":              0.014,
			"insufficientScopeCount": 5,
		},
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
			"pkceRequired":              true,
			"resourceIndicatorRequired": true,
			"canonicalResourceUri":      server.CanonicalResourceURI,
			"registrationModes":         []string{"pre_registered", "cimd", "dcr"},
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
			"canPublish":        server.DeploymentStatus == models.ServerDeploymentDeployed && server.PricingAmount > 0,
		},
	})
}
