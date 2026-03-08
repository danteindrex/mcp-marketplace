package auth

import (
	"testing"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func TestJWTManagerValidatesIssuerAndPurpose(t *testing.T) {
	cfg := config.Config{BaseURL: "http://localhost:8080"}
	manager, err := NewJWTManager(cfg)
	if err != nil {
		t.Fatalf("NewJWTManager() error = %v", err)
	}
	user := models.User{ID: "user-1", TenantID: "tenant-1", Role: models.RoleBuyer}

	appToken, err := manager.GenerateAppToken(user)
	if err != nil {
		t.Fatalf("GenerateAppToken() error = %v", err)
	}
	oauthToken, err := manager.GenerateOAuthAccessToken(user, "https://resource.local/hub/tenant-1/user-1", []string{"db:read"})
	if err != nil {
		t.Fatalf("GenerateOAuthAccessToken() error = %v", err)
	}

	appClaims, err := manager.ParseForPurpose(appToken, TokenPurposeAppAuth)
	if err != nil {
		t.Fatalf("ParseForPurpose(app) error = %v", err)
	}
	if appClaims.Issuer != cfg.BaseURL {
		t.Fatalf("expected issuer %q, got %q", cfg.BaseURL, appClaims.Issuer)
	}
	if appClaims.Purpose != TokenPurposeAppAuth {
		t.Fatalf("expected purpose %q, got %q", TokenPurposeAppAuth, appClaims.Purpose)
	}

	oauthClaims, err := manager.ParseForPurpose(oauthToken, TokenPurposeOAuthAccess)
	if err != nil {
		t.Fatalf("ParseForPurpose(oauth) error = %v", err)
	}
	if oauthClaims.Resource != "https://resource.local/hub/tenant-1/user-1" {
		t.Fatalf("unexpected resource claim %q", oauthClaims.Resource)
	}
	if len(oauthClaims.Scopes) != 1 || oauthClaims.Scopes[0] != "db:read" {
		t.Fatalf("unexpected scopes %#v", oauthClaims.Scopes)
	}

	if _, err := manager.ParseForPurpose(appToken, TokenPurposeOAuthAccess); err == nil {
		t.Fatalf("expected app token purpose mismatch to fail")
	}

	otherManager, err := NewJWTManager(config.Config{BaseURL: "http://other-issuer.local"})
	if err != nil {
		t.Fatalf("NewJWTManager(other) error = %v", err)
	}
	otherManager.publicKey = manager.publicKey
	if _, err := otherManager.ParseForPurpose(appToken, TokenPurposeAppAuth); err == nil {
		t.Fatalf("expected issuer mismatch to fail")
	}
}
