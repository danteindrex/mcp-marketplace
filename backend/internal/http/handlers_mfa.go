package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pquerna/otp/totp"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type verifyMFARequest struct {
	Code string `json:"code"`
}

type disableMFARequest struct {
	CurrentPassword string `json:"currentPassword"`
	Code            string `json:"code"`
}

func (a *App) getMFAStatus(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	method := ""
	if user.MFAEnabled {
		method = "totp"
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"mfaEnabled": user.MFAEnabled,
		"method":     method,
	})
}

func (a *App) setupTOTP(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "MCP Marketplace",
		AccountName: user.Email,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate totp secret"})
		return
	}
	user.MFATOTPSecret = key.Secret()
	user.MFAEnabled = false
	if ok := a.store.UpdateUser(user); !ok {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "failed to update user mfa settings"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"secret":      key.Secret(),
		"otpauthURL":  key.URL(),
		"issuer":      "MCP Marketplace",
		"accountName": user.Email,
	})
}

func (a *App) verifyTOTP(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	if strings.TrimSpace(user.MFATOTPSecret) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "mfa setup is not initialized"})
		return
	}
	var req verifyMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	code := strings.TrimSpace(req.Code)
	if code == "" || !totp.Validate(code, user.MFATOTPSecret) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid code"})
		return
	}
	user.MFAEnabled = true
	if ok := a.store.UpdateUser(user); !ok {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "failed to enable mfa"})
		return
	}
	a.store.AddAuditLog(models.AuditLog{
		TenantID:   claims.TenantID,
		ActorID:    claims.UserID,
		Action:     "settings.mfa.enable",
		TargetType: "user",
		TargetID:   claims.UserID,
		Outcome:    "success",
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{"mfaEnabled": true, "method": "totp"})
}

func (a *App) disableTOTP(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	var req disableMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.CurrentPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "currentPassword is required"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "current password is invalid"})
		return
	}
	if user.MFAEnabled {
		code := strings.TrimSpace(req.Code)
		if code == "" || !totp.Validate(code, user.MFATOTPSecret) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid mfa code is required"})
			return
		}
	}
	user.MFAEnabled = false
	user.MFATOTPSecret = ""
	if ok := a.store.UpdateUser(user); !ok {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "failed to disable mfa"})
		return
	}
	a.store.AddAuditLog(models.AuditLog{
		TenantID:   claims.TenantID,
		ActorID:    claims.UserID,
		Action:     "settings.mfa.disable",
		TargetType: "user",
		TargetID:   claims.UserID,
		Outcome:    "success",
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{"mfaEnabled": false, "method": ""})
}
