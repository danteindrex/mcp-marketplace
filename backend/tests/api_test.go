package tests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/yourorg/mcp-marketplace/backend/internal/auth"
	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	api "github.com/yourorg/mcp-marketplace/backend/internal/http"
	"github.com/yourorg/mcp-marketplace/backend/internal/store"
)

func newTestServer() http.Handler {
	cfg := config.Config{
		Port:               "8080",
		JWTSecret:          "test-secret",
		BaseURL:            "http://localhost:8080",
		SuperAdminEmail:    "admin@platform.local",
		SuperAdminPassword: "admin-pass",
	}
	st := store.NewMemoryStore(cfg)
	jwt := auth.NewJWTManager(cfg.JWTSecret)
	return api.NewRouter(cfg, st, jwt)
}

func signup(t *testing.T, h http.Handler, email, password, role string) map[string]interface{} {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{
		"email":      email,
		"password":   password,
		"name":       "Test User",
		"role":       role,
		"tenantName": role + " tenant",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("signup failed: %d %s", res.Code, res.Body.String())
	}
	var body map[string]interface{}
	_ = json.Unmarshal(res.Body.Bytes(), &body)
	return body
}

func login(t *testing.T, h http.Handler, email, password string) string {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", res.Code, res.Body.String())
	}
	var body map[string]interface{}
	_ = json.Unmarshal(res.Body.Bytes(), &body)
	return body["accessToken"].(string)
}

func call(t *testing.T, h http.Handler, method, path string, token string, payload interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if payload != nil {
		_ = json.NewEncoder(&buf).Encode(payload)
	}
	req := httptest.NewRequest(method, path, &buf)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}

type failingHealthStore struct {
	store.Store
}

func (f failingHealthStore) StoreType() string {
	return "mongo"
}

func (f failingHealthStore) Health(context.Context) error {
	return errors.New("mongo unavailable")
}

