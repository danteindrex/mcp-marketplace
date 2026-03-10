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
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/auth"
	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	api "github.com/yourorg/mcp-marketplace/backend/internal/http"
	storemodels "github.com/yourorg/mcp-marketplace/backend/internal/models"
	"github.com/yourorg/mcp-marketplace/backend/internal/store"
)

func newTestServer() http.Handler {
	return newTestServerWithConfig(config.Config{
		AllowInsecureDefaults: true,
		X402Mode:              "disabled",
	})
}

func newTestServerWithConfig(cfg config.Config) http.Handler {
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}
	if cfg.SuperAdminEmail == "" {
		cfg.SuperAdminEmail = "admin@platform.local"
	}
	if cfg.SuperAdminPassword == "" {
		cfg.SuperAdminPassword = "admin-pass"
	}
	st := store.NewMemoryStore(cfg)
	jwt, err := auth.NewJWTManager(cfg)
	if err != nil {
		panic(err)
	}
	return api.NewRouter(cfg, st, jwt)
}

func newTestServerWithStore(cfg config.Config, st store.Store) http.Handler {
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}
	if cfg.SuperAdminEmail == "" {
		cfg.SuperAdminEmail = "admin@platform.local"
	}
	if cfg.SuperAdminPassword == "" {
		cfg.SuperAdminPassword = "admin-pass"
	}
	jwt, err := auth.NewJWTManager(cfg)
	if err != nil {
		panic(err)
	}
	return api.NewRouter(cfg, st, jwt)
}

