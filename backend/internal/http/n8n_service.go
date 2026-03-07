package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type n8nService struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type n8nDeployResult struct {
	WorkflowID  string
	WorkflowURL string
	WebhookPath string
}

type n8nHTTPError struct {
	status  int
	message string
}

func (e *n8nHTTPError) Error() string {
	if strings.TrimSpace(e.message) == "" {
		return fmt.Sprintf("n8n request failed with status %d", e.status)
	}
	return fmt.Sprintf("n8n request failed with status %d: %s", e.status, e.message)
}

func newN8NService(cfg config.Config) *n8nService {
	timeout := cfg.N8NTimeoutSeconds
	if timeout <= 0 {
		timeout = 12
	}
	return &n8nService{
		baseURL: strings.TrimRight(strings.TrimSpace(cfg.N8NBaseURL), "/"),
		apiKey:  strings.TrimSpace(cfg.N8NAPIKey),
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

func (s *n8nService) configured() bool {
	return s != nil && strings.TrimSpace(s.baseURL) != ""
}

func (s *n8nService) deployWorkflow(ctx context.Context, server models.Server, preferredWorkflowID string) (n8nDeployResult, error) {
	if !s.configured() {
		return n8nDeployResult{}, nil
	}

	payload, webhookPath := s.workflowPayload(server)
	workflowID := strings.TrimSpace(preferredWorkflowID)
	if workflowID == "" {
		workflowID = strings.TrimSpace(server.N8nWorkflowID)
	}

	var err error
	if workflowID != "" {
		err = s.withRetry(ctx, func() error {
			return s.updateWorkflow(ctx, workflowID, payload)
		})
		if err != nil {
			if !isN8NStatus(err, http.StatusNotFound) {
				return n8nDeployResult{}, err
			}
			workflowID = ""
		}
	}

	if workflowID == "" {
		err = s.withRetry(ctx, func() error {
			var createErr error
			workflowID, createErr = s.createWorkflow(ctx, payload)
			return createErr
		})
		if err != nil {
			return n8nDeployResult{}, err
		}
	}

	if err := s.withRetry(ctx, func() error {
		return s.activateWorkflow(ctx, workflowID)
	}); err != nil {
		return n8nDeployResult{}, err
	}

	return n8nDeployResult{
		WorkflowID:  workflowID,
		WorkflowURL: s.workflowURL(workflowID),
		WebhookPath: webhookPath,
	}, nil
}

func (s *n8nService) workflowPayload(server models.Server) (map[string]interface{}, string) {
	shortID := strings.TrimSpace(server.ID)
	if len(shortID) > 8 {
		shortID = shortID[len(shortID)-8:]
	}
	if shortID == "" {
		shortID = "draft"
	}
	slug := sanitizeN8NPath(server.Slug)
	if slug == "" {
		slug = "agent"
	}
	webhookPath := fmt.Sprintf("mcp-%s-%s", slug, shortID)

	responseBody := fmt.Sprintf("{\"ok\":true,\"serverId\":\"%s\",\"serverSlug\":\"%s\",\"source\":\"n8n\"}", server.ID, server.Slug)
	return map[string]interface{}{
		"name": fmt.Sprintf("%s (%s)", strings.TrimSpace(server.Name), shortID),
		"nodes": []map[string]interface{}{
			{
				"id":          "webhook-node",
				"name":        "Webhook",
				"type":        "n8n-nodes-base.webhook",
				"typeVersion": 2,
				"position":    []int{240, 280},
				"parameters": map[string]interface{}{
					"httpMethod":   "POST",
					"path":         webhookPath,
					"responseMode": "responseNode",
					"options":      map[string]interface{}{},
				},
			},
			{
				"id":          "respond-node",
				"name":        "Respond to Webhook",
				"type":        "n8n-nodes-base.respondToWebhook",
				"typeVersion": 1,
				"position":    []int{560, 280},
				"parameters": map[string]interface{}{
					"respondWith":  "json",
					"responseBody": responseBody,
					"options":      map[string]interface{}{},
				},
			},
		},
		"connections": map[string]interface{}{
			"Webhook": map[string]interface{}{
				"main": []interface{}{
					[]interface{}{
						map[string]interface{}{
							"node":  "Respond to Webhook",
							"type":  "main",
							"index": 0,
						},
					},
				},
			},
		},
		"settings": map[string]interface{}{},
		"active":   false,
	}, webhookPath
}

func sanitizeN8NPath(in string) string {
	raw := strings.ToLower(strings.TrimSpace(in))
	if raw == "" {
		return ""
	}
	out := strings.Builder{}
	lastDash := false
	for _, ch := range raw {
		isAlphaNum := (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
		if isAlphaNum {
			out.WriteRune(ch)
			lastDash = false
			continue
		}
		if !lastDash {
			out.WriteRune('-')
			lastDash = true
		}
	}
	result := strings.Trim(out.String(), "-")
	if len(result) > 60 {
		result = result[:60]
		result = strings.Trim(result, "-")
	}
	return result
}

func (s *n8nService) createWorkflow(ctx context.Context, payload map[string]interface{}) (string, error) {
	endpoints := []string{"/api/v1/workflows", "/rest/workflows"}
	var lastErr error
	for _, endpoint := range endpoints {
		resp, err := s.requestJSON(ctx, http.MethodPost, endpoint, payload)
		if err != nil {
			lastErr = err
			continue
		}
		workflowID := workflowIDFromResponse(resp)
		if workflowID != "" {
			return workflowID, nil
		}
		lastErr = fmt.Errorf("n8n create workflow succeeded but no workflow id returned")
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("n8n create workflow failed")
	}
	return "", lastErr
}

func (s *n8nService) updateWorkflow(ctx context.Context, workflowID string, payload map[string]interface{}) error {
	id := url.PathEscape(strings.TrimSpace(workflowID))
	candidates := []struct {
		method   string
		endpoint string
	}{
		{method: http.MethodPatch, endpoint: "/api/v1/workflows/" + id},
		{method: http.MethodPut, endpoint: "/api/v1/workflows/" + id},
		{method: http.MethodPatch, endpoint: "/rest/workflows/" + id},
		{method: http.MethodPut, endpoint: "/rest/workflows/" + id},
	}
	var lastErr error
	for _, candidate := range candidates {
		_, err := s.requestJSON(ctx, candidate.method, candidate.endpoint, payload)
		if err == nil {
			return nil
		}
		lastErr = err
		if isN8NStatus(err, http.StatusNotFound) {
			continue
		}
		if isN8NStatus(err, http.StatusMethodNotAllowed) {
			continue
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("n8n update workflow failed")
	}
	return lastErr
}

func (s *n8nService) activateWorkflow(ctx context.Context, workflowID string) error {
	id := url.PathEscape(strings.TrimSpace(workflowID))
	candidates := []struct {
		method   string
		endpoint string
		body     map[string]interface{}
	}{
		{method: http.MethodPost, endpoint: "/api/v1/workflows/" + id + "/activate", body: nil},
		{method: http.MethodPost, endpoint: "/rest/workflows/" + id + "/activate", body: nil},
		{method: http.MethodPatch, endpoint: "/api/v1/workflows/" + id, body: map[string]interface{}{"active": true}},
		{method: http.MethodPatch, endpoint: "/rest/workflows/" + id, body: map[string]interface{}{"active": true}},
	}
	var lastErr error
	for _, candidate := range candidates {
		_, err := s.requestJSON(ctx, candidate.method, candidate.endpoint, candidate.body)
		if err == nil {
			return nil
		}
		lastErr = err
		if isN8NStatus(err, http.StatusNotFound, http.StatusMethodNotAllowed) {
			continue
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("n8n activate workflow failed")
	}
	return lastErr
}

func (s *n8nService) workflowURL(workflowID string) string {
	return s.baseURL + "/workflow/" + url.PathEscape(strings.TrimSpace(workflowID))
}

func workflowIDFromResponse(resp map[string]interface{}) string {
	if id := strings.TrimSpace(n8nString(resp["id"])); id != "" {
		return id
	}
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if id := strings.TrimSpace(n8nString(data["id"])); id != "" {
			return id
		}
	}
	if wf, ok := resp["workflow"].(map[string]interface{}); ok {
		if id := strings.TrimSpace(n8nString(wf["id"])); id != "" {
			return id
		}
	}
	return ""
}

func (s *n8nService) requestJSON(ctx context.Context, method, endpoint string, payload interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if s.apiKey != "" {
		req.Header.Set("X-N8N-API-KEY", s.apiKey)
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parsed := map[string]interface{}{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := strings.TrimSpace(stringFromAny(parsed["message"]))
		if msg == "" {
			msg = strings.TrimSpace(stringFromAny(parsed["error"]))
		}
		if msg == "" {
			msg = strings.TrimSpace(string(raw))
		}
		return nil, &n8nHTTPError{
			status:  resp.StatusCode,
			message: msg,
		}
	}

	return parsed, nil
}

func isN8NStatus(err error, statuses ...int) bool {
	n8nErr, ok := err.(*n8nHTTPError)
	if !ok {
		return false
	}
	for _, status := range statuses {
		if n8nErr.status == status {
			return true
		}
	}
	return false
}

func (s *n8nService) withRetry(ctx context.Context, fn func() error) error {
	const maxAttempts = 3
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryableN8NError(err) || attempt == maxAttempts {
			break
		}
		time.Sleep(time.Duration(attempt*200) * time.Millisecond)
	}
	return lastErr
}

func isRetryableN8NError(err error) bool {
	n8nErr, ok := err.(*n8nHTTPError)
	if !ok {
		// Network/transport errors are generally retryable.
		return true
	}
	if n8nErr.status == http.StatusTooManyRequests || n8nErr.status == http.StatusRequestTimeout {
		return true
	}
	return n8nErr.status >= 500
}

func n8nString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%.0f", val))
	case int:
		return strings.TrimSpace(fmt.Sprintf("%d", val))
	default:
		return ""
	}
}
