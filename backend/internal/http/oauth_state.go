package http

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"sync"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type oauthClient struct {
	ClientID                string
	ClientName              string
	RedirectURIs            []string
	GrantTypes              []string
	TokenEndpointAuthMethod string
	ClientSecret            string
	CreatedAt               time.Time
}

type authCode struct {
	Code                string
	ClientID            string
	UserID              string
	TenantID            string
	RedirectURI         string
	Resource            string
	Scopes              []string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
	Consumed            bool
}

type oauthStateData struct {
	Provider    models.OAuthProvider
	Nonce       string
	CreatedAt   time.Time
	CallbackURL string
}

type oauthState struct {
	mu      sync.Mutex
	clients map[string]oauthClient
	codes   map[string]authCode
	states  map[string]oauthStateData
}

func newOAuthState() *oauthState {
	now := time.Now().UTC()
	return &oauthState{
		clients: map[string]oauthClient{
			"vscode-dev": {
				ClientID:                "vscode-dev",
				ClientName:              "VS Code Dev",
				RedirectURIs:            []string{"http://127.0.0.1:33418", "https://vscode.dev/redirect"},
				GrantTypes:              []string{"authorization_code"},
				TokenEndpointAuthMethod: "none",
				CreatedAt:               now,
			},
		},
		codes:  map[string]authCode{},
		states: map[string]oauthStateData{},
	}
}

func randomToken(prefix string) string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return prefix + base64.RawURLEncoding.EncodeToString(b)
}

func verifyPKCE(verifier, method, challenge string) bool {
	if method == "" {
		method = "S256"
	}
	switch strings.ToUpper(method) {
	case "S256":
		h := sha256.Sum256([]byte(verifier))
		computed := base64.RawURLEncoding.EncodeToString(h[:])
		return computed == challenge
	default:
		return false
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func splitScopes(scopes string) []string {
	if strings.TrimSpace(scopes) == "" {
		return []string{}
	}
	parts := strings.Fields(scopes)
	return parts
}

func joinScopes(scopes []string) string {
	return strings.Join(scopes, " ")
}

func tokenAudience(resource string) string {
	return resource
}