func oauthAccessToken(t *testing.T, h http.Handler, appToken string, tenantID string, userID string, scopes []string) string {
	t.Helper()
	resource := "http://localhost:8080/mcp/hub/" + tenantID + "/" + userID
	dcr := call(t, h, http.MethodPost, "/oauth/register", "", map[string]interface{}{
		"client_name":                "Test OAuth Client",
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
	authorizePath := "/oauth/authorize?response_type=code&client_id=" + clientID + "&redirect_uri=http://127.0.0.1:33418&state=abc123&resource=" + resource + "&scope=" + strings.Join(scopes, "%20") + "&code_challenge=" + challenge + "&code_challenge_method=S256"
	authRes := call(t, h, http.MethodGet, authorizePath, appToken, nil)
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
	var tokenBody map[string]interface{}
	_ = json.Unmarshal(res.Body.Bytes(), &tokenBody)
	return tokenBody["access_token"].(string)
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

func blockingReasonCodes(t *testing.T, res *httptest.ResponseRecorder) []string {
	t.Helper()
	var body map[string]interface{}
	_ = json.Unmarshal(res.Body.Bytes(), &body)
	items, _ := body["blockingReasons"].([]interface{})
	codes := make([]string, 0, len(items))
	for _, item := range items {
		reason, _ := item.(map[string]interface{})
		code, _ := reason["code"].(string)
		if code != "" {
			codes = append(codes, code)
		}
	}
	return codes
}

func storeServerFixture(name, slug string, pricingAmount float64, paymentMethods []string) storemodels.Server {
	now := time.Now().UTC()
	return storemodels.Server{
		TenantID:             "tenant_fixture",
		Author:               "Fixture Merchant",
		Name:                 name,
		Slug:                 slug,
		Description:          "fixture",
		Category:             "integration",
		Version:              "1.0.0",
		DockerImage:          "tenant/fixture:1.0.0",
		CanonicalResourceURI: "https://mcp.marketplace.local/resource/" + slug,
		RequiredScopes:       []string{"db:read"},
		PricingType:          "x402",
		PricingAmount:        pricingAmount,
		SupportsCloud:        true,
		SupportsLocal:        true,
		PaymentMethods:       paymentMethods,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
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
	if res := call(t, h, http.MethodGet, "/.well-known/oauth-protected-resource?resource=http://localhost:8080/mcp/hub/tenant-a/user-a/", "", nil); res.Code != http.StatusOK {
		t.Fatalf("protected resource metadata canonical status %d body=%s", res.Code, res.Body.String())
	} else {
		var body map[string]interface{}
		_ = json.Unmarshal(res.Body.Bytes(), &body)
		if body["resource"] != "http://localhost:8080/mcp/hub/tenant-a/user-a" {
			t.Fatalf("expected canonical protected resource, got %v", body["resource"])
		}
	}
	if res := call(t, h, http.MethodGet, "/.well-known/oauth-authorization-server", "", nil); res.Code != http.StatusOK {
		t.Fatalf("authorization server metadata status %d", res.Code)
	}
}

func TestHealthFailsWhenRequiredStorageIsDown(t *testing.T) {
	cfg := config.Config{
		Port:                  "8080",
		BaseURL:               "http://localhost:8080",
		SuperAdminEmail:       "admin@platform.local",
		SuperAdminPassword:    "admin-pass",
		AllowInsecureDefaults: true,
		MongoRequired:         true,
	}
	baseStore := store.NewMemoryStore(cfg)
	st := failingHealthStore{Store: baseStore}
	jwt, err := auth.NewJWTManager(cfg)
	if err != nil {
		t.Fatalf("jwt init failed: %v", err)
	}
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
	deploy := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverID+"/deploy", merchantToken, nil)
	if deploy.Code != http.StatusOK {
		t.Fatalf("deploy installable server status %d body=%s", deploy.Code, deploy.Body.String())
	}
	publish := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverID+"/publish", merchantToken, map[string]interface{}{"pricingAmount": 1})
	if publish.Code != http.StatusOK {
		t.Fatalf("publish installable server status %d body=%s", publish.Code, publish.Body.String())
	}

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
	hubToken := oauthAccessToken(t, h, buyerToken, buyerUser["tenantId"].(string), buyerUser["id"].(string), []string{"db:read", "db:write"})

	if res := call(t, h, http.MethodGet, hubPath, hubToken, nil); res.Code != http.StatusOK {
		t.Fatalf("mcp hub get status %d body=%s", res.Code, res.Body.String())
	}

	res := call(t, h, http.MethodPost, hubPath, hubToken, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "tools-list-1",
		"method":  "tools/list",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("mcp hub post status %d body=%s", res.Code, res.Body.String())
	}
}

func TestMarketplaceInstallAndMCPHubWithSDKMode(t *testing.T) {
	var upstreamCalls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode upstream request: %v", err)
		}
		if body["method"] != "tools/call" {
			t.Fatalf("unexpected upstream method: %v", body["method"])
		}
		params, ok := body["params"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected params to be an object, got %T (%v)", body["params"], body["params"])
		}
		if _, ok := params["arguments"].(map[string]interface{}); !ok {
			t.Fatalf("expected params.arguments object, got %T", params["arguments"])
		}
		upstreamCalls.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      body["id"],
			"result": map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "ok"},
				},
			},
		})
	}))
	defer upstream.Close()

	h := newTestServerWithConfig(config.Config{
		AllowInsecureDefaults: true,
		X402Mode:              "disabled",
		MCPSDKEnabled:         true,
	})
	buyerSignup := signup(t, h, "sdk-install-buyer@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "sdk-install-merchant@acme.local", "MerchantPass123!@", "merchant")
	buyerToken := login(t, h, "sdk-install-buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "sdk-install-merchant@acme.local", "MerchantPass123!@")

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "SDK Installable Server",
		"slug":                 "sdk-installable-server",
		"description":          "Install flow test",
		"category":             "integration",
		"dockerImage":          "tenant/sdk-installable:1.0.0",
		"canonicalResourceUri": upstream.URL,
		"requiredScopes":       []string{"db:read", "db:write"},
		"pricingType":          "free",
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
	deploy := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverID+"/deploy", merchantToken, nil)
	if deploy.Code != http.StatusOK {
		t.Fatalf("deploy installable server status %d body=%s", deploy.Code, deploy.Body.String())
	}
	publish := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverID+"/publish", merchantToken, map[string]interface{}{"pricingAmount": 1})
	if publish.Code != http.StatusOK {
		t.Fatalf("publish installable server status %d body=%s", publish.Code, publish.Body.String())
	}

	install := call(t, h, http.MethodPost, "/v1/marketplace/servers/"+serverSlug+"/install", buyerToken, map[string]interface{}{
		"client": "vscode",
	})
	if install.Code != http.StatusCreated {
		t.Fatalf("install endpoint status %d body=%s", install.Code, install.Body.String())
	}

	buyerUser := buyerSignup["user"].(map[string]interface{})
	hubPath := "/mcp/hub/" + buyerUser["tenantId"].(string) + "/" + buyerUser["id"].(string)
	hubToken := oauthAccessToken(t, h, buyerToken, buyerUser["tenantId"].(string), buyerUser["id"].(string), []string{"db:read", "db:write"})

	initRes := call(t, h, http.MethodPost, hubPath, hubToken, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "init-1",
		"method":  "initialize",
	})
	if initRes.Code != http.StatusOK {
		t.Fatalf("mcp initialize status %d body=%s", initRes.Code, initRes.Body.String())
	}
	var initBody map[string]interface{}
	_ = json.Unmarshal(initRes.Body.Bytes(), &initBody)
	initResult, _ := initBody["result"].(map[string]interface{})
	serverInfo, _ := initResult["serverInfo"].(map[string]interface{})
	if serverInfo["name"] != "mcp-marketplace-hub" {
		t.Fatalf("initialize serverInfo.name mismatch: %v", serverInfo["name"])
	}

	listRes := call(t, h, http.MethodPost, hubPath, hubToken, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "tools-list-1",
		"method":  "tools/list",
	})
	if listRes.Code != http.StatusOK {
		t.Fatalf("mcp hub tools/list status %d body=%s", listRes.Code, listRes.Body.String())
	}
	var listBody map[string]interface{}
	_ = json.Unmarshal(listRes.Body.Bytes(), &listBody)
	listResult, _ := listBody["result"].(map[string]interface{})
	tools, _ := listResult["tools"].([]interface{})
	if len(tools) == 0 {
		t.Fatalf("expected at least one tool in tools/list response")
	}

	firstTool, _ := tools[0].(map[string]interface{})
	toolName, _ := firstTool["name"].(string)
	if toolName == "" {
		t.Fatalf("expected a tool name in tools/list response")
	}

	callRes := call(t, h, http.MethodPost, hubPath, hubToken, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "call-1",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": toolName,
			"arguments": map[string]interface{}{
				"query": "ping",
			},
		},
	})
	if callRes.Code != http.StatusOK {
		t.Fatalf("mcp hub tools/call status %d body=%s", callRes.Code, callRes.Body.String())
	}
	var callBody map[string]interface{}
	_ = json.Unmarshal(callRes.Body.Bytes(), &callBody)
	if callBody["error"] != nil {
		t.Fatalf("unexpected tools/call error: %s", callRes.Body.String())
	}
	if upstreamCalls.Load() != 1 {
		t.Fatalf("expected one upstream tools/call invocation, got %d", upstreamCalls.Load())
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
	if createdBody["status"] != "draft" {
		t.Fatalf("expected new merchant server to start as draft, got %v", createdBody["status"])
	}
	if createdBody["deploymentStatus"] != "not_deployed" {
		t.Fatalf("expected new merchant server to start as not_deployed, got %v", createdBody["deploymentStatus"])
	}
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
	if codes := blockingReasonCodes(t, publishBeforeDeploy); !strings.Contains(strings.Join(codes, ","), "server_not_deployed") {
		t.Fatalf("expected deployment blocker, got %v", codes)
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
	if codes := blockingReasonCodes(t, publishWithoutPrice); !strings.Contains(strings.Join(codes, ","), "pricing_amount_required") {
		t.Fatalf("expected pricing blocker, got %v", codes)
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

func TestMerchantServerRejectsDirectPublishedBypasses(t *testing.T) {
	h := newTestServer()
	signup(t, h, "bypass-merchant@acme.local", "MerchantPass123!@", "merchant")
	merchantToken := login(t, h, "bypass-merchant@acme.local", "MerchantPass123!@")

	createPublished := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":        "Bypass Server",
		"slug":        "bypass-server",
		"dockerImage": "tenant/bypass:1.0.0",
		"status":      "published",
	})
	if createPublished.Code != http.StatusBadRequest {
		t.Fatalf("expected direct published create to fail, got %d body=%s", createPublished.Code, createPublished.Body.String())
	}

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Protected Published Server",
		"slug":                 "protected-published-server",
		"description":          "Prevent invalid published updates",
		"category":             "automation",
		"dockerImage":          "tenant/protected:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/protected-published-server",
		"requiredScopes":       []string{"db:read"},
		"pricingType":          "x402",
		"pricingAmount":        1.0,
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
	publish := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverID+"/publish", merchantToken, nil)
	if publish.Code != http.StatusOK {
		t.Fatalf("publish status %d body=%s", publish.Code, publish.Body.String())
	}

	updatePublished := call(t, h, http.MethodPut, "/v1/merchant/servers/"+serverID, merchantToken, map[string]interface{}{
		"pricingType":   "x402",
		"pricingAmount": 0.0,
	})
	if updatePublished.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected invalid published update to fail, got %d body=%s", updatePublished.Code, updatePublished.Body.String())
	}
	if codes := blockingReasonCodes(t, updatePublished); !strings.Contains(strings.Join(codes, ","), "pricing_amount_required") {
		t.Fatalf("expected pricing blocker, got %v", codes)
	}

	serverRes := call(t, h, http.MethodGet, "/v1/merchant/servers/"+serverID, merchantToken, nil)
	if serverRes.Code != http.StatusOK {
		t.Fatalf("get server status %d body=%s", serverRes.Code, serverRes.Body.String())
	}
	var serverBody map[string]interface{}
	_ = json.Unmarshal(serverRes.Body.Bytes(), &serverBody)
	serverObj := serverBody["server"].(map[string]interface{})
	if serverObj["status"] != "published" {
		t.Fatalf("expected server to remain published after rejected update, got %v", serverObj["status"])
	}
	if serverObj["pricingAmount"].(float64) != 1.0 {
		t.Fatalf("expected original pricing amount to remain intact, got %v", serverObj["pricingAmount"])
	}
}

