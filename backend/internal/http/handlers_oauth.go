package http

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type dcrRequest struct {
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

func (a *App) oauthProtectedResourceMetadata(w http.ResponseWriter, r *http.Request) {
	resource := strings.TrimSpace(r.URL.Query().Get("resource"))
	if resource == "" {
		resource = strings.TrimRight(a.cfg.BaseURL, "/") + "/mcp/hub/{tenantID}/{userID}"
	} else {
		canonical, _, _, ok := canonicalHubResource(a.cfg.BaseURL, resource)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_resource"})
			return
		}
		resource = canonical
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"resource":                 resource,
		"authorization_servers":    []string{a.cfg.BaseURL},
		"bearer_methods_supported": []string{"header"},
		"scopes_supported":         []string{"mcp:invoke", "mcp:manage", "db:read", "db:write", "documents:read", "ai:inference"},
	})
}

func (a *App) oauthAuthorizationServerMetadata(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"issuer":                                a.cfg.BaseURL,
		"authorization_endpoint":                a.cfg.BaseURL + "/oauth/authorize",
		"token_endpoint":                        a.cfg.BaseURL + "/oauth/token",
		"registration_endpoint":                 a.cfg.BaseURL + "/oauth/register",
		"jwks_uri":                              a.cfg.BaseURL + "/.well-known/jwks.json",
		"code_challenge_methods_supported":      []string{"S256"},
		"grant_types_supported":                 []string{"authorization_code"},
		"response_types_supported":              []string{"code"},
		"token_endpoint_auth_methods_supported": []string{"none", "client_secret_post"},
		"client_id_metadata_document_supported": true,
	})
}

func (a *App) jwks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.jwt.JWKS())
}

func (a *App) oauthRegisterClient(w http.ResponseWriter, r *http.Request) {
	var req dcrRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ClientName == "" || len(req.RedirectURIs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_client_metadata"})
		return
	}
	clientID := randomToken("cli_")
	clientSecret := randomToken("sec_")
	if req.TokenEndpointAuthMethod == "" {
		req.TokenEndpointAuthMethod = "none"
	}
	if len(req.GrantTypes) == 0 {
		req.GrantTypes = []string{"authorization_code"}
	}
	for _, grantType := range req.GrantTypes {
		if grantType != "authorization_code" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
			return
		}
	}
	client := a.store.CreateOAuthClient(models.OAuthClient{
		ClientID:                clientID,
		ClientName:              req.ClientName,
		RedirectURIs:            req.RedirectURIs,
		GrantTypes:              req.GrantTypes,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		ClientSecret:            clientSecret,
	})
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"client_id":                  client.ClientID,
		"client_secret":              client.ClientSecret,
		"redirect_uris":              client.RedirectURIs,
		"grant_types":                client.GrantTypes,
		"token_endpoint_auth_method": client.TokenEndpointAuthMethod,
	})
}

func (a *App) oauthAuthorize(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	q := r.URL.Query()
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	resource := q.Get("resource")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")
	scope := q.Get("scope")

	if q.Get("response_type") != "code" || clientID == "" || redirectURI == "" || resource == "" || codeChallenge == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	canonicalResource, tenantID, userID, ok := canonicalHubResource(a.cfg.BaseURL, resource)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_resource"})
		return
	}
	resource = canonicalResource
	if claims.UserID != userID || claims.TenantID != tenantID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "resource subject mismatch"})
		return
	}
	if user, exists := a.store.GetUserByID(userID); !exists || user.TenantID != tenantID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_resource_subject"})
		return
	}
	client, ok := a.store.GetOAuthClient(clientID)
	if !ok || !containsString(client.RedirectURIs, redirectURI) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_client_or_redirect"})
		return
	}
	code := randomToken("code_")
	_ = a.store.CreateOAuthAuthCode(models.OAuthAuthCode{
		Code:                code,
		ClientID:            clientID,
		UserID:              userID,
		TenantID:            tenantID,
		RedirectURI:         redirectURI,
		Resource:            resource,
		Scopes:              splitScopes(scope),
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		ExpiresAt:           time.Now().UTC().Add(5 * time.Minute),
	})

	rURL, _ := url.Parse(redirectURI)
	params := rURL.Query()
	params.Set("code", code)
	if state != "" {
		params.Set("state", state)
	}
	rURL.RawQuery = params.Encode()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"redirect_to": rURL.String(),
		"code":        code,
		"state":       state,
	})
}

func (a *App) oauthToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	grantType := r.PostFormValue("grant_type")
	if grantType != "authorization_code" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
		return
	}

	clientID := r.PostFormValue("client_id")
	clientSecret := r.PostFormValue("client_secret")
	code := r.PostFormValue("code")
	redirectURI := r.PostFormValue("redirect_uri")
	verifier := r.PostFormValue("code_verifier")
	resource := r.PostFormValue("resource")
	canonicalResource, _, _, ok := canonicalHubResource(a.cfg.BaseURL, resource)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_resource"})
		return
	}
	resource = canonicalResource

	client, ok := a.store.GetOAuthClient(clientID)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
		return
	}
	if strings.EqualFold(client.TokenEndpointAuthMethod, "client_secret_post") && client.ClientSecret != clientSecret {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client_secret"})
		return
	}
	ac, ok := a.store.GetOAuthAuthCode(code)
	if !ok || ac.Consumed || time.Now().UTC().After(ac.ExpiresAt) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	if ac.ClientID != clientID || ac.RedirectURI != redirectURI || ac.Resource != resource {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant_binding"})
		return
	}
	if !verifyPKCE(verifier, ac.CodeChallengeMethod, ac.CodeChallenge) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_code_verifier"})
		return
	}
	ac, _ = a.store.ConsumeOAuthAuthCode(code)

	user, ok := a.store.GetUserByID(ac.UserID)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_not_found"})
		return
	}
	accessToken, err := a.jwt.GenerateOAuthAccessToken(user, resource, ac.Scopes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token_generation_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   28800,
		"scope":        joinScopes(ac.Scopes),
		"audience":     tokenAudience(resource),
		"resource":     resource,
	})
}
