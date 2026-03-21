package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type updatePlatformIntegrationsRequest struct {
	Google struct {
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
		RedirectBase string `json:"redirectBase"`
	} `json:"google"`
	GitHub struct {
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
		RedirectBase string `json:"redirectBase"`
	} `json:"github"`
	Stripe struct {
		PublishableKey       string  `json:"publishableKey"`
		SecretKey            string  `json:"secretKey"`
		WebhookSecret        string  `json:"webhookSecret"`
		OnrampReturnURL      string  `json:"onrampReturnUrl"`
		OnrampRefreshURL     string  `json:"onrampRefreshUrl"`
		OnrampMinUSD         float64 `json:"onrampMinUsd"`
		OnrampDefaultUSD     float64 `json:"onrampDefaultUsd"`
		ConnectReturnURL     string  `json:"connectReturnUrl"`
		ConnectRefreshURL    string  `json:"connectRefreshUrl"`
		ConnectWebhookSecret string  `json:"connectWebhookSecret"`
	} `json:"stripe"`
	Wallet struct {
		Provider                  string `json:"provider"`
		ManagedAutoPayEnabled     bool   `json:"managedAutoPayEnabled"`
		LegacyPaymentModeEnabled  bool   `json:"legacyPaymentModeEnabled"`
		ExternalWalletsEnabled    bool   `json:"externalWalletsEnabled"`
		CDPEnabled                bool   `json:"cdpEnabled"`
		FireflyEnabled            bool   `json:"fireflyEnabled"`
		CDPAPIKeyID               string `json:"cdpApiKeyId"`
		CDPAPIKeySecret           string `json:"cdpApiKeySecret"`
		CDPWalletSecret           string `json:"cdpWalletSecret"`
		FireflySignerURL          string `json:"fireflySignerUrl"`
		FireflyAuthToken          string `json:"fireflyAuthToken"`
		FireflyKeystoreDir        string `json:"fireflyKeystoreDir"`
		FireflyKeystorePassphrase string `json:"fireflyKeystorePassphrase"`
		DefaultNetwork            string `json:"defaultNetwork"`
		DefaultAsset              string `json:"defaultAsset"`
		CustodyMode               string `json:"custodyMode"`
	} `json:"wallet"`
	X402 struct {
		Mode              string `json:"mode"`
		FacilitatorURL    string `json:"facilitatorUrl"`
		FacilitatorAPIKey string `json:"facilitatorApiKey"`
	} `json:"x402"`
	N8N struct {
		BaseURL        string `json:"baseUrl"`
		APIKey         string `json:"apiKey"`
		TimeoutSeconds int    `json:"timeoutSeconds"`
	} `json:"n8n"`
}

func (a *App) getRuntimeConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.runtimePublicConfig())
}

