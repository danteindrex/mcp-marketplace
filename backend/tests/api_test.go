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
	cfg := config.Config{Port: "8080", JWTSecret: "test-secret", BaseURL: "http://localhost:8080"}
	st := store.NewMemoryStore()
	jwt := auth.NewJWTManager(cfg.JWTSecret)
	return api.NewRouter(cfg, st, jwt)
}

func login(t *testing.T, h http.Handler, email string) string {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{"email": email})
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
	if res := call(t, h, http.MethodGet, "/v1/marketplace/servers/postgresql-assistant", "", nil); res.Code != http.StatusOK {
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
	buyerToken := login(t, h, "buyer@acme.local")

	if res := call(t, h, http.MethodGet, "/v1/buyer/entitlements", buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("entitlements status %d", res.Code)
	}
	if res := call(t, h, http.MethodGet, "/v1/buyer/hub", buyerToken, nil); res.Code != http.StatusOK {
		t.Fatalf("hub status %d", res.Code)
	}

	res := call(t, h, http.MethodPost, "/v1/buyer/connections", buyerToken, map[string]interface{}{
		"client": "vscode", "resource": "https://mcp.marketplace.local/hub/tenant_acme/user_buyer", "grantedScopes": []string{"db:read"},
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

func TestMerchantAccessControl(t *testing.T) {
	h := newTestServer()
	buyerToken := login(t, h, "buyer@acme.local")
	merchantToken := login(t, h, "merchant@dataflow.local")

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
	adminToken := login(t, h, "admin@platform.local")
	buyerToken := login(t, h, "buyer@acme.local")

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
		"tenantId": "tenant_acme", "userId": "user_buyer", "serverId": "srv_doc", "allowedScopes": []string{"documents:read"}, "cloudAllowed": true, "localAllowed": true,
	})
	if grant.Code != http.StatusCreated {
		t.Fatalf("grant entitlement status %d body=%s", grant.Code, grant.Body.String())
	}

	intent := call(t, h, http.MethodPost, "/v1/billing/x402/intents", buyerToken, map[string]interface{}{
		"serverId": "srv_doc", "toolName": "extract_pdf", "amount": 0.05,
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
	authorizePath := "/oauth/authorize?response_type=code&client_id=" + clientID + "&redirect_uri=http://127.0.0.1:33418&state=abc123&resource=https://mcp.marketplace.local/hub/tenant_acme/user_buyer&scope=db:read%20db:write&code_challenge=" + challenge + "&code_challenge_method=S256"
	authRes := call(t, h, http.MethodGet, authorizePath, "", nil)
	if authRes.Code != http.StatusOK {
		t.Fatalf("authorize status %d body=%s", authRes.Code, authRes.Body.String())
	}
	var authBody map[string]interface{}
	_ = json.Unmarshal(authRes.Body.Bytes(), &authBody)
	code := authBody["code"].(string)

	form := "grant_type=authorization_code&client_id=" + clientID + "&code=" + code + "&redirect_uri=http://127.0.0.1:33418&code_verifier=" + verifier + "&resource=https://mcp.marketplace.local/hub/tenant_acme/user_buyer"
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("token status %d body=%s", res.Code, res.Body.String())
	}
}
