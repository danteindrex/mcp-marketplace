package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/pquerna/otp/totp"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	github "golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

const appSessionMaxAgeSeconds = 60 * 60 * 8

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	MFACode  string `json:"mfaCode,omitempty"`
}

type signupRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	TenantName string `json:"tenantName"`
}

type oauthSignupIntent struct {
	Mode       string
	Role       models.Role
	Name       string
	TenantName string
}

type completeOAuthSignupRequest struct {
	SignupToken string `json:"signupToken"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	TenantName  string `json:"tenantName"`
}

func isStrongPassword(v string) bool {
	if len(v) < 12 {
		return false
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSymbol := false
	for _, r := range v {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSymbol = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSymbol
}

func normalizeRole(role string) models.Role {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case string(models.RoleMerchant):
		return models.RoleMerchant
	case string(models.RoleAdmin):
		return models.RoleAdmin
	default:
		return models.RoleBuyer
	}
}

func slugifyTenant(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	if slug == "" {
		slug = "tenant"
	}
	return slug
}

func loginResponse(user models.User, token string) map[string]interface{} {
	return map[string]interface{}{
		"accessToken": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"tenantId": user.TenantID,
			"email":    user.Email,
			"name":     user.Name,
			"role":     roleForUser(user),
		},
		"oauth": map[string]interface{}{
			"pkceRequired":              true,
			"resourceIndicatorRequired": true,
			"registrationModes":         []string{"pre_registered", "cimd", "dcr"},
		},
	}
}

func (a *App) signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	req.TenantName = strings.TrimSpace(req.TenantName)
	if req.Email == "" || req.Password == "" || req.Name == "" || req.TenantName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email, password, name, tenantName are required"})
		return
	}
	if !isStrongPassword(req.Password) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 12 characters and include upper, lower, number, and symbol"})
		return
	}
	if _, exists := a.store.GetUserByEmail(req.Email); exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
		return
	}
	role := normalizeRole(req.Role)
	if role == models.RoleAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin signup is disabled; bootstrap super admin only"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	tenant := a.store.CreateTenant(models.Tenant{
		Name:     req.TenantName,
		Slug:     slugifyTenant(req.TenantName),
		PlanTier: "professional",
		Status:   "active",
	})

	user, ok := a.store.CreateUser(models.User{
		TenantID:     tenant.ID,
		Email:        req.Email,
		Name:         req.Name,
		Role:         role,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now().UTC(),
	})
	if !ok {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
		return
	}
	tenant.OwnerUserID = user.ID
	_ = a.store.UpdateTenant(tenant)
	a.store.AddAuditLog(models.AuditLog{
		TenantID:   user.TenantID,
		ActorID:    user.ID,
		Action:     "auth.signup.success",
		TargetType: "user",
		TargetID:   user.ID,
		Outcome:    "success",
		Metadata:   map[string]interface{}{"email": user.Email, "role": roleName(user.Role)},
	})

	token, err := a.jwt.GenerateAppToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
		return
	}
	writeJSON(w, http.StatusCreated, loginResponse(user, token))
}

func (a *App) completeOAuthSignup(w http.ResponseWriter, r *http.Request) {
	var req completeOAuthSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	req.SignupToken = strings.TrimSpace(req.SignupToken)
	req.Name = strings.TrimSpace(req.Name)
	req.TenantName = strings.TrimSpace(req.TenantName)
	if req.SignupToken == "" || req.Name == "" || req.TenantName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "signupToken, name, and tenantName are required"})
		return
	}

	role := normalizeRole(req.Role)
	if role == models.RoleAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin signup is disabled; bootstrap super admin only"})
		return
	}

	pending, ok := a.consumePendingOAuthSignup(req.SignupToken)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired signup token"})
		return
	}

	if _, exists := a.store.GetOAuthAccount(pending.Provider, pending.ProviderID); exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "oauth account already linked"})
		return
	}
	if existingUser, exists := a.store.GetUserByEmail(strings.ToLower(pending.Email)); exists {
		a.store.CreateOAuthAccount(models.OAuthAccount{
			UserID:     existingUser.ID,
			Provider:   pending.Provider,
			ProviderID: pending.ProviderID,
			Email:      strings.ToLower(pending.Email),
		})
		token, err := a.jwt.GenerateAppToken(existingUser)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
			return
		}
		writeJSON(w, http.StatusOK, loginResponse(existingUser, token))
		return
	}

	user, err := a.createOAuthUser(
		pending.Provider,
		pending.ProviderID,
		pending.Email,
		req.Name,
		pending.AvatarURL,
		role,
		req.TenantName,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	a.store.AddAuditLog(models.AuditLog{
		TenantID:   user.TenantID,
		ActorID:    user.ID,
		Action:     "auth.signup.oauth.complete",
		TargetType: "user",
		TargetID:   user.ID,
		Outcome:    "success",
		Metadata:   map[string]interface{}{"email": user.Email, "provider": string(pending.Provider), "role": roleName(user.Role)},
	})

	token, err := a.jwt.GenerateAppToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
		return
	}
	writeJSON(w, http.StatusOK, loginResponse(user, token))
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	user, ok := a.store.GetUserByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if !ok {
		a.store.AddAuditLog(models.AuditLog{
			TenantID:   "public",
			ActorID:    "anonymous",
			Action:     "auth.login.failed",
			TargetType: "user",
			TargetID:   "",
			Outcome:    "failure",
			Metadata:   map[string]interface{}{"email": strings.ToLower(strings.TrimSpace(req.Email)), "reason": "user_not_found"},
		})
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		a.store.AddAuditLog(models.AuditLog{
			TenantID:   user.TenantID,
			ActorID:    user.ID,
			Action:     "auth.login.failed",
			TargetType: "user",
			TargetID:   user.ID,
			Outcome:    "failure",
			Metadata:   map[string]interface{}{"email": user.Email, "reason": "invalid_password"},
		})
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	if user.MFAEnabled {
		code := strings.TrimSpace(req.MFACode)
		if code == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "mfa_required"})
			return
		}
		if !totp.Validate(code, user.MFATOTPSecret) {
			a.store.AddAuditLog(models.AuditLog{
				TenantID:   user.TenantID,
				ActorID:    user.ID,
				Action:     "auth.login.failed",
				TargetType: "user",
				TargetID:   user.ID,
				Outcome:    "failure",
				Metadata:   map[string]interface{}{"email": user.Email, "reason": "invalid_mfa_code"},
			})
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid mfa code"})
			return
		}
	}
	a.store.AddAuditLog(models.AuditLog{
		TenantID:   user.TenantID,
		ActorID:    user.ID,
		Action:     "auth.login.success",
		TargetType: "user",
		TargetID:   user.ID,
		Outcome:    "success",
		Metadata:   map[string]interface{}{"email": user.Email},
	})
	token, err := a.jwt.GenerateAppToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
		return
	}
	writeJSON(w, http.StatusOK, loginResponse(user, token))
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func roleName(role models.Role) string {
	return string(role)
}

// oauthGoogleStart initiates Google OAuth flow
func (a *App) oauthGoogleStart(w http.ResponseWriter, r *http.Request) {
	integrations := a.resolvedIntegrations()
	if integrations.Google.ClientID == "" || integrations.Google.ClientSecret == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Google OAuth not configured"})
		return
	}

	intent, err := parseOAuthSignupIntent(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Generate state and nonce for CSRF and replay protection
	state := randomToken("state_")
	nonce := randomToken("nonce_")
	redirectBase := integrations.Google.RedirectBase
	if redirectBase == "" {
		redirectBase = a.cfg.BaseURL
	}
	callbackURL := strings.TrimRight(redirectBase, "/") + "/auth/oauth/google/callback"

	oauthCfg := &oauth2.Config{
		ClientID:     integrations.Google.ClientID,
		ClientSecret: integrations.Google.ClientSecret,
		RedirectURL:  callbackURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile", "openid"},
		Endpoint:     google.Endpoint,
	}

	// Store state for validation with expiry (5 minutes)
	a.oauth.mu.Lock()
	a.oauth.states[state] = oauthStateData{
		Provider:    models.OAuthProviderGoogle,
		Nonce:       nonce,
		CreatedAt:   time.Now().UTC(),
		CallbackURL: callbackURL,
		Mode:        intent.Mode,
		Role:        intent.Role,
		Name:        intent.Name,
		TenantName:  intent.TenantName,
	}
	a.oauth.mu.Unlock()

	// Clean up old states (older than 10 minutes)
	a.oauth.mu.Lock()
	for k, v := range a.oauth.states {
		if time.Since(v.CreatedAt) > 10*time.Minute {
			delete(a.oauth.states, k)
		}
	}
	a.oauth.mu.Unlock()

	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline, oauth2.ApprovalForce)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"authorization_url": authURL,
		"state":             state,
		"nonce":             nonce,
	})
}

// oauthGoogleCallback handles Google OAuth callback
func (a *App) oauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	nonce := r.URL.Query().Get("nonce") // Optional: can be used for additional validation

	if code == "" || state == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing code or state"})
		return
	}

	// Validate state
	a.oauth.mu.Lock()
	stateData, ok := a.oauth.states[state]
	if !ok || stateData.Provider != models.OAuthProviderGoogle {
		a.oauth.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired state"})
		return
	}
	// Validate nonce if provided
	if nonce != "" && stateData.Nonce != nonce {
		a.oauth.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nonce mismatch"})
		return
	}
	// Check state expiry (5 minutes max)
	if time.Since(stateData.CreatedAt) > 5*time.Minute {
		delete(a.oauth.states, state)
		a.oauth.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "state expired"})
		return
	}
	delete(a.oauth.states, state)
	a.oauth.mu.Unlock()

	integrations := a.resolvedIntegrations()
	callbackURL := strings.TrimSpace(stateData.CallbackURL)
	if callbackURL == "" {
		redirectBase := integrations.Google.RedirectBase
		if redirectBase == "" {
			redirectBase = a.cfg.BaseURL
		}
		callbackURL = strings.TrimRight(redirectBase, "/") + "/auth/oauth/google/callback"
	}

	oauthCfg := &oauth2.Config{
		ClientID:     integrations.Google.ClientID,
		ClientSecret: integrations.Google.ClientSecret,
		RedirectURL:  callbackURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile", "openid"},
		Endpoint:     google.Endpoint,
	}

	token, err := oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to exchange code"})
		return
	}

	// Get user info
	client := oauthCfg.Client(r.Context(), token)
	userInfoResp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user info"})
		return
	}
	defer userInfoResp.Body.Close()

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfo); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to parse user info"})
		return
	}

	if userInfo.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no email from OAuth provider"})
		return
	}

	// Find or create user by OAuth account (provider + subject)
	user, isNewUser, pendingSignup, err := a.findOrCreateOAuthUser(
		models.OAuthProviderGoogle,
		userInfo.ID,
		userInfo.Email,
		userInfo.Name,
		userInfo.Picture,
		stateData.Role,
		stateData.Name,
		stateData.TenantName,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if pendingSignup != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"signupRequired": true,
			"signupToken":    pendingSignup.Token,
			"profile": map[string]interface{}{
				"email":     pendingSignup.Email,
				"name":      pendingSignup.Name,
				"avatarUrl": pendingSignup.AvatarURL,
			},
		})
		return
	}

	// Generate JWT token
	jwtToken, err := a.jwt.GenerateAppToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
		return
	}

	action := "auth.login.oauth.google"
	if isNewUser {
		action = "auth.signup.oauth.google"
	}

	a.store.AddAuditLog(models.AuditLog{
		TenantID:   user.TenantID,
		ActorID:    user.ID,
		Action:     action,
		TargetType: "user",
		TargetID:   user.ID,
		Outcome:    "success",
		Metadata:   map[string]interface{}{"email": user.Email, "provider": "google", "isNewUser": isNewUser},
	})

	// Return success response matching loginResponse format
	a.respondOAuthSuccess(w, r, user, jwtToken)
}

// oauthGitHubStart initiates GitHub OAuth flow
func (a *App) oauthGitHubStart(w http.ResponseWriter, r *http.Request) {
	integrations := a.resolvedIntegrations()
	if integrations.GitHub.ClientID == "" || integrations.GitHub.ClientSecret == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "GitHub OAuth not configured"})
		return
	}

	intent, err := parseOAuthSignupIntent(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Generate state and nonce for CSRF protection
	state := randomToken("state_")
	nonce := randomToken("nonce_")
	redirectBase := integrations.GitHub.RedirectBase
	if redirectBase == "" {
		redirectBase = a.cfg.BaseURL
	}
	callbackURL := strings.TrimRight(redirectBase, "/") + "/auth/oauth/github/callback"

	oauthCfg := &oauth2.Config{
		ClientID:     integrations.GitHub.ClientID,
		ClientSecret: integrations.GitHub.ClientSecret,
		RedirectURL:  callbackURL,
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}

	// Store state for validation with expiry
	a.oauth.mu.Lock()
	a.oauth.states[state] = oauthStateData{
		Provider:    models.OAuthProviderGitHub,
		Nonce:       nonce,
		CreatedAt:   time.Now().UTC(),
		CallbackURL: callbackURL,
		Mode:        intent.Mode,
		Role:        intent.Role,
		Name:        intent.Name,
		TenantName:  intent.TenantName,
	}
	a.oauth.mu.Unlock()

	// Clean up old states
	a.oauth.mu.Lock()
	for k, v := range a.oauth.states {
		if time.Since(v.CreatedAt) > 10*time.Minute {
			delete(a.oauth.states, k)
		}
	}
	a.oauth.mu.Unlock()

	authURL := oauthCfg.AuthCodeURL(state)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"authorization_url": authURL,
		"state":             state,
		"nonce":             nonce,
	})
}

// oauthGitHubCallback handles GitHub OAuth callback
func (a *App) oauthGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	nonce := r.URL.Query().Get("nonce")

	if code == "" || state == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing code or state"})
		return
	}

	// Validate state
	a.oauth.mu.Lock()
	stateData, ok := a.oauth.states[state]
	if !ok || stateData.Provider != models.OAuthProviderGitHub {
		a.oauth.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired state"})
		return
	}
	if nonce != "" && stateData.Nonce != nonce {
		a.oauth.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nonce mismatch"})
		return
	}
	if time.Since(stateData.CreatedAt) > 5*time.Minute {
		delete(a.oauth.states, state)
		a.oauth.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "state expired"})
		return
	}
	delete(a.oauth.states, state)
	a.oauth.mu.Unlock()

	integrations := a.resolvedIntegrations()
	callbackURL := strings.TrimSpace(stateData.CallbackURL)
	if callbackURL == "" {
		redirectBase := integrations.GitHub.RedirectBase
		if redirectBase == "" {
			redirectBase = a.cfg.BaseURL
		}
		callbackURL = strings.TrimRight(redirectBase, "/") + "/auth/oauth/github/callback"
	}

	oauthCfg := &oauth2.Config{
		ClientID:     integrations.GitHub.ClientID,
		ClientSecret: integrations.GitHub.ClientSecret,
		RedirectURL:  callbackURL,
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}

	token, err := oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to exchange code"})
		return
	}

	// Get user info from GitHub
	client := oauthCfg.Client(r.Context(), token)

	// Get user email
	emailResp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user emails"})
		return
	}
	defer emailResp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to parse emails"})
		return
	}

	var userEmail string
	for _, e := range emails {
		if e.Primary && e.Verified {
			userEmail = e.Email
			break
		}
	}
	if userEmail == "" && len(emails) > 0 {
		userEmail = emails[0].Email
	}

	if userEmail == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no verified email found"})
		return
	}

	// Get user name and ID
	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user info"})
		return
	}
	defer userResp.Body.Close()

	var userInfo struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to parse user info"})
		return
	}

	userName := userInfo.Name
	if userName == "" {
		userName = userInfo.Login
	}
	providerID := fmt.Sprintf("%d", userInfo.ID)

	// Find or create user by OAuth account (provider + subject)
	user, isNewUser, pendingSignup, err := a.findOrCreateOAuthUser(
		models.OAuthProviderGitHub,
		providerID,
		userEmail,
		userName,
		"",
		stateData.Role,
		stateData.Name,
		stateData.TenantName,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if pendingSignup != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"signupRequired": true,
			"signupToken":    pendingSignup.Token,
			"profile": map[string]interface{}{
				"email":     pendingSignup.Email,
				"name":      pendingSignup.Name,
				"avatarUrl": pendingSignup.AvatarURL,
			},
		})
		return
	}

	// Generate JWT token
	jwtToken, err := a.jwt.GenerateAppToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
		return
	}

	action := "auth.login.oauth.github"
	if isNewUser {
		action = "auth.signup.oauth.github"
	}

	a.store.AddAuditLog(models.AuditLog{
		TenantID:   user.TenantID,
		ActorID:    user.ID,
		Action:     action,
		TargetType: "user",
		TargetID:   user.ID,
		Outcome:    "success",
		Metadata:   map[string]interface{}{"email": user.Email, "provider": "github", "isNewUser": isNewUser},
	})

	// Return success response matching loginResponse format
	a.respondOAuthSuccess(w, r, user, jwtToken)
}

func (a *App) respondOAuthSuccess(w http.ResponseWriter, r *http.Request, user models.User, token string) {
	secure := false
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(aBaseURL(r))), "https://") {
		secure = true
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "mcp_access_token",
		Value:    token,
		Path:     "/",
		MaxAge:   appSessionMaxAgeSeconds,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "mcp_active_role",
		Value:    roleForUser(user),
		Path:     "/",
		MaxAge:   appSessionMaxAgeSeconds,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
	target := strings.TrimRight(a.frontendBaseURLFromRequest(r), "/") + "/login?oauth=success"
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

func (a *App) frontendBaseURLFromRequest(r *http.Request) string {
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		return origin
	}
	if referer := strings.TrimSpace(r.Header.Get("Referer")); referer != "" {
		if parsed, err := url.Parse(referer); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			return parsed.Scheme + "://" + parsed.Host
		}
	}
	if preferred := preferredFrontendOrigin(a.cfg.CORSAllowedOrigins); preferred != "" {
		return preferred
	}
	return strings.TrimRight(strings.TrimSpace(a.cfg.BaseURL), "/")
}

func preferredFrontendOrigin(origins []string) string {
	var first string
	for _, origin := range origins {
		origin = strings.TrimRight(strings.TrimSpace(origin), "/")
		if origin == "" {
			continue
		}
		if first == "" {
			first = origin
		}
		lower := strings.ToLower(origin)
		if strings.Contains(lower, "localhost") || strings.Contains(lower, "127.0.0.1") {
			continue
		}
		if strings.HasPrefix(lower, "https://") {
			return origin
		}
	}
	return first
}

func aBaseURL(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		return ""
	}
	return scheme + "://" + host
}

// findOrCreateOAuthUser finds or creates a user based on OAuth provider and subject ID
// This implements proper account linking: find by provider+subject, then by email
func parseOAuthSignupIntent(r *http.Request) (oauthSignupIntent, error) {
	mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	if mode == "" {
		mode = "login"
	}

	intent := oauthSignupIntent{
		Mode:       mode,
		Role:       models.RoleBuyer,
		Name:       strings.TrimSpace(r.URL.Query().Get("name")),
		TenantName: strings.TrimSpace(r.URL.Query().Get("tenantName")),
	}

	if mode == "signup" {
		intent.Role = normalizeRole(r.URL.Query().Get("role"))
		if intent.Role == models.RoleAdmin {
			return oauthSignupIntent{}, fmt.Errorf("admin signup is disabled; bootstrap super admin only")
		}
	}

	return intent, nil
}

func (a *App) consumePendingOAuthSignup(token string) (pendingOAuthSignup, bool) {
	a.oauth.mu.Lock()
	defer a.oauth.mu.Unlock()
	pending, ok := a.oauth.pending[token]
	if !ok {
		return pendingOAuthSignup{}, false
	}
	delete(a.oauth.pending, token)
	if time.Since(pending.CreatedAt) > 10*time.Minute {
		return pendingOAuthSignup{}, false
	}
	return pending, true
}

func (a *App) savePendingOAuthSignup(provider models.OAuthProvider, providerID, email, name, avatarURL string) pendingOAuthSignup {
	a.oauth.mu.Lock()
	defer a.oauth.mu.Unlock()
	now := time.Now().UTC()
	for key, item := range a.oauth.pending {
		if time.Since(item.CreatedAt) > 10*time.Minute {
			delete(a.oauth.pending, key)
		}
	}
	token := randomToken("oauth_signup_")
	item := pendingOAuthSignup{
		Token:      token,
		Provider:   provider,
		ProviderID: providerID,
		Email:      strings.ToLower(strings.TrimSpace(email)),
		Name:       strings.TrimSpace(name),
		AvatarURL:  strings.TrimSpace(avatarURL),
		CreatedAt:  now,
	}
	a.oauth.pending[token] = item
	return item
}

func (a *App) createOAuthUser(
	provider models.OAuthProvider,
	providerID, email, displayName, avatarURL string,
	role models.Role,
	tenantName string,
) (models.User, error) {
	tenant := a.store.CreateTenant(models.Tenant{
		Name:     tenantName,
		Slug:     slugifyTenant(tenantName),
		PlanTier: "professional",
		Status:   "active",
	})

	user, created := a.store.CreateUser(models.User{
		TenantID:  tenant.ID,
		Email:     strings.ToLower(email),
		Name:      displayName,
		AvatarURL: avatarURL,
		Role:      role,
		CreatedAt: time.Now().UTC(),
	})
	if !created {
		return models.User{}, fmt.Errorf("failed to create user")
	}

	tenant.OwnerUserID = user.ID
	_ = a.store.UpdateTenant(tenant)
	a.store.CreateOAuthAccount(models.OAuthAccount{
		UserID:     user.ID,
		Provider:   provider,
		ProviderID: providerID,
		Email:      strings.ToLower(email),
	})
	return user, nil
}

func (a *App) findOrCreateOAuthUser(
	provider models.OAuthProvider,
	providerID, email, name, avatarURL string,
	requestedRole models.Role,
	requestedName string,
	requestedTenantName string,
) (models.User, bool, *pendingOAuthSignup, error) {
	// Step 1: Check if OAuth account exists with this provider + providerID
	existingOAuth, exists := a.store.GetOAuthAccount(provider, providerID)
	if exists {
		// User already linked with this OAuth account
		user, userFound := a.store.GetUserByID(existingOAuth.UserID)
		if !userFound {
			// OAuth account exists but user doesn't - shouldn't happen, but handle it
			return models.User{}, false, nil, fmt.Errorf("oauth account orphaned")
		}
		// Update avatar if provided and different
		if avatarURL != "" && user.AvatarURL != avatarURL {
			user.AvatarURL = avatarURL
			a.store.UpdateUser(user)
		}
		return user, false, nil, nil
	}

	// Step 2: Check if user exists with the same email
	existingUser, userExists := a.store.GetUserByEmail(strings.ToLower(email))
	if userExists {
		// Link this OAuth account to existing user
		a.store.CreateOAuthAccount(models.OAuthAccount{
			UserID:     existingUser.ID,
			Provider:   provider,
			ProviderID: providerID,
			Email:      strings.ToLower(email),
		})
		// Update avatar if provided
		if avatarURL != "" && existingUser.AvatarURL != avatarURL {
			existingUser.AvatarURL = avatarURL
			a.store.UpdateUser(existingUser)
		}
		return existingUser, false, nil, nil
	}

	displayName := strings.TrimSpace(requestedName)
	if displayName == "" {
		displayName = strings.TrimSpace(name)
	}
	if displayName == "" {
		displayName = strings.TrimSpace(strings.Split(strings.ToLower(email), "@")[0])
	}
	tenantName := strings.TrimSpace(requestedTenantName)
	if tenantName == "" {
		tenantName = displayName
	}
	role := requestedRole
	if role == "" {
		role = models.RoleBuyer
	}
	_ = tenantName
	_ = role
	pending := a.savePendingOAuthSignup(provider, providerID, email, displayName, avatarURL)
	return models.User{}, false, &pending, nil
}
