package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signupRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	TenantName string `json:"tenantName"`
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

	token, err := a.jwt.Generate(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
		return
	}
	writeJSON(w, http.StatusCreated, loginResponse(user, token))
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	user, ok := a.store.GetUserByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	token, err := a.jwt.Generate(user)
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