func (a *App) adminIntegrations(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, a.integrationSettingsResponse())
		return
	}

	claims, _ := getClaims(r.Context())
	var req updatePlatformIntegrationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	settings, _ := a.storedPlatformIntegrationSettings()
	if strings.TrimSpace(settings.Key) == "" {
		settings.Key = "platform"
	}

	settings.Google.ClientID = strings.TrimSpace(req.Google.ClientID)
	if strings.TrimSpace(req.Google.ClientSecret) != "" {
		settings.Google.ClientSecret = strings.TrimSpace(req.Google.ClientSecret)
	}
	settings.Google.RedirectBase = strings.TrimSpace(req.Google.RedirectBase)

	settings.GitHub.ClientID = strings.TrimSpace(req.GitHub.ClientID)
	if strings.TrimSpace(req.GitHub.ClientSecret) != "" {
		settings.GitHub.ClientSecret = strings.TrimSpace(req.GitHub.ClientSecret)
	}
	settings.GitHub.RedirectBase = strings.TrimSpace(req.GitHub.RedirectBase)

	settings.Stripe.PublishableKey = strings.TrimSpace(req.Stripe.PublishableKey)
	if strings.TrimSpace(req.Stripe.SecretKey) != "" {
		settings.Stripe.SecretKey = strings.TrimSpace(req.Stripe.SecretKey)
	}
	if strings.TrimSpace(req.Stripe.WebhookSecret) != "" {
		settings.Stripe.WebhookSecret = strings.TrimSpace(req.Stripe.WebhookSecret)
	}
	settings.Stripe.OnrampReturnURL = strings.TrimSpace(req.Stripe.OnrampReturnURL)
	settings.Stripe.OnrampRefreshURL = strings.TrimSpace(req.Stripe.OnrampRefreshURL)
	if req.Stripe.OnrampMinUSD > 0 {
		settings.Stripe.OnrampMinUSD = req.Stripe.OnrampMinUSD
	}
	if req.Stripe.OnrampDefaultUSD > 0 {
		settings.Stripe.OnrampDefaultUSD = req.Stripe.OnrampDefaultUSD
	}
	settings.Stripe.ConnectReturnURL = strings.TrimSpace(req.Stripe.ConnectReturnURL)
	settings.Stripe.ConnectRefreshURL = strings.TrimSpace(req.Stripe.ConnectRefreshURL)
	if strings.TrimSpace(req.Stripe.ConnectWebhookSecret) != "" {
		settings.Stripe.ConnectWebhookSecret = strings.TrimSpace(req.Stripe.ConnectWebhookSecret)
	}

	if strings.TrimSpace(req.Wallet.Provider) != "" {
		settings.Wallet.Provider = strings.ToLower(strings.TrimSpace(req.Wallet.Provider))
	}
	settings.Wallet.ManagedAutoPayEnabled = req.Wallet.ManagedAutoPayEnabled
	settings.Wallet.LegacyPaymentModeEnabled = req.Wallet.LegacyPaymentModeEnabled
	settings.Wallet.ExternalWalletsEnabled = req.Wallet.ExternalWalletsEnabled
	settings.Wallet.CDPEnabled = req.Wallet.CDPEnabled
	settings.Wallet.FireflyEnabled = req.Wallet.FireflyEnabled
	if strings.TrimSpace(req.Wallet.CDPAPIKeyID) != "" {
		settings.Wallet.CDPAPIKeyID = strings.TrimSpace(req.Wallet.CDPAPIKeyID)
	}
	if strings.TrimSpace(req.Wallet.CDPAPIKeySecret) != "" {
		settings.Wallet.CDPAPIKeySecret = strings.TrimSpace(req.Wallet.CDPAPIKeySecret)
	}
	if strings.TrimSpace(req.Wallet.CDPWalletSecret) != "" {
		settings.Wallet.CDPWalletSecret = strings.TrimSpace(req.Wallet.CDPWalletSecret)
	}
	if strings.TrimSpace(req.Wallet.FireflySignerURL) != "" {
		settings.Wallet.FireflySignerURL = strings.TrimSpace(req.Wallet.FireflySignerURL)
	}
	if strings.TrimSpace(req.Wallet.FireflyAuthToken) != "" {
		settings.Wallet.FireflyAuthToken = strings.TrimSpace(req.Wallet.FireflyAuthToken)
	}
	if strings.TrimSpace(req.Wallet.FireflyKeystoreDir) != "" {
		settings.Wallet.FireflyKeystoreDir = strings.TrimSpace(req.Wallet.FireflyKeystoreDir)
	}
	if strings.TrimSpace(req.Wallet.FireflyKeystorePassphrase) != "" {
		settings.Wallet.FireflyKeystorePassphrase = strings.TrimSpace(req.Wallet.FireflyKeystorePassphrase)
	}
	if strings.TrimSpace(req.Wallet.DefaultNetwork) != "" {
		settings.Wallet.DefaultNetwork = strings.TrimSpace(req.Wallet.DefaultNetwork)
	}
	if strings.TrimSpace(req.Wallet.DefaultAsset) != "" {
		settings.Wallet.DefaultAsset = strings.TrimSpace(req.Wallet.DefaultAsset)
	}
	if strings.TrimSpace(req.Wallet.CustodyMode) != "" {
		settings.Wallet.CustodyMode = strings.TrimSpace(req.Wallet.CustodyMode)
	}

	mode := strings.ToLower(strings.TrimSpace(req.X402.Mode))
	if mode != "" {
		settings.X402.Mode = mode
	}
	settings.X402.FacilitatorURL = strings.TrimSpace(req.X402.FacilitatorURL)
	if strings.TrimSpace(req.X402.FacilitatorAPIKey) != "" {
		settings.X402.FacilitatorAPIKey = strings.TrimSpace(req.X402.FacilitatorAPIKey)
	}

	settings.N8N.BaseURL = strings.TrimSpace(req.N8N.BaseURL)
	if strings.TrimSpace(req.N8N.APIKey) != "" {
		settings.N8N.APIKey = strings.TrimSpace(req.N8N.APIKey)
	}
	if req.N8N.TimeoutSeconds > 0 {
		settings.N8N.TimeoutSeconds = req.N8N.TimeoutSeconds
	}

	settings.UpdatedBy = claims.UserID
	saved := a.store.UpsertPlatformIntegrationSettings(settings)
	a.store.AddAuditLog(models.AuditLog{
		TenantID:   claims.TenantID,
		ActorID:    claims.UserID,
		Action:     "admin.integrations.updated",
		TargetType: "platform_integrations",
		TargetID:   saved.Key,
		Outcome:    "success",
		Metadata: map[string]interface{}{
			"googleConfigured": strings.TrimSpace(saved.Google.ClientID) != "" && strings.TrimSpace(saved.Google.ClientSecret) != "",
			"githubConfigured": strings.TrimSpace(saved.GitHub.ClientID) != "" && strings.TrimSpace(saved.GitHub.ClientSecret) != "",
			"stripeConfigured": strings.TrimSpace(saved.Stripe.SecretKey) != "",
			"walletConfigured": strings.TrimSpace(saved.Wallet.Provider) != "",
			"walletProvider":   saved.Wallet.Provider,
			"walletAutoPay":    saved.Wallet.ManagedAutoPayEnabled,
			"walletLegacy":     saved.Wallet.LegacyPaymentModeEnabled,
			"x402Mode":         saved.X402.Mode,
			"n8nConfigured":    strings.TrimSpace(saved.N8N.BaseURL) != "",
		},
	})
	writeJSON(w, http.StatusOK, a.integrationSettingsResponse())
}