func TestMarketplaceOnlyExposesValidPublishedAndDeployedServers(t *testing.T) {
	cfg := config.Config{
		Port:                  "8080",
		BaseURL:               "http://localhost:8080",
		SuperAdminEmail:       "admin@platform.local",
		SuperAdminPassword:    "admin-pass",
		AllowInsecureDefaults: true,
	}
	st := store.NewMemoryStore(cfg)
	h := newTestServerWithStore(cfg, st)

	valid := st.CreateServer(storeServerFixture("Visible Server", "visible-server", 1.0, []string{"x402_wallet"}))
	valid.Status = "published"
	valid.DeploymentStatus = "deployed"
	st.UpdateServer(valid)

	invalid := st.CreateServer(storeServerFixture("Hidden Server", "hidden-server", 0.0, []string{"x402_wallet"}))
	invalid.Status = "published"
	invalid.DeploymentStatus = "deployed"
	st.UpdateServer(invalid)

	list := call(t, h, http.MethodGet, "/v1/marketplace/servers", "", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("marketplace list status %d body=%s", list.Code, list.Body.String())
	}
	var listBody map[string]interface{}
	_ = json.Unmarshal(list.Body.Bytes(), &listBody)
	items := listBody["items"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("expected only one visible marketplace server, got %d body=%s", len(items), list.Body.String())
	}
	visible := items[0].(map[string]interface{})
	if visible["slug"] != "visible-server" {
		t.Fatalf("expected valid published server in marketplace, got %v", visible["slug"])
	}

	if res := call(t, h, http.MethodGet, "/v1/marketplace/servers/visible-server", "", nil); res.Code != http.StatusOK {
		t.Fatalf("expected valid marketplace detail, got %d body=%s", res.Code, res.Body.String())
	}
	if res := call(t, h, http.MethodGet, "/v1/marketplace/servers/hidden-server", "", nil); res.Code != http.StatusNotFound {
		t.Fatalf("expected invalid marketplace detail hidden, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestMerchantDeploySyncsWithN8NWhenConfigured(t *testing.T) {
	var createCalls int32
	var activateCalls int32

	n8n := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workflows":
			atomic.AddInt32(&createCalls, 1)
			if got := r.Header.Get("X-N8N-API-KEY"); got != "test-api-key" {
				t.Fatalf("expected X-N8N-API-KEY header, got %q", got)
			}
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create workflow body: %v", err)
			}
			if _, exists := body["active"]; exists {
				t.Fatalf("create workflow payload must not include read-only active field: %+v", body)
			}
			for _, field := range []string{"name", "nodes", "connections", "settings"} {
				if _, ok := body[field]; !ok {
					t.Fatalf("create workflow payload missing %s: %+v", field, body)
				}
			}
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
		Port:                  "8080",
		BaseURL:               "http://localhost:8080",
		N8NBaseURL:            n8n.URL,
		N8NAPIKey:             "test-api-key",
		SuperAdminEmail:       "admin@platform.local",
		SuperAdminPassword:    "admin-pass",
		AllowInsecureDefaults: true,
	}
	st := store.NewMemoryStore(cfg)
	jwt, err := auth.NewJWTManager(cfg)
	if err != nil {
		t.Fatalf("jwt init failed: %v", err)
	}
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
	if deploy.Code != http.StatusAccepted {
		t.Fatalf("deploy status %d body=%s", deploy.Code, deploy.Body.String())
	}

	deadline := time.Now().Add(3 * time.Second)
	workflowPersisted := false
	for time.Now().Before(deadline) {
		serverRes := call(t, h, http.MethodGet, "/v1/merchant/servers/"+serverID, merchantToken, nil)
		if serverRes.Code != http.StatusOK {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		var serverBody map[string]interface{}
		_ = json.Unmarshal(serverRes.Body.Bytes(), &serverBody)
		serverObj := serverBody["server"].(map[string]interface{})
		workflowID, _ := serverObj["n8nWorkflowId"].(string)
		deploymentStatus, _ := serverObj["deploymentStatus"].(string)
		if strings.TrimSpace(workflowID) == "wf_test_123" &&
			strings.TrimSpace(deploymentStatus) == "deployed" {
			workflowPersisted = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !workflowPersisted {
		t.Fatalf("expected queued deploy to complete and persist n8n workflow id")
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
	resource := "http://localhost:8080/mcp/hub/" + buyerUser["tenantId"].(string) + "/" + buyerUser["id"].(string)

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
	var tokenBody map[string]interface{}
	_ = json.Unmarshal(res.Body.Bytes(), &tokenBody)
	accessToken := tokenBody["access_token"].(string)
	if tokenBody["resource"] != resource {
		t.Fatalf("expected token resource %s, got %v", resource, tokenBody["resource"])
	}
	if tokenBody["audience"] != resource {
		t.Fatalf("expected token audience %s, got %v", resource, tokenBody["audience"])
	}
	hubPath := "/mcp/hub/" + buyerUser["tenantId"].(string) + "/" + buyerUser["id"].(string)
	if hubWithAppToken := call(t, h, http.MethodGet, hubPath, buyerToken, nil); hubWithAppToken.Code != http.StatusUnauthorized {
		t.Fatalf("expected app token rejected by mcp hub, got %d body=%s", hubWithAppToken.Code, hubWithAppToken.Body.String())
	}
	if meWithOAuthToken := call(t, h, http.MethodGet, "/v1/me", accessToken, nil); meWithOAuthToken.Code != http.StatusUnauthorized {
		t.Fatalf("expected oauth token rejected by app auth, got %d body=%s", meWithOAuthToken.Code, meWithOAuthToken.Body.String())
	}
	if hubWithOAuthToken := call(t, h, http.MethodGet, hubPath, accessToken, nil); hubWithOAuthToken.Code != http.StatusOK {
		t.Fatalf("expected oauth token accepted by mcp hub, got %d body=%s", hubWithOAuthToken.Code, hubWithOAuthToken.Body.String())
	}
}

func TestOAuthAuthorizeRejectsNonCanonicalResource(t *testing.T) {
	h := newTestServer()
	buyerSignup := signup(t, h, "oauth-canonical@acme.local", "BuyerPass123!@", "buyer")
	buyerToken := login(t, h, "oauth-canonical@acme.local", "BuyerPass123!@")
	buyerUser := buyerSignup["user"].(map[string]interface{})

	dcr := call(t, h, http.MethodPost, "/oauth/register", "", map[string]interface{}{
		"client_name":                "Canonical Resource Test Client",
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

	verifier := "canonical-verifier-1234567890"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	resource := "https://mcp.marketplace.local/mcp/hub/" + buyerUser["tenantId"].(string) + "/" + buyerUser["id"].(string)
	authorizePath := "/oauth/authorize?response_type=code&client_id=" + clientID + "&redirect_uri=http://127.0.0.1:33418&state=abc123&resource=" + resource + "&scope=db:read&code_challenge=" + challenge + "&code_challenge_method=S256"
	authRes := call(t, h, http.MethodGet, authorizePath, buyerToken, nil)
	if authRes.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid resource for off-origin hub URL, got %d body=%s", authRes.Code, authRes.Body.String())
	}
	var body map[string]interface{}
	_ = json.Unmarshal(authRes.Body.Bytes(), &body)
	if body["error"] != "invalid_resource" {
		t.Fatalf("expected invalid_resource error, got %v", body["error"])
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
	resourceB := "http://localhost:8080/mcp/hub/" + userBMeta["tenantId"].(string) + "/" + userBMeta["id"].(string)
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

func TestMCPHubFiltersToolsByScopeAndCloudEntitlement(t *testing.T) {
	h := newTestServer()
	buyerSignup := signup(t, h, "hub-filter-buyer@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "hub-filter-merchant@acme.local", "MerchantPass123!@", "merchant")
	buyerToken := login(t, h, "hub-filter-buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "hub-filter-merchant@acme.local", "MerchantPass123!@")
	adminToken := login(t, h, "admin@platform.local", "admin-pass")
	buyerUser := buyerSignup["user"].(map[string]interface{})

	createServer := func(name, slug string, scopes []string) string {
		t.Helper()
		created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
			"name":                 name,
			"slug":                 slug,
			"description":          "Hub filter test",
			"category":             "integration",
			"dockerImage":          "tenant/" + slug + ":1.0.0",
			"canonicalResourceUri": "https://mcp.marketplace.local/resource/" + slug,
			"requiredScopes":       scopes,
			"pricingType":          "free",
			"supportsCloud":        true,
			"supportsLocal":        true,
		})
		if created.Code != http.StatusCreated {
			t.Fatalf("create server %s status %d body=%s", slug, created.Code, created.Body.String())
		}
		var body map[string]interface{}
		_ = json.Unmarshal(created.Body.Bytes(), &body)
		return body["id"].(string)
	}

	allowedServerID := createServer("Allowed Cloud Tool", "allowed-cloud-tool", []string{"db:read"})
	scopeBlockedServerID := createServer("Scope Blocked Tool", "scope-blocked-tool", []string{"db:read", "db:write"})
	cloudBlockedServerID := createServer("Cloud Blocked Tool", "cloud-blocked-tool", []string{"db:read"})

	grant := func(serverID string, scopes []string, cloudAllowed bool) {
		t.Helper()
		res := call(t, h, http.MethodPost, "/v1/admin/entitlements", adminToken, map[string]interface{}{
			"tenantId":      buyerUser["tenantId"],
			"userId":        buyerUser["id"],
			"serverId":      serverID,
			"allowedScopes": scopes,
			"cloudAllowed":  cloudAllowed,
			"localAllowed":  true,
		})
		if res.Code != http.StatusCreated {
			t.Fatalf("grant entitlement status %d body=%s", res.Code, res.Body.String())
		}
	}

	grant(allowedServerID, []string{"db:read"}, true)
	grant(scopeBlockedServerID, []string{"db:read"}, true)
	grant(cloudBlockedServerID, []string{"db:read"}, false)

	hubToken := oauthAccessToken(t, h, buyerToken, buyerUser["tenantId"].(string), buyerUser["id"].(string), []string{"db:read"})
	hubPath := "/mcp/hub/" + buyerUser["tenantId"].(string) + "/" + buyerUser["id"].(string)

	listRes := call(t, h, http.MethodPost, hubPath, hubToken, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "tools-list-filtered",
		"method":  "tools/list",
	})
	if listRes.Code != http.StatusOK {
		t.Fatalf("tools/list status %d body=%s", listRes.Code, listRes.Body.String())
	}
	var listBody map[string]interface{}
	_ = json.Unmarshal(listRes.Body.Bytes(), &listBody)
	result := listBody["result"].(map[string]interface{})
	tools := result["tools"].([]interface{})
	toolNames := map[string]bool{}
	for _, raw := range tools {
		tool := raw.(map[string]interface{})
		toolNames[tool["name"].(string)] = true
	}
	if !toolNames["invoke_allowed_cloud_tool"] {
		t.Fatalf("expected granted cloud tool in tools/list, got %v", toolNames)
	}
	if toolNames["invoke_scope_blocked_tool"] {
		t.Fatalf("expected scope-blocked tool to be filtered, got %v", toolNames)
	}
	if toolNames["invoke_cloud_blocked_tool"] {
		t.Fatalf("expected cloud-blocked tool to be filtered, got %v", toolNames)
	}

	callRes := call(t, h, http.MethodPost, hubPath, hubToken, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "call-filtered-tool",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "invoke_scope_blocked_tool",
		},
	})
	if callRes.Code != http.StatusOK {
		t.Fatalf("tools/call status %d body=%s", callRes.Code, callRes.Body.String())
	}
	var callBody map[string]interface{}
	_ = json.Unmarshal(callRes.Body.Bytes(), &callBody)
	errorBody := callBody["error"].(map[string]interface{})
	if errorBody["message"] != "tool not found" {
		t.Fatalf("expected filtered tool call to return tool not found, got %v", errorBody)
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

func TestMerchantServerBuilderCanPersistConfig(t *testing.T) {
	h := newTestServer()
	signup(t, h, "merchant-builder@acme.local", "MerchantPass123!@", "merchant")
	token := login(t, h, "merchant-builder@acme.local", "MerchantPass123!@")

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", token, map[string]interface{}{
		"name":                 "Builder Server",
		"slug":                 "builder-server",
		"description":          "Builder persistence test",
		"category":             "automation",
		"dockerImage":          "docker.io/acme/builder-server:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/builder-server",
		"requiredScopes":       []string{"documents:read"},
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

	update := call(t, h, http.MethodPut, "/v1/merchant/servers/"+id+"/builder", token, map[string]interface{}{
		"framework":    "FastMCP",
		"template":     "docker-import",
		"instructions": "Persist the builder config",
		"scopeMappings": []string{
			"documents:read",
			"agents:run",
		},
		"toolCatalog": []map[string]interface{}{
			{
				"name":        "search_docs",
				"description": "Search documents",
				"inputSchema": map[string]string{"query": "string"},
				"outputSchema": map[string]string{
					"rows": "array",
				},
			},
		},
	})
	if update.Code != http.StatusOK {
		t.Fatalf("update builder status %d body=%s", update.Code, update.Body.String())
	}

	get := call(t, h, http.MethodGet, "/v1/merchant/servers/"+id+"/builder", token, nil)
	if get.Code != http.StatusOK {
		t.Fatalf("get builder status %d body=%s", get.Code, get.Body.String())
	}
	var builderBody map[string]interface{}
	_ = json.Unmarshal(get.Body.Bytes(), &builderBody)
	if builderBody["framework"] != "FastMCP" {
		t.Fatalf("expected framework FastMCP, got %v", builderBody["framework"])
	}
	if builderBody["template"] != "docker-import" {
		t.Fatalf("expected template docker-import, got %v", builderBody["template"])
	}
	tools := builderBody["toolCatalog"].([]interface{})
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	server := call(t, h, http.MethodGet, "/v1/merchant/servers/"+id, token, nil)
	if server.Code != http.StatusOK {
		t.Fatalf("get server status %d body=%s", server.Code, server.Body.String())
	}
	var serverBody map[string]interface{}
	_ = json.Unmarshal(server.Body.Bytes(), &serverBody)
	serverObj := serverBody["server"].(map[string]interface{})
	requiredScopes := serverObj["requiredScopes"].([]interface{})
	if len(requiredScopes) != 2 {
		t.Fatalf("expected required scopes to sync from builder, got %d", len(requiredScopes))
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
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "mcp-pay-2",
			"result": map[string]interface{}{
				"content": []map[string]string{{
					"type": "text",
					"text": "upstream tool executed",
				}},
				"isError": false,
			},
		})
	}))
	defer upstream.Close()
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
		"canonicalResourceUri": upstream.URL,
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
	hubToken := oauthAccessToken(t, h, buyerToken, buyerUser["tenantId"].(string), buyerUser["id"].(string), []string{"documents:read"})

	needPay := call(t, h, http.MethodPost, hubPath, hubToken, map[string]interface{}{
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

	paid := call(t, h, http.MethodPost, hubPath, hubToken, map[string]interface{}{
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

	replay := call(t, h, http.MethodPost, hubPath, hubToken, map[string]interface{}{
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
		"paymentMethods": []string{"x402_wallet"},
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

func TestPaidInstallFlowAndInstallAutoSettle(t *testing.T) {
	h := newTestServer()
	signup(t, h, "installpay-buyer@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "installpay-merchant@acme.local", "MerchantPass123!@", "merchant")
	buyerToken := login(t, h, "installpay-buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "installpay-merchant@acme.local", "MerchantPass123!@")

	controls := call(t, h, http.MethodPut, "/v1/buyer/payments/controls", buyerToken, map[string]interface{}{
		"allowedMethods":     []string{"wallet_balance"},
		"minimumBalanceUsdc": 0.0,
		"hardStopOnLowFunds": false,
		"walletAddress":      "0xinstallpaybuyer",
	})
	if controls.Code != http.StatusOK {
		t.Fatalf("update buyer controls status %d body=%s", controls.Code, controls.Body.String())
	}

	topup := call(t, h, http.MethodPost, "/v1/buyer/payments/topups/stripe/session", buyerToken, map[string]interface{}{
		"amountUsd":     15.0,
		"walletAddress": "0xinstallpaybuyer",
	})
	if topup.Code != http.StatusCreated {
		t.Fatalf("create topup session status %d body=%s", topup.Code, topup.Body.String())
	}
	var topupBody map[string]interface{}
	_ = json.Unmarshal(topup.Body.Bytes(), &topupBody)
	topupObj := topupBody["topup"].(map[string]interface{})
	providerSessionID := topupObj["providerSessionId"].(string)
	webhook := call(t, h, http.MethodPost, "/webhooks/stripe/onramp", "", map[string]interface{}{
		"id":   "evt_test_installpay_1",
		"type": "crypto.onramp_session_updated",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":                 providerSessionID,
				"status":             "fulfillment_complete",
				"destination_amount": 15.0,
			},
		},
	})
	if webhook.Code != http.StatusOK {
		t.Fatalf("stripe webhook status %d body=%s", webhook.Code, webhook.Body.String())
	}

	createdA := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Paid Install A",
		"slug":                 "paid-install-a",
		"description":          "Paid install A",
		"category":             "ai",
		"dockerImage":          "tenant/paid-install-a:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/paid-install-a",
		"requiredScopes":       []string{"documents:read"},
		"pricingType":          "x402",
		"pricingAmount":        0.5,
		"supportsCloud":        true,
		"supportsLocal":        true,
		"paymentMethods":       []string{"wallet_balance"},
	})
	if createdA.Code != http.StatusCreated {
		t.Fatalf("create server A status %d body=%s", createdA.Code, createdA.Body.String())
	}
	var createdABody map[string]interface{}
	_ = json.Unmarshal(createdA.Body.Bytes(), &createdABody)
	serverIDA := createdABody["id"].(string)
	slugA := createdABody["slug"].(string)
	if deployA := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverIDA+"/deploy", merchantToken, nil); deployA.Code != http.StatusOK {
		t.Fatalf("deploy server A status %d body=%s", deployA.Code, deployA.Body.String())
	}
	if publishA := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverIDA+"/publish", merchantToken, nil); publishA.Code != http.StatusOK {
		t.Fatalf("publish server A status %d body=%s", publishA.Code, publishA.Body.String())
	}

	installNeedsPayment := call(t, h, http.MethodPost, "/v1/marketplace/servers/"+slugA+"/install", buyerToken, map[string]interface{}{
		"client":        "vscode",
		"paymentMethod": "wallet_balance",
	})
	if installNeedsPayment.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402 for paid install without entitlement, got %d body=%s", installNeedsPayment.Code, installNeedsPayment.Body.String())
	}
	var installNeedBody map[string]interface{}
	_ = json.Unmarshal(installNeedsPayment.Body.Bytes(), &installNeedBody)
	intentObj, _ := installNeedBody["intent"].(map[string]interface{})
	intentID, _ := intentObj["id"].(string)
	if intentID == "" {
		t.Fatalf("expected payment-required install response to include intent id: %s", installNeedsPayment.Body.String())
	}

	settle := call(t, h, http.MethodPost, "/v1/billing/x402/intents/"+intentID+"/settle", buyerToken, nil)
	if settle.Code != http.StatusOK {
		t.Fatalf("settle install intent status %d body=%s", settle.Code, settle.Body.String())
	}

	installAfterPay := call(t, h, http.MethodPost, "/v1/marketplace/servers/"+slugA+"/install", buyerToken, map[string]interface{}{
		"client":        "vscode",
		"paymentMethod": "wallet_balance",
	})
	if installAfterPay.Code != http.StatusCreated {
		t.Fatalf("expected install success after payment settle, got %d body=%s", installAfterPay.Code, installAfterPay.Body.String())
	}

	createdB := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Paid Install B",
		"slug":                 "paid-install-b",
		"description":          "Paid install B",
		"category":             "ai",
		"dockerImage":          "tenant/paid-install-b:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/paid-install-b",
		"requiredScopes":       []string{"documents:read"},
		"pricingType":          "x402",
		"pricingAmount":        0.25,
		"supportsCloud":        true,
		"supportsLocal":        true,
		"paymentMethods":       []string{"wallet_balance"},
	})
	if createdB.Code != http.StatusCreated {
		t.Fatalf("create server B status %d body=%s", createdB.Code, createdB.Body.String())
	}
	var createdBBody map[string]interface{}
	_ = json.Unmarshal(createdB.Body.Bytes(), &createdBBody)
	serverIDB := createdBBody["id"].(string)
	slugB := createdBBody["slug"].(string)
	if deployB := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverIDB+"/deploy", merchantToken, nil); deployB.Code != http.StatusOK {
		t.Fatalf("deploy server B status %d body=%s", deployB.Code, deployB.Body.String())
	}
	if publishB := call(t, h, http.MethodPost, "/v1/merchant/servers/"+serverIDB+"/publish", merchantToken, nil); publishB.Code != http.StatusOK {
		t.Fatalf("publish server B status %d body=%s", publishB.Code, publishB.Body.String())
	}

	autoSettleInstall := call(t, h, http.MethodPost, "/v1/marketplace/servers/"+slugB+"/install", buyerToken, map[string]interface{}{
		"client":        "vscode",
		"paymentMethod": "wallet_balance",
		"autoSettle":    true,
		"toolName":      "install_paid_install_b",
	})
	if autoSettleInstall.Code != http.StatusCreated {
		t.Fatalf("expected auto-settle install success, got %d body=%s", autoSettleInstall.Code, autoSettleInstall.Body.String())
	}
}

func TestX402IntentPreservesChallengeAfterSettle(t *testing.T) {
	h := newTestServer()
	signup(t, h, "preserve-buyer@acme.local", "BuyerPass123!@", "buyer")
	signup(t, h, "preserve-merchant@acme.local", "MerchantPass123!@", "merchant")
	buyerToken := login(t, h, "preserve-buyer@acme.local", "BuyerPass123!@")
	merchantToken := login(t, h, "preserve-merchant@acme.local", "MerchantPass123!@")

	created := call(t, h, http.MethodPost, "/v1/merchant/servers", merchantToken, map[string]interface{}{
		"name":                 "Challenge Server",
		"slug":                 "challenge-server",
		"description":          "Challenge server",
		"category":             "integration",
		"dockerImage":          "tenant/challenge-server:1.0.0",
		"canonicalResourceUri": "https://mcp.marketplace.local/resource/challenge-server",
		"requiredScopes":       []string{"documents:read"},
		"pricingType":          "x402",
		"pricingAmount":        0.25,
		"supportsCloud":        true,
		"supportsLocal":        true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create merchant server status %d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]interface{}
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)
	serverID := createdBody["id"].(string)

	intent := call(t, h, http.MethodPost, "/v1/billing/x402/intents", buyerToken, map[string]interface{}{
		"serverId":       serverID,
		"toolName":       "extract_pdf",
		"paymentMethod":  "x402_wallet",
		"idempotencyKey": "preserve_intent_1",
	})
	if intent.Code != http.StatusPaymentRequired {
		t.Fatalf("x402 intent status %d body=%s", intent.Code, intent.Body.String())
	}
	var intentBody map[string]interface{}
	_ = json.Unmarshal(intent.Body.Bytes(), &intentBody)
	intentObj := intentBody["intent"].(map[string]interface{})
	challenge := intentObj["challenge"].(string)
	if !strings.Contains(challenge, "preserve_intent_1") || !strings.Contains(challenge, serverID) {
		t.Fatalf("expected challenge payload to be preserved, got %s", challenge)
	}
	if got := intent.Header().Get("PAYMENT-REQUIRED"); got != challenge {
		t.Fatalf("expected PAYMENT-REQUIRED header to match stored challenge\nheader=%s\nchallenge=%s", got, challenge)
	}
	intentID := intentObj["id"].(string)

	settled := call(t, h, http.MethodPost, "/v1/billing/x402/intents/"+intentID+"/settle", buyerToken, map[string]interface{}{
		"paymentResponse": map[string]interface{}{"paymentIdentifier": "pay_preserve_1", "method": "x402_wallet"},
	})
	if settled.Code != http.StatusOK {
		t.Fatalf("settle status %d body=%s", settled.Code, settled.Body.String())
	}

	list := call(t, h, http.MethodGet, "/v1/billing/x402/intents", buyerToken, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list intents status %d body=%s", list.Code, list.Body.String())
	}
	var listBody map[string]interface{}
	_ = json.Unmarshal(list.Body.Bytes(), &listBody)
	items := listBody["items"].([]interface{})
	stored := items[0].(map[string]interface{})
	if stored["challenge"].(string) != challenge {
		t.Fatalf("expected settled intent to retain original challenge\nwant=%s\ngot=%s", challenge, stored["challenge"].(string))
	}
}

func TestPaymentMethodCatalogIsTruthfulAndDefaultsStaySafe(t *testing.T) {
	h := newTestServerWithConfig(config.Config{
		AllowInsecureDefaults: true,
		SupportedPayMethods:   []string{"x402_wallet", "wallet_balance", "stripe", "coinbase_commerce"},
	})
	signup(t, h, "catalog-buyer@acme.local", "BuyerPass123!@", "buyer")
	buyerToken := login(t, h, "catalog-buyer@acme.local", "BuyerPass123!@")

	controls := call(t, h, http.MethodGet, "/v1/buyer/payments/controls", buyerToken, nil)
	if controls.Code != http.StatusOK {
		t.Fatalf("buyer payment controls status %d body=%s", controls.Code, controls.Body.String())
	}
	var controlsBody map[string]interface{}
	_ = json.Unmarshal(controls.Body.Bytes(), &controlsBody)
	policy := controlsBody["policy"].(map[string]interface{})
	allowed := policy["allowedMethods"].([]interface{})
	if len(allowed) != 2 || allowed[0].(string) != "wallet_balance" || allowed[1].(string) != "x402_wallet" {
		t.Fatalf("expected safe default allowed methods, got %v", allowed)
	}

	methods := controlsBody["methods"].([]interface{})
	byID := map[string]map[string]interface{}{}
	for _, raw := range methods {
		item := raw.(map[string]interface{})
		byID[item["id"].(string)] = item
	}
	if byID["stripe"]["enabled"] != false || byID["stripe"]["readiness"] != "not_configured" {
		t.Fatalf("expected stripe to be disabled not_configured, got %v", byID["stripe"])
	}
	if byID["coinbase_commerce"]["enabled"] != false || byID["coinbase_commerce"]["readiness"] != "not_configured" {
		t.Fatalf("expected coinbase_commerce to be disabled not_configured, got %v", byID["coinbase_commerce"])
	}

	blocked := call(t, h, http.MethodPut, "/v1/buyer/payments/controls", buyerToken, map[string]interface{}{
		"allowedMethods": []string{"stripe"},
	})
	if blocked.Code != http.StatusBadRequest {
		t.Fatalf("expected not_configured payment method to be rejected, got %d body=%s", blocked.Code, blocked.Body.String())
	}
}

func TestStripeFailsClosedOutsideDevMode(t *testing.T) {
	h := newTestServerWithConfig(config.Config{AllowInsecureDefaults: false})
	signup(t, h, "strict-buyer@acme.local", "BuyerPass123!@", "buyer")
	buyerToken := login(t, h, "strict-buyer@acme.local", "BuyerPass123!@")

	topup := call(t, h, http.MethodPost, "/v1/buyer/payments/topups/stripe/session", buyerToken, map[string]interface{}{
		"amountUsd":     25.0,
		"walletAddress": "0xstrictbuyer",
	})
	if topup.Code != http.StatusBadGateway {
		t.Fatalf("expected strict mode stripe topup to fail closed, got %d body=%s", topup.Code, topup.Body.String())
	}

	webhook := call(t, h, http.MethodPost, "/webhooks/stripe/onramp", "", map[string]interface{}{
		"id":   "evt_strict_1",
		"type": "crypto.onramp_session_updated",
		"data": map[string]interface{}{"object": map[string]interface{}{"id": "session_1"}},
	})
	if webhook.Code != http.StatusUnauthorized {
		t.Fatalf("expected strict mode webhook to fail closed, got %d body=%s", webhook.Code, webhook.Body.String())
	}
}