func TestPublicEndpoints(t *testing.T) {
	h := newTestServer()
	health := call(t, h, http.MethodGet, "/health", "", nil)
	if health.Code != http.StatusOK {
		t.Fatalf("health status %d", health.Code)
	}
	var healthBody map[string]interface{}
	_ = json.Unmarshal(health.Body.Bytes(), &healthBody)
	storage, _ := healthBody["storage"].(map[string]interface{})
	if storage == nil {
		t.Fatalf("health response missing storage details")
	}
	if storage["engine"] != "memory" {
		t.Fatalf("expected memory store health engine, got %v", storage["engine"])
	}
	if storage["connected"] != true {
		t.Fatalf("expected storage connected=true, got %v", storage["connected"])
	}
	if res := call(t, h, http.MethodGet, "/v1/marketplace/servers", "", nil); res.Code != http.StatusOK {
		t.Fatalf("marketplace list status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/marketplace/servers/postgresql-assistant", "", nil); res.Code != http.StatusNotFound {
		t.Fatalf("marketplace detail status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/.well-known/oauth-protected-resource", "", nil); res.Code != http.StatusOK {
		t.Fatalf("protected resource metadata status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/.well-known/oauth-authorization-server", "", nil); res.Code != http.StatusOK {
		t.Fatalf("authorization server metadata status %d", res.Code)
	}
}

func TestHealthFailsWhenRequiredStorageIsDown(t *testing.T) {
	cfg := config.Config{
		Port:               "8080",
		JWTSecret:          "test-secret",
		BaseURL:            "http://localhost:8080",
		SuperAdminEmail:    "admin@platform.local",
		SuperAdminPassword: "admin-pass",
		MongoRequired:      true,
	}
	baseStore := store.NewMemoryStore(cfg)
	st := failingHealthStore{Store: baseStore}
	jwt := auth.NewJWTManager(cfg.JWTSecret)
	h := api.NewRouter(cfg, st, jwt)

	res := call(t, h, http.MethodGet, "/health", "", nil)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected health 503 when required storage is unavailable, got %d", res.Code)
	}
	var body map[string]interface{}
	_ = json.Unmarshal(res.Body.Bytes(), &body)
	if body["status"] != "degraded" {
		t.Fatalf("expected degraded health status, got %v", body["status"])
	}
	storage, _ := body["storage"].(map[string]interface{})
	if storage == nil {
		t.Fatalf("health response missing storage payload")
	}
	if storage["connected"] != false {
		t.Fatalf("expected storage connected=false, got %v", storage["connected"])
	}
}

func TestBuyerFlowAndHub(t *testing.T) {
	h := newTestServer()
	signup(t, h, "buyer@acme.local", "BuyerPass123!@", "buyer")
	buyerToken := login(t, h, "buyer@acme.local", "BuyerPass123!@")

	if res := call(t, h, http.MethodGet, "/v1/buyer/entitlements", buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("entitlements status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/buyer/hub", buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("hub status %d", res.Code)
	}

	res := call(t, h, http.MethodPost, "/v1/buyer/connections", buyerToken, map[string]interface{}{
		"client": "vscode", "resource": "https://mcp.marketplace.local/hub/self", "grantedScopes": []string{"db:read"},
	})
	if res.Code != http.StatusCreated {
		t.Fatalf("create connection status %d body=%s", res.Code, res.Body.String())
	}

	if res := call(t, h, http.MethodGet, "/v1/buyer/connections", buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("list connections status %d", res.Code)
	}

	if res := call(t, h, http.MethodPost, "/v1/buyer/local-agents", buyerToken, map[string]interface{}{"deviceId": "device-1", "version": "0.1.0", "tunnelStatus": "connected"}); res.Code != http.StatusCreated {
		t.Fatalf("upsert local agent status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/buyer/local-agents", buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("list local agents status %d", res.Code)
	}
}

func TestMarketplaceInstallAndMCPHub(t *testing.T) {
	h := newTestServer()
	buyerSignup := signup(t, h, "install-buyer@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "install-merchant@acme.local", "MerchantPass123!@", "merchant")
	buyerToken := login(t, h, "install-buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "install-merchant@acme.local", "MerchantPass123!@")

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Installable Server",
		"slug":                 "installable-server",
		"description":          "Install flow test",
		"category":             "integration",
		"dockerImage":          "tenant/installable:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/installable-server",
		"requiredScopes":       []string{"db:read", "db:write"},
		"pricingType":          "free",
		"status":               "published",
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	serverID := createdBody["id"].(string)
	serverSlug := createdBody["slug"].(string)

	install := call(t, h, http.MethodPost, "/v1/marketplace/servers/"+serverSlug+"/install", buyerToken, map[string]interface{}{
		"client": "vscode",
	})
	if install.Code != http.StatusCreated {
		t.Fatalf("install endpoint status %d body=%s", install.Code, install.Body.String())
	}
	var installBody map[string]interface{}
	_ = json.Unmarshal(install.Body.Bytes(), &installBody)
	connection := installBody["connection"].(map[string]interface{})
	if connection["serverId"] != serverID {
		t.Fatalf("install connection must reference server id")
	}

	buyerUser := buyerSignup["user"].(map[string]interface{})
	hubPath := "/mcp/hub/" + buyerUser["tenantId"].(string) + "/" + buyerUser["id"].(string)

	if res := call(t, h, http.MethodGet, hubPath, buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("mcp hub get status %d body=%s", res.Code, res.Body.String())
	}

	res := call(t, h, http.MethodPost, hubPath, buyerToken, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "tools-list-1",
		"method":  "tools/list",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("mcp hub post status %d body=%s", res.Code, res.Body.String())
	}
}

func TestUserSettingsFlow(t *testing.T) {
	h := newTestServer()
	signup(t, h, "settings-buyer@acme.local", "BuyerPass123!@", "buyer")
	buyerToken := login(t, h, "settings-buyer@acme.local", "BuyerPass123!@")

	profile := call(t, h, http.MethodGet, "/v1/settings/profile", buyerToken, nil)
	if profile.Code != http.StatusOK {
		t.Fatalf("settings profile status %d body=%s", profile.Code, profile.Body.String())
	}

	prefs := call(t, h, http.MethodGet, "/v1/settings/preferences", buyerToken, nil)
	if prefs.Code != http.StatusOK {
		t.Fatalf("settings preferences status %d body=%s", prefs.Code, prefs.Body.String())
	}

	notifications := call(t, h, http.MethodGet, "/v1/settings/notifications", buyerToken, nil)
	if notifications.Code != http.StatusOK {
		t.Fatalf("settings notifications status %d body=%s", notifications.Code, notifications.Body.String())
	}

	profileUpdate := call(t, h, http.MethodPut, "/v1/settings/profile", buyerToken, map[string]interface{}{
		"name":      "Settings Buyer Updated",
		"email":     "settings-buyer@acme.local",
		"phone":     "+1-415-555-0101",
		"avatarUrl": "https://example.com/avatar.png",
		"locale":    "en-US",
		"timezone":  "America/New_York",
	})
	if profileUpdate.Code != http.StatusOK {
		t.Fatalf("update profile status %d body=%s", profileUpdate.Code, profileUpdate.Body.String())
	}

	prefsUpdate := call(t, h, http.MethodPut, "/v1/settings/preferences", buyerToken, map[string]interface{}{
		"theme":          "dark",
		"language":       "en",
		"timezone":       "America/New_York",
		"defaultLanding": "/buyer/dashboard",
		"compactMode":    true,
	})
	if prefsUpdate.Code != http.StatusOK {
		t.Fatalf("update preferences status %d body=%s", prefsUpdate.Code, prefsUpdate.Body.String())
	}

	notificationsUpdate := call(t, h, http.MethodPut, "/v1/settings/notifications", buyerToken, map[string]interface{}{
		"productUpdates": false,
		"securityAlerts": true,
		"billingAlerts":  true,
		"marketingEmail": false,
		"weeklyDigest":   false,
	})
	if notificationsUpdate.Code != http.StatusOK {
		t.Fatalf("update notifications status %d body=%s", notificationsUpdate.Code, notificationsUpdate.Body.String())
	}

	passwordChange := call(t, h, http.MethodPut, "/v1/settings/security/password", buyerToken, map[string]interface{}{
		"currentPassword": "BuyerPass123!@",
		"newPassword":     "BuyerPass456!@",
		"confirmPassword": "BuyerPass456!@",
	})
	if passwordChange.Code != http.StatusOK {
		t.Fatalf("change password status %d body=%s", passwordChange.Code, passwordChange.Body.String())
	}

	_ = login(t, h, "settings-buyer@acme.local", "BuyerPass456!@")
}

func TestMerchantAccessControl(t *testing.T) {
	h := newTestServer()
	signup(t, h, "buyer@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "merchant@dataflow.local", "MerchantPass123!@", "merchant")
	buyerToken := login(t, h, "buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "merchant@dataflow.local", "MerchantPass123!@")

	if res := call(t, h, http.MethodGet, "/v1/merchant/servers", buyerToken, nil); res.Code != http.StatusForbidden {
		t.Fatalf("buyer should be forbidden, got %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/merchant/servers", merchantToken, nil); res.Code != http.StatusOK {
		t.Fatalf("merchant list status %d", res.Code)
	}

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name": "GitHub Suite", "slug": "github-suite", "description": "GitHub tools", "category": "integration", "dockerImage": "acme/github-suite:1.0.0", "canonicalResourceUri": "https://mcp.marketplace.local/resource/srv_new", "requiredScopes": []string{"github:repos"}, "pricingType": "free", "supportsCloud": true, "supportsLocal": true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create server status %d body=%s", created.Code, created.Body.String())
	}

	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	id := createdBody["id"].(string)
	if res := call(t, h, http.MethodGet, "/v1/merchant/servers/"+id+"/pricing", merchantToken, nil); res.Code != http.StatusOK {
		t.Fatalf("pricing status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/merchant/servers/"+id+"/auth", merchantToken, nil); res.Code != http.StatusOK {
		t.Fatalf("auth status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/merchant/servers/"+id+"/observability", merchantToken, nil); res.Code != http.StatusOK {
		t.Fatalf("observability status %d", res.Code)
	}
}

func TestMerchantDeployThenPublishLifecycle(t *testing.T) {
	h := newTestServer()
	signup(t, h, "lifecycle-merchant@acme.local", "MerchantPass123!@", "merchant")
	merchantToken := login(t, h, "lifecycle-merchant@acme.local", "MerchantPass123!@")

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Lifecycle Server",
		"slug":                 "lifecycle-server",
		"description":          "Lifecycle flow test",
		"category":             "automation",
		"dockerImage":          "tenant/lifecycle:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/lifecycle-server",
		"requiredScopes":       []string{"db:read"},
		"pricingType":          "x402",
		"pricingAmount":        0.0,
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	serverID := createdBody["id"].(string)
	serverSlug := createdBody["slug"].(string)

	publishBeforeDeploy := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverID+"/publish", merchantToken, nil)
	if publishBeforeDeploy.Code != http.StatusConflict {
		t.Fatalf("expected publish before deploy to fail with 409, got %d body=%s", publishBeforeDeploy.Code, publishBeforeDeploy.Body.String())
	}

	deploy := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverID+"/deploy", merchantToken, map[string]interface{}{
		"deploymentTarget": "us-west-1",
		"n8nWorkflowId":    "wf_123",
	})
	if deploy.Code != http.StatusOK {
		t.Fatalf("deploy status %d body=%s", deploy.Code, deploy.Body.String())
	}

	publishWithoutPrice := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverID+"/publish", merchantToken, nil)
	if publishWithoutPrice.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected publish without price to fail with 422, got %d body=%s", publishWithoutPrice.Code, publishWithoutPrice.Body.String())
	}

	updatePrice := call(t, h, http.MethodPut, "/v1/merchant/servers/"+serverID, merchantToken, map[string]interface{}{
		"pricingType":   "x402",
		"pricingAmount": 0.25,
	})
	if updatePrice.Code != http.StatusOK {
		t.Fatalf("update price status %d body=%s", updatePrice.Code, updatePrice.Body.String())
	}

	publish := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverID+"/publish", merchantToken, nil)
	if publish.Code != http.StatusOK {
		t.Fatalf("publish status %d body=%s", publish.Code, publish.Body.String())
	}

	if res := call(t, h, http.MethodGet, "/v1/marketplace/servers/"+serverSlug, "", nil); res.Code != http.StatusOK {
		t.Fatalf("marketplace detail should be available after publish, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestMerchantDeploySyncsWithN8NWhenConfigured(t *testing.T) {
	var createCalls int32
	var activateCalls int32

	n8n := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workflows":
			atomic.AddInt32(&createCalls, 1)
			write := map[string]interface{}{"id": "wf_test_123"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(write)
			return
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workflows/wf_test_123/activate":
			atomic.AddInt32(&activateCalls, 1)
			write := map[string]interface{}{"id": "wf_test_123", "active": true}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(write)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
			return
		}
	}))
	defer n8n.Close()

	cfg := config.Config{
		Port:               "8080",
		JWTSecret:          "test-secret",
		BaseURL:            "http://localhost:8080",
		N8NBaseURL:         n8n.URL,
		SuperAdminEmail:    "admin@platform.local",
		SuperAdminPassword: "admin-pass",
	}
	st := store.NewMemoryStore(cfg)
	jwt := auth.NewJWTManager(cfg.JWTSecret)
	h := api.NewRouter(cfg, st, jwt)

	signup(t, h, "n8n-merchant@acme.local", "MerchantPass123!@", "merchant")
	merchantToken := login(t, h, "n8n-merchant@acme.local", "MerchantPass123!@")

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "N8N Deploy Server",
		"slug":                 "n8n-deploy-server",
		"description":          "N8N deploy test",
		"category":             "automation",
		"dockerImage":          "tenant/n8n-deploy:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/n8n-deploy-server",
		"requiredScopes":       []string{"agent:invoke"},
		"pricingType":          "x402",
		"pricingAmount":        0.0,
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	serverID := createdBody["id"].(string)

	deploy := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverID+"/deploy", merchantToken, nil)
	if deploy.Code != http.StatusOK {
		t.Fatalf("deploy status %d body=%s", deploy.Code, deploy.Body.String())
	}

	var deployBody map[string]interface{}
	_ = json.Unmarshal(deploy.Body.Bytes(), &deployBody)
	serverObj := deployBody["server"].(map[string]interface{})
	if strings.TrimSpace(serverObj["n8nWorkflowId"].(string)) != "wf_test_123" {
		t.Fatalf("expected n8n workflow id to be persisted, got body=%s", deploy.Body.String())
	}
	if atomic.LoadInt32(&createCalls) == 0 || atomic.LoadInt32(&activateCalls) == 0 {
		t.Fatalf("expected n8n create+activate calls, got create=%d activate=%d", createCalls, activateCalls)
	}
}

func TestAdminAndX402(t *testing.T) {
	h := newTestServer()
	buyerSignup := signup(t, h, "buyer@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "merchant@acme.local", "MerchantPass123!@", "merchant")
	adminToken := login(t, h, "admin@platform.local", "admin-pass")
	buyerToken := login(t, h, "buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "merchant@acme.local", "MerchantPass123!@")
	buyerUser := buyerSignup["user"].(map[string]interface{})

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Doc Extractor",
		"slug":                 "doc-extractor",
		"description":          "Extract text",
		"category":             "ai",
		"dockerImage":          "local/doc-extractor:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/doc-extractor",
		"requiredScopes":       []string{"documents:read"},
		"pricingType":          "x402",
		"pricingAmount":        0.05,
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create merchant server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	serverID := createdBody["id"].(string)

	if res := call(t, h, http.MethodGet, "/v1/admin/tenants", adminToken, nil); res.Code != http.StatusOK {
		t.Fatalf("admin tenants status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/admin/security-events", adminToken, nil); res.Code != http.StatusOK {
		t.Fatalf("admin security status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/admin/audit-logs", adminToken, nil); res.Code != http.StatusOK {
		t.Fatalf("admin audit status %d", res.Code)
	}

	grant := call(t, h, http.MethodPost, "/v1/admin/entitlements", adminToken, map[string]interface{}{
		"tenantId": buyerUser["tenantId"], "userId": buyerUser["id"], "serverId": serverID, "allowedScopes": []string{"documents:read"}, "cloudAllowed": true, "localAllowed": true,
	})
	if grant.Code != http.StatusCreated {
		t.Fatalf("grant entitlement status %d body=%s", grant.Code, grant.Body.String())
	}

	intent := call(t, h, http.MethodPost, "/v1/billing/x402/intents", buyerToken, map[string]interface{}{
		"serverId": serverID, "toolName": "extract_pdf", "amount": 0.05,
	})
	if intent.Code != http.StatusPaymentRequired {
		t.Fatalf("x402 intent status %d body=%s", intent.Code, intent.Body.String())
	}
	if got := intent.Header().Get("PAYMENT-REQUIRED"); got == "" {
		t.Fatalf("missing PAYMENT-REQUIRED header")
	}

	var ibody map[string]interface{}
	_ = json.Unmarshal(intent.Body.Bytes(), &ibody)
	intentObj := ibody["intent"].(map[string]interface{})
	intentID := intentObj["id"].(string)
	if res := call(t, h, http.MethodPost, "/v1/billing/x402/intents/"+intentID+"/settle", buyerToken, map[string]interface{}{
		"paymentResponse": map[string]interface{}{"paymentIdentifier": "pay_admin_x402_1", "method": "x402_wallet"},
	}); res.Code != http.StatusOK {
		t.Fatalf("settle status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/billing/x402/intents", buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("list intents status %d", res.Code)
	}
}

func TestOAuthAuthorizationCodePKCEFlow(t *testing.T) {
	h := newTestServer()
	buyerSignup := signup(t, h, "oauth-buyer@acme.local", "BuyerPass123!@", "buyer")
	buyerToken := login(t, h, "oauth-buyer@acme.local", "BuyerPass123!@")
	buyerUser := buyerSignup["user"].(map[string]interface{})
	resource := "https://mcp.marketplace.local/hub/" + buyerUser["tenantId"].(string) + "/" + buyerUser["id"].(string)

	dcr := call(t, h, http.MethodPost, "/oauth/register", "", map[string]interface{}{
		"client_name":                "Test Client",
		"redirect_uris":              []string{"http://127.0.0.1:33418"},
		"grant_types":                []string{"authorization_code"},
		"token_endpoint_auth_method": "none",
	})
	if dcr.Code != http.StatusCreated {
		t.Fatalf("dcr status %d body=%s", dcr.Code, dcr.Body.String())
	}
	var dcrBody map[string]interface{}
	_ = json.Unmarshal(dcr.Body.Bytes(), &dcrBody)
	clientID := dcrBody["client_id"].(string)

	verifier := "test-verifier-1234567890"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	authorizePath := "/oauth/authorize?response_type=code&client_id=" + clientID + "&redirect_uri=http://127.0.0.1:33418&state=abc123&resource=" + resource + "&scope=db:read%20db:write&code_challenge=" + challenge + "&code_challenge_method=S256"
	authRes := call(t, h, http.MethodGet, authorizePath, buyerToken, nil)
	if authRes.Code != http.StatusOK {
		t.Fatalf("authorize status %d body=%s", authRes.Code, authRes.Body.String())
	}
	var authBody map[string]interface{}
	_ = json.Unmarshal(authRes.Body.Bytes(), &authBody)
	code := authBody["code"].(string)

	form := "grant_type=authorization_code&client_id=" + clientID + "&code=" + code + "&redirect_uri=http://127.0.0.1:33418&code_verifier=" + verifier + "&resource=" + resource
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("token status %d body=%s", res.Code, res.Body.String())
	}
}

func TestOAuthAuthorizeRejectsResourceImpersonation(t *testing.T) {
	h := newTestServer()
	userA := signup(t, h, "oauth-a@acme.local", "BuyerPass123!@", "buyer")
	userB := signup(t, h, "oauth-b@acme.local", "BuyerPass123!@", "buyer")
	tokenA := login(t, h, "oauth-a@acme.local", "BuyerPass123!@")

	dcr := call(t, h, http.MethodPost, "/oauth/register", "", map[string]interface{}{
		"client_name":                "Impersonation Test Client",
		"redirect_uris":              []string{"http://127.0.0.1:33418"},
		"grant_types":                []string{"authorization_code"},
		"token_endpoint_auth_method": "none",
	})
	if dcr.Code != http.StatusCreated {
		t.Fatalf("dcr status %d body=%s", dcr.Code, dcr.Body.String())
	}
	var dcrBody map[string]interface{}
	_ = json.Unmarshal(dcr.Body.Bytes(), &dcrBody)
	clientID := dcrBody["client_id"].(string)

	userBMeta := userB["user"].(map[string]interface{})
	resourceB := "https://mcp.marketplace.local/hub/" + userBMeta["tenantId"].(string) + "/" + userBMeta["id"].(string)
	_ = userA // kept to make intent explicit for test setup
	verifier := "impersonation-verifier-1234567890"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	authorizePath := "/oauth/authorize?response_type=code&client_id=" + clientID + "&redirect_uri=http://127.0.0.1:33418&state=abc123&resource=" + resourceB + "&scope=db:read&code_challenge=" + challenge + "&code_challenge_method=S256"
	authRes := call(t, h, http.MethodGet, authorizePath, tokenA, nil)
	if authRes.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden for impersonation attempt, got %d body=%s", authRes.Code, authRes.Body.String())
	}
}

func TestMerchantServerReadEndpointsEnforceTenantIsolation(t *testing.T) {
	h := newTestServer()
	signup(t, h, "merchant-one@acme.local", "MerchantPass123!@", "merchant")
	signup(t, h, "merchant-two@acme.local", "MerchantPass123!@", "merchant")
	merchantOneToken := login(t, h, "merchant-one@acme.local", "MerchantPass123!@")
	merchantTwoToken := login(t, h, "merchant-two@acme.local", "MerchantPass123!@")
	adminToken := login(t, h, "admin@platform.local", "admin-pass")

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantOneToken, map[string]interface{}{
		"name":                 "Tenant One Server",
		"slug":                 "tenant-one-server",
		"description":          "Tenant one only",
		"category":             "integration",
		"dockerImage":          "tenant/one-server:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/tenant-one-server",
		"requiredScopes":       []string{"db:read"},
		"pricingType":          "free",
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	id := createdBody["id"].(string)

	paths := []string{
		"/v1/merchant/servers/" + id + "/auth",
		"/v1/merchant/servers/" + id + "/pricing",
		"/v1/merchant/servers/" + id + "/observability",
		"/v1/merchant/servers/" + id + "/deployments",
		"/v1/merchant/servers/" + id + "/builder",
	}
	for _, path := range paths {
		if res := call(t, h, http.MethodGet, path, merchantTwoToken, nil); res.Code != http.StatusForbidden {
			t.Fatalf("expected forbidden for %s, got %d body=%s", path, res.Code, res.Body.String())
		}
		if res := call(t, h, http.MethodGet, path, adminToken, nil); res.Code != http.StatusOK {
			t.Fatalf("expected admin success for %s, got %d body=%s", path, res.Code, res.Body.String())
		}
	}
}

func TestSettleX402IntentRequiresOwner(t *testing.T) {
	h := newTestServer()
	buyerOneSignup := signup(t, h, "buyer-one@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "buyer-two@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "merchant-pay@acme.local", "MerchantPass123!@", "merchant")
	adminToken := login(t, h, "admin@platform.local", "admin-pass")
	buyerOneToken := login(t, h, "buyer-one@acme.local", "BuyerPass123!@")
	buyerTwoToken := login(t, h, "buyer-two@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "merchant-pay@acme.local", "MerchantPass123!@")
	buyerOneUser := buyerOneSignup["user"].(map[string]interface{})

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Chargeable Server",
		"slug":                 "chargeable-server",
		"description":          "Chargeable tools",
		"category":             "ai",
		"dockerImage":          "tenant/chargeable:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/chargeable-server",
		"requiredScopes":       []string{"documents:read"},
		"pricingType":          "x402",
		"pricingAmount":        0.05,
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	serverID := createdBody["id"].(string)

	grant := call(t, h, http.MethodPost, "/v1/admin/entitlements", adminToken, map[string]interface{}{
		"tenantId": buyerOneUser["tenantId"], "userId": buyerOneUser["id"], "serverId": serverID, "allowedScopes": []string{"documents:read"}, "cloudAllowed": true, "localAllowed": true,
	})
	if grant.Code != http.StatusCreated {
		t.Fatalf("grant entitlement status %d body=%s", grant.Code, grant.Body.String())
	}

	intent := call(t, h, http.MethodPost, "/v1/billing/x402/intents", buyerOneToken, map[string]interface{}{
		"serverId": serverID, "toolName": "extract_pdf", "amount": 0.05,
	})
	if intent.Code != http.StatusPaymentRequired {
		t.Fatalf("x402 intent status %d body=%s", intent.Code, intent.Body.String())
	}
	var intentBody map[string]interface{}
	_ = json.Unmarshal(intent.Body.Bytes(), &intentBody)
	intentObj := intentBody["intent"].(map[string]interface{})
	intentID := intentObj["id"].(string)

	if res := call(t, h, http.MethodPost, "/v1/billing/x402/intents/"+intentID+"/settle", buyerTwoToken, nil); res.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden for non-owner settle, got %d body=%s", res.Code, res.Body.String())
	}
	if res := call(t, h, http.MethodPost, "/v1/billing/x402/intents/"+intentID+"/settle", buyerOneToken, map[string]interface{}{
		"paymentResponse": map[string]interface{}{"paymentIdentifier": "pay_owner_x402_1", "method": "x402_wallet"},
	}); res.Code != http.StatusOK {
		t.Fatalf("expected owner settle success, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestMCPToolsCallX402MetaFlow(t *testing.T) {
	h := newTestServer()
	buyerSignup := signup(t, h, "mcp-buyer@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "mcp-merchant@acme.local", "MerchantPass123!@", "merchant")
	adminToken := login(t, h, "admin@platform.local", "admin-pass")
	buyerToken := login(t, h, "mcp-buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "mcp-merchant@acme.local", "MerchantPass123!@")
	buyerUser := buyerSignup["user"].(map[string]interface{})

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "MCP Paid Tool",
		"slug":                 "mcp-paid-tool",
		"description":          "paid tool",
		"category":             "ai",
		"dockerImage":          "tenant/mcp-paid:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/mcp-paid-tool",
		"requiredScopes":       []string{"documents:read"},
		"pricingType":          "x402",
		"pricingAmount":        0.05,
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create merchant server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	serverID := createdBody["id"].(string)
	serverSlug := createdBody["slug"].(string)

	grant := call(t, h, http.MethodPost, "/v1/admin/entitlements", adminToken, map[string]interface{}{
		"tenantId": buyerUser["tenantId"], "userId": buyerUser["id"], "serverId": serverID, "allowedScopes": []string{"documents:read"}, "cloudAllowed": true, "localAllowed": true,
	})
	if grant.Code != http.StatusCreated {
		t.Fatalf("grant entitlement status %d body=%s", grant.Code, grant.Body.String())
	}

	hubPath := "/mcp/hub/" + buyerUser["tenantId"].(string) + "/" + buyerUser["id"].(string)
	toolName := "invoke_" + strings.ReplaceAll(serverSlug, "-", "_")

	needPay := call(t, h, http.MethodPost, hubPath, buyerToken, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "mcp-pay-1",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{"query": "hello"},
		},
	})
	if needPay.Code != http.StatusOK {
		t.Fatalf("expected jsonrpc envelope status 200, got %d body=%s", needPay.Code, needPay.Body.String())
	}
	var needPayBody map[string]interface{}
	_ = json.Unmarshal(needPay.Body.Bytes(), &needPayBody)
	errObj, _ := needPayBody["error"].(map[string]interface{})
	if int(errObj["code"].(float64)) != -32002 {
		t.Fatalf("expected payment required error -32002, got %v body=%s", errObj["code"], needPay.Body.String())
	}

	paid := call(t, h, http.MethodPost, hubPath, buyerToken, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "mcp-pay-2",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{"query": "hello"},
			"_meta": map[string]interface{}{
				"x402/payment-response": map[string]interface{}{
					"paymentIdentifier": "pay_meta_1",
					"method":            "x402_wallet",
				},
			},
		},
	})
	if paid.Code != http.StatusOK {
		t.Fatalf("expected jsonrpc envelope status 200, got %d body=%s", paid.Code, paid.Body.String())
	}
	var paidBody map[string]interface{}
	_ = json.Unmarshal(paid.Body.Bytes(), &paidBody)
	if paidBody["error"] != nil {
		t.Fatalf("expected paid tools/call success, got error body=%s", paid.Body.String())
	}

	replay := call(t, h, http.MethodPost, hubPath, buyerToken, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "mcp-pay-3",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{"query": "hello"},
			"_meta": map[string]interface{}{
				"x402/payment-response": map[string]interface{}{
					"paymentIdentifier": "pay_meta_1",
					"method":            "x402_wallet",
				},
			},
		},
	})
	if replay.Code != http.StatusOK {
		t.Fatalf("expected jsonrpc envelope status 200, got %d body=%s", replay.Code, replay.Body.String())
	}
	var replayBody map[string]interface{}
	_ = json.Unmarshal(replay.Body.Bytes(), &replayBody)
	replayErr, _ := replayBody["error"].(map[string]interface{})
	if replayErr == nil || int(replayErr["code"].(float64)) != -32002 {
		t.Fatalf("expected replay to be rejected with -32002, got body=%s", replay.Body.String())
	}
}

func TestPaymentsControlEndpoints(t *testing.T) {
	h := newTestServer()
	signup(t, h, "controls-buyer@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "controls-merchant@acme.local", "MerchantPass123!@", "merchant")
	buyerToken := login(t, h, "controls-buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "controls-merchant@acme.local", "MerchantPass123!@")
	adminToken := login(t, h, "admin@platform.local", "admin-pass")

	updated := call(t, h, http.MethodPut, "/v1/buyer/payments/controls", buyerToken, map[string]interface{}{
		"monthlySpendCapUsdc": 50.0,
		"dailySpendCapUsdc":   10.0,
		"perCallCapUsdc":      1.0,
		"allowedMethods":      []string{"x402_wallet"},
		"siwxWallet":          "0xabc",
	})
	if updated.Code != http.StatusOK {
		t.Fatalf("buyer payment controls update status %d body=%s", updated.Code, updated.Body.String())
	}
	if res := call(t, h, http.MethodGet, "/v1/buyer/payments/controls", buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("buyer payment controls get status %d body=%s", res.Code, res.Body.String())
	}

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Control Server",
		"slug":                 "control-server",
		"description":          "Control server",
		"category":             "integration",
		"dockerImage":          "tenant/control-server:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/control-server",
		"requiredScopes":       []string{"documents:read"},
		"pricingType":          "x402",
		"pricingAmount":        0.1,
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create merchant server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	serverID := createdBody["id"].(string)

	cfg := call(t, h, http.MethodPut, "/v1/merchant/servers/"+serverID+"/payments/config", merchantToken, map[string]interface{}{
		"paymentAddress": "0xmerchant",
		"perCallCapUsdc": 2.0,
		"dailyCapUsdc":   20.0,
		"monthlyCapUsdc": 200.0,
		"paymentMethods": []string{"x402_wallet", "stripe"},
	})
	if cfg.Code != http.StatusOK {
		t.Fatalf("merchant payment config update status %d body=%s", cfg.Code, cfg.Body.String())
	}
	if res := call(t, h, http.MethodGet, "/v1/merchant/payments/overview", merchantToken, nil); res.Code != http.StatusOK {
		t.Fatalf("merchant payments overview status %d body=%s", res.Code, res.Body.String())
	}
	if res := call(t, h, http.MethodGet, "/v1/admin/payments/overview", adminToken, nil); res.Code != http.StatusOK {
		t.Fatalf("admin payments overview status %d body=%s", res.Code, res.Body.String())
	}
}

func TestBuyerWalletTopUpAndWalletBalanceSettle(t *testing.T) {
	h := newTestServer()
	signup(t, h, "wallet-buyer@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "wallet-merchant@acme.local", "MerchantPass123!@", "merchant")
	buyerToken := login(t, h, "wallet-buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "wallet-merchant@acme.local", "MerchantPass123!@")

	controls := call(t, h, http.MethodPut, "/v1/buyer/payments/controls", buyerToken, map[string]interface{}{
		"allowedMethods":     []string{"wallet_balance"},
		"minimumBalanceUsdc": 0.0,
		"hardStopOnLowFunds": false,
		"walletAddress":      "0xwalletbuyer",
	})
	if controls.Code != http.StatusOK {
		t.Fatalf("update wallet controls status %d body=%s", controls.Code, controls.Body.String())
	}

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Wallet Billable",
		"slug":                 "wallet-billable",
		"description":          "Wallet billed tool",
		"category":             "ai",
		"dockerImage":          "tenant/wallet-billable:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/wallet-billable",
		"requiredScopes":       []string{"documents:read"},
		"pricingType":          "x402",
		"pricingAmount":        0.05,
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create merchant server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	serverID := createdBody["id"].(string)

	topup := call(t, h, http.MethodPost, "/v1/buyer/payments/topups/stripe/session", buyerToken, map[string]interface{}{
		"amountUsd":     25.0,
		"walletAddress": "0xwalletbuyer",
	})
	if topup.Code != http.StatusCreated {
		t.Fatalf("create stripe topup session status %d body=%s", topup.Code, topup.Body.String())
	}
	var topupBody map[string]interface{}
	_ = json.Unmarshal(topup.Body.Bytes(), &topupBody)
	topupObj := topupBody["topup"].(map[string]interface{})
	providerSessionID := topupObj["providerSessionId"].(string)

	webhook := call(t, h, http.MethodPost, "/webhooks/stripe/onramp", "", map[string]interface{}{
		"id":   "evt_test_wallet_1",
		"type": "crypto.onramp_session_updated",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":                 providerSessionID,
				"status":             "fulfillment_complete",
				"destination_amount": 25.0,
			},
		},
	})
	if webhook.Code != http.StatusOK {
		t.Fatalf("stripe webhook status %d body=%s", webhook.Code, webhook.Body.String())
	}

	afterCredit := call(t, h, http.MethodGet, "/v1/buyer/payments/controls", buyerToken, nil)
	if afterCredit.Code != http.StatusOK {
		t.Fatalf("payment controls after credit status %d body=%s", afterCredit.Code, afterCredit.Body.String())
	}
	var afterCreditBody map[string]interface{}
	_ = json.Unmarshal(afterCredit.Body.Bytes(), &afterCreditBody)
	walletInfo := afterCreditBody["wallet"].(map[string]interface{})
	if walletInfo["balanceUsdc"].(float64) < 25.0 {
		t.Fatalf("expected wallet balance to be credited, got body=%s", afterCredit.Body.String())
	}

	intent := call(t, h, http.MethodPost, "/v1/billing/x402/intents", buyerToken, map[string]interface{}{
		"serverId":       serverID,
		"toolName":       "wallet_tool",
		"paymentMethod":  "wallet_balance",
		"idempotencyKey": "wallet_intent_1",
	})
	if intent.Code != http.StatusPaymentRequired {
		t.Fatalf("wallet x402 intent status %d body=%s", intent.Code, intent.Body.String())
	}
	var intentBody map[string]interface{}
	_ = json.Unmarshal(intent.Body.Bytes(), &intentBody)
	intentObj := intentBody["intent"].(map[string]interface{})
	intentID := intentObj["id"].(string)

	settled := call(t, h, http.MethodPost, "/v1/billing/x402/intents/"+intentID+"/settle", buyerToken, nil)
	if settled.Code != http.StatusOK {
		t.Fatalf("wallet settle status %d body=%s", settled.Code, settled.Body.String())
	}

	afterSpend := call(t, h, http.MethodGet, "/v1/buyer/payments/controls", buyerToken, nil)
	if afterSpend.Code != http.StatusOK {
		t.Fatalf("payment controls after spend status %d body=%s", afterSpend.Code, afterSpend.Body.String())
	}
	var afterSpendBody map[string]interface{}
	_ = json.Unmarshal(afterSpend.Body.Bytes(), &afterSpendBody)
	walletAfterSpend := afterSpendBody["wallet"].(map[string]interface{})
	if walletAfterSpend["balanceUsdc"].(float64) >= 25.0 {
		t.Fatalf("expected wallet balance to decrease after settle, got body=%s", afterSpend.Body.String())
	}
}

func TestFeePolicyPayoutProfileAndPayoutRun(t *testing.T) {
	h := newTestServer()
	buyerSignup := signup(t, h, "payout-buyer@acme.local", "BuyerPass123!@", "buyer")
	merchantSignup := signup(t, h, "payout-merchant@acme.local", "MerchantPass123!@", "merchant")
	adminToken := login(t, h, "admin@platform.local", "admin-pass")
	buyerToken := login(t, h, "payout-buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "payout-merchant@acme.local", "MerchantPass123!@")
	buyerUser := buyerSignup["user"].(map[string]interface{})
	merchantUser := merchantSignup["user"].(map[string]interface{})

	feePolicy := call(t, h, http.MethodPut, "/v1/admin/payments/fee-policies", adminToken, map[string]interface{}{
		"scope":          "global",
		"platformFeeBps": 1000,
		"enabled":        true,
		"holdDays":       0,
		"payoutCadence":  "manual",
	})
	if feePolicy.Code != http.StatusOK {
		t.Fatalf("fee policy update status %d body=%s", feePolicy.Code, feePolicy.Body.String())
	}

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Payout Test Server",
		"slug":                 "payout-test-server",
		"description":          "Payout test",
		"category":             "ai",
		"dockerImage":          "tenant/payout-test:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/payout-test-server",
		"requiredScopes":       []string{"documents:read"},
		"pricingType":          "x402",
		"pricingAmount":        1.0,
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create merchant server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	serverID := createdBody["id"].(string)

	grant := call(t, h, http.MethodPost, "/v1/admin/entitlements", adminToken, map[string]interface{}{
		"tenantId": buyerUser["tenantId"], "userId": buyerUser["id"], "serverId": serverID, "allowedScopes": []string{"documents:read"}, "cloudAllowed": true, "localAllowed": true,
	})
	if grant.Code != http.StatusCreated {
		t.Fatalf("grant entitlement status %d body=%s", grant.Code, grant.Body.String())
	}

	intent := call(t, h, http.MethodPost, "/v1/billing/x402/intents", buyerToken, map[string]interface{}{
		"serverId": serverID, "toolName": "run", "amount": 1.0,
	})
	if intent.Code != http.StatusPaymentRequired {
		t.Fatalf("x402 intent status %d body=%s", intent.Code, intent.Body.String())
	}
	var intentBody map[string]interface{}
	_ = json.Unmarshal(intent.Body.Bytes(), &intentBody)
	intentObj := intentBody["intent"].(map[string]interface{})
	intentID := intentObj["id"].(string)

	settle := call(t, h, http.MethodPost, "/v1/billing/x402/intents/"+intentID+"/settle", buyerToken, map[string]interface{}{
		"paymentResponse": map[string]interface{}{"paymentIdentifier": "pay_payout_test_1", "method": "x402_wallet"},
	})
	if settle.Code != http.StatusOK {
		t.Fatalf("settle status %d body=%s", settle.Code, settle.Body.String())
	}

	updateProfile := call(t, h, http.MethodPut, "/v1/merchant/payments/payout-profile", merchantToken, map[string]interface{}{
		"preferredMethod":   "stablecoin",
		"stablecoinAddress": "0xmerchantwallet",
		"minPayoutUsdc":     0,
		"holdDays":          0,
	})
	if updateProfile.Code != http.StatusOK {
		t.Fatalf("merchant payout profile update status %d body=%s", updateProfile.Code, updateProfile.Body.String())
	}

	run := call(t, h, http.MethodPost, "/v1/admin/payments/payouts/run", adminToken, map[string]interface{}{
		"tenantId": merchantUser["tenantId"],
		"method":   "stablecoin",
		"force":    true,
	})
	if run.Code != http.StatusOK {
		t.Fatalf("run payout status %d body=%s", run.Code, run.Body.String())
	}
	var runBody map[string]interface{}
	_ = json.Unmarshal(run.Body.Bytes(), &runBody)
	record, _ := runBody["record"].(map[string]interface{})
	if record["status"] != "completed" {
		t.Fatalf("expected payout record completed, got body=%s", run.Body.String())
	}

	merchantPayouts := call(t, h, http.MethodGet, "/v1/merchant/payments/payouts", merchantToken, nil)
	if merchantPayouts.Code != http.StatusOK {
		t.Fatalf("merchant payouts status %d body=%s", merchantPayouts.Code, merchantPayouts.Body.String())
	}
}
