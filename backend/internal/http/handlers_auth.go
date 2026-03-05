package http

import (
	"encoding/json"
	"net/http"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type loginRequest struct {
	Email string `json:"email"`
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	user, ok := a.store.GetUserByEmail(req.Email)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unknown user"})
		return
	}
	token, err := a.jwt.Generate(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
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
	})
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
