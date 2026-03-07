package tests

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestPublicEndpoints(t *testing.T) {
	h := newTestServer()
	if res := call(t, h, http.MethodGet, "/health", "", nil); res.Code != http.StatusOK {
		t.Fatalf("health status %d", res.Code)
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
	intentID := ibody["id"].(string)
	if res := call(t, h, http.MethodPost, "/v1/billing/x402/intents/"+intentID+"/settle", buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("settle status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/billing/x402/intents", buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("list intents status %d", res.Code)
	}
}

func TestOAuthAuthorizationCodePKCEFlow(t *testing.T) {
	h := newTestServer()
	buyerSignup := signup(t, h, "oauth-buyer@acme.local", "BuyerPass123!@", "buyer")
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
	authRes := call(t, h, http.MethodGet, authorizePath, "", nil)
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
