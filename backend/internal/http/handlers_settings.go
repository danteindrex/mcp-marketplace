package http

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type updateProfileRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatarUrl"`
	Locale    string `json:"locale"`
	Timezone  string `json:"timezone"`
}

type updatePreferencesRequest struct {
	Theme          string `json:"theme"`
	Language       string `json:"language"`
	Timezone       string `json:"timezone"`
	DefaultLanding string `json:"defaultLanding"`
	CompactMode    *bool  `json:"compactMode"`
}

type updateNotificationsRequest struct {
	ProductUpdates *bool `json:"productUpdates"`
	SecurityAlerts *bool `json:"securityAlerts"`
	BillingAlerts  *bool `json:"billingAlerts"`
	MarketingEmail *bool `json:"marketingEmail"`
	WeeklyDigest   *bool `json:"weeklyDigest"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func (a *App) getUserProfile(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"profile": map[string]interface{}{
			"id":        user.ID,
			"tenantId":  user.TenantID,
			"email":     user.Email,
			"name":      user.Name,
			"phone":     user.Phone,
			"avatarUrl": user.AvatarURL,
			"locale":    user.Locale,
			"timezone":  user.Timezone,
			"role":      user.Role,
		},
	})
}

func (a *App) updateUserProfile(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if name == "" || email == "" || !emailPattern.MatchString(email) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid name and email are required"})
		return
	}

	if existing, ok := a.store.GetUserByEmail(email); ok && existing.ID != user.ID {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already in use"})
		return
	}

	user.Name = name
	user.Email = email
	user.Phone = strings.TrimSpace(req.Phone)
	user.AvatarURL = strings.TrimSpace(req.AvatarURL)
	if strings.TrimSpace(req.Locale) != "" {
		user.Locale = strings.TrimSpace(req.Locale)
	}
	if strings.TrimSpace(req.Timezone) != "" {
		user.Timezone = strings.TrimSpace(req.Timezone)
	}

	if ok := a.store.UpdateUser(user); !ok {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "unable to update profile"})
		return
	}

	if settings, ok := a.store.GetUserSettings(user.ID); ok {
		settings.Preferences.Timezone = user.Timezone
		a.store.UpsertUserSettings(settings)
	}
	a.store.AddAuditLog(models.AuditLog{
		TenantID: claims.TenantID,
		ActorID:  claims.UserID,
		Action:   "settings.profile.update",
		TargetType: "user",
		TargetID: claims.UserID,
		Outcome:  "success",
		Metadata: map[string]interface{}{"email": user.Email},
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"profile": map[string]interface{}{
			"id":        user.ID,
			"tenantId":  user.TenantID,
			"email":     user.Email,
			"name":      user.Name,
			"phone":     user.Phone,
			"avatarUrl": user.AvatarURL,
			"locale":    user.Locale,
			"timezone":  user.Timezone,
			"role":      user.Role,
		},
	})
}

func defaultSettingsFor(user models.User) models.UserSettings {
	defaultLanding := "/buyer/dashboard"
	if user.Role == models.RoleMerchant {
		defaultLanding = "/merchant/onboarding"
	}
	if user.Role == models.RoleAdmin {
		defaultLanding = "/admin/tenants"
	}
	timezone := user.Timezone
	if timezone == "" {
		timezone = "America/Los_Angeles"
	}
	return models.UserSettings{
		UserID: user.ID,
		Preferences: models.UserPreferences{
			Theme:          "system",
			Language:       "en",
			Timezone:       timezone,
			DefaultLanding: defaultLanding,
			CompactMode:    false,
		},
		Notifications: models.NotificationSettings{
			ProductUpdates: true,
			SecurityAlerts: true,
			BillingAlerts:  true,
			MarketingEmail: false,
			WeeklyDigest:   true,
		},
	}
}

func (a *App) getOrCreateSettings(user models.User) models.UserSettings {
	if settings, ok := a.store.GetUserSettings(user.ID); ok {
		return settings
	}
	return a.store.UpsertUserSettings(defaultSettingsFor(user))
}

func (a *App) getUserPreferences(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	settings := a.getOrCreateSettings(user)
	writeJSON(w, http.StatusOK, map[string]interface{}{"preferences": settings.Preferences})
}

func (a *App) updateUserPreferences(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	settings := a.getOrCreateSettings(user)

	var req updatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	switch req.Theme {
	case "", "light", "dark", "system":
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "theme must be light, dark, or system"})
		return
	}

	if req.Theme != "" {
		settings.Preferences.Theme = req.Theme
	}
	if strings.TrimSpace(req.Language) != "" {
		settings.Preferences.Language = strings.TrimSpace(req.Language)
	}
	if strings.TrimSpace(req.Timezone) != "" {
		settings.Preferences.Timezone = strings.TrimSpace(req.Timezone)
		user.Timezone = settings.Preferences.Timezone
		a.store.UpdateUser(user)
	}
	if strings.TrimSpace(req.DefaultLanding) != "" {
		settings.Preferences.DefaultLanding = strings.TrimSpace(req.DefaultLanding)
	}
	if req.CompactMode != nil {
		settings.Preferences.CompactMode = *req.CompactMode
	}

	settings = a.store.UpsertUserSettings(settings)
	writeJSON(w, http.StatusOK, map[string]interface{}{"preferences": settings.Preferences})
}

func (a *App) getUserNotifications(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	settings := a.getOrCreateSettings(user)
	writeJSON(w, http.StatusOK, map[string]interface{}{"notifications": settings.Notifications})
}

func (a *App) updateUserNotifications(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	settings := a.getOrCreateSettings(user)

	var req updateNotificationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.ProductUpdates != nil {
		settings.Notifications.ProductUpdates = *req.ProductUpdates
	}
	if req.SecurityAlerts != nil {
		settings.Notifications.SecurityAlerts = *req.SecurityAlerts
	}
	if req.BillingAlerts != nil {
		settings.Notifications.BillingAlerts = *req.BillingAlerts
	}
	if req.MarketingEmail != nil {
		settings.Notifications.MarketingEmail = *req.MarketingEmail
	}
	if req.WeeklyDigest != nil {
		settings.Notifications.WeeklyDigest = *req.WeeklyDigest
	}

	settings = a.store.UpsertUserSettings(settings)
	writeJSON(w, http.StatusOK, map[string]interface{}{"notifications": settings.Notifications})
}

func (a *App) changeUserPassword(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	user, found := a.store.GetUserByID(claims.UserID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "currentPassword and newPassword are required"})
		return
	}
	if !isStrongPassword(req.NewPassword) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "new password must be at least 12 characters and include upper, lower, number, and symbol"})
		return
	}
	if req.ConfirmPassword != "" && req.ConfirmPassword != req.NewPassword {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "confirmPassword does not match"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "current password is invalid"})
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}
	user.PasswordHash = string(newHash)
	a.store.UpdateUser(user)
	a.store.AddAuditLog(models.AuditLog{
		TenantID:   claims.TenantID,
		ActorID:    claims.UserID,
		Action:     "settings.password.change",
		TargetType: "user",
		TargetID:   claims.UserID,
		Outcome:    "success",
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
