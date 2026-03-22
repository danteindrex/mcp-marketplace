package http

import (
	"strings"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type resolvedIntegrations struct {
	Google models.OAuthProviderSettings
	GitHub models.OAuthProviderSettings
	Stripe models.StripeIntegrationSettings
	Wallet models.WalletIntegrationSettings
	X402   models.X402IntegrationSettings
	N8N    models.N8NIntegrationSettings
}

type integrationStatus struct {
	Configured bool   `json:"configured"`
	Source     string `json:"source"`
	Notes      string `json:"notes,omitempty"`
}

type integrationSecretField struct {
	Set    bool   `json:"set"`
	Masked string `json:"masked,omitempty"`
}

func (a *App) storedPlatformIntegrationSettings() (models.PlatformIntegrationSettings, bool) {
	settings, ok := a.store.GetPlatformIntegrationSettings()
	if !ok || strings.TrimSpace(settings.Key) == "" {
		return models.PlatformIntegrationSettings{}, false
	}
	return settings, true
}

func (a *App) resolvedIntegrations() resolvedIntegrations {
	stored, _ := a.storedPlatformIntegrationSettings()
	out := resolvedIntegrations{
		Google: models.OAuthProviderSettings{
			ClientID:     strings.TrimSpace(a.cfg.GoogleClientID),
			ClientSecret: strings.TrimSpace(a.cfg.GoogleClientSecret),
			RedirectBase: strings.TrimSpace(a.cfg.OAuthRedirectBase),
		},
		GitHub: models.OAuthProviderSettings{
			ClientID:     strings.TrimSpace(a.cfg.GitHubClientID),
			ClientSecret: strings.TrimSpace(a.cfg.GitHubClientSecret),
			RedirectBase: strings.TrimSpace(a.cfg.OAuthRedirectBase),
		},
		Stripe: models.StripeIntegrationSettings{
			SecretKey:            strings.TrimSpace(a.cfg.StripeSecretKey),
			WebhookSecret:        strings.TrimSpace(a.cfg.StripeWebhookSecret),
			OnrampReturnURL:      strings.TrimSpace(a.cfg.StripeOnrampReturnURL),
			OnrampRefreshURL:     strings.TrimSpace(a.cfg.StripeOnrampRefreshURL),
			OnrampMinUSD:         a.cfg.StripeOnrampMinUSD,
			OnrampDefaultUSD:     a.cfg.StripeOnrampDefaultUSD,
			ConnectReturnURL:     strings.TrimSpace(a.cfg.StripeConnectReturnURL),
			ConnectRefreshURL:    strings.TrimSpace(a.cfg.StripeConnectRefreshURL),
			ConnectWebhookSecret: strings.TrimSpace(a.cfg.StripeConnectWebhookSecret),
		},
		Wallet: models.WalletIntegrationSettings{
			Provider:                  strings.TrimSpace(a.cfg.WalletProvider),
			ManagedAutoPayEnabled:     a.cfg.WalletManagedAutoPayEnabled,
			LegacyPaymentModeEnabled:  a.cfg.WalletLegacyPaymentModeEnabled,
			ExternalWalletsEnabled:    a.cfg.WalletExternalWalletsEnabled,
			CDPEnabled:                a.cfg.WalletCDPEnabled,
			FireflyEnabled:            a.cfg.WalletFireflyEnabled,
			CDPAPIKeyID:               strings.TrimSpace(a.cfg.CDPAPIKeyID),
			CDPAPIKeySecret:           strings.TrimSpace(a.cfg.CDPAPIKeySecret),
			CDPWalletSecret:           strings.TrimSpace(a.cfg.CDPWalletSecret),
			FireflySignerURL:          strings.TrimSpace(a.cfg.FireflySignerURL),
			FireflyAuthToken:          strings.TrimSpace(a.cfg.FireflyAuthToken),
			FireflyKeystoreDir:        strings.TrimSpace(a.cfg.FireflyKeystoreDir),
			FireflyKeystorePassphrase: strings.TrimSpace(a.cfg.FireflyKeystorePassphrase),
			DefaultNetwork:            strings.TrimSpace(a.cfg.WalletDefaultNetwork),
			DefaultAsset:              strings.TrimSpace(a.cfg.WalletDefaultAsset),
			CustodyMode:               strings.TrimSpace(a.cfg.WalletCustodyMode),
		},
		X402: models.X402IntegrationSettings{
			Mode:              strings.TrimSpace(a.cfg.X402Mode),
			FacilitatorURL:    strings.TrimSpace(a.cfg.X402FacilitatorURL),
			FacilitatorAPIKey: strings.TrimSpace(a.cfg.X402FacilitatorAPIKey),
		},
		N8N: models.N8NIntegrationSettings{
			BaseURL:        strings.TrimSpace(a.cfg.N8NBaseURL),
			APIKey:         strings.TrimSpace(a.cfg.N8NAPIKey),
			TimeoutSeconds: a.cfg.N8NTimeoutSeconds,
		},
	}
	mergeOAuthSettings(&out.Google, stored.Google)
	mergeOAuthSettings(&out.GitHub, stored.GitHub)
	mergeStripeSettings(&out.Stripe, stored.Stripe)
	mergeWalletSettings(&out.Wallet, stored.Wallet)
	out.Wallet = normalizeWalletSettings(out.Wallet)
	mergeX402Settings(&out.X402, stored.X402)
	mergeN8NSettings(&out.N8N, stored.N8N, a.cfg)
	return out
}

func mergeWalletSettings(dst *models.WalletIntegrationSettings, src models.WalletIntegrationSettings) {
	if strings.TrimSpace(src.Provider) != "" {
		dst.Provider = strings.TrimSpace(src.Provider)
	}
	dst.ManagedAutoPayEnabled = src.ManagedAutoPayEnabled
	dst.LegacyPaymentModeEnabled = src.LegacyPaymentModeEnabled
	dst.ExternalWalletsEnabled = src.ExternalWalletsEnabled
	dst.CDPEnabled = src.CDPEnabled
	dst.FireflyEnabled = src.FireflyEnabled
	if strings.TrimSpace(src.CDPAPIKeyID) != "" {
		dst.CDPAPIKeyID = strings.TrimSpace(src.CDPAPIKeyID)
	}
	if strings.TrimSpace(src.CDPAPIKeySecret) != "" {
		dst.CDPAPIKeySecret = strings.TrimSpace(src.CDPAPIKeySecret)
	}
	if strings.TrimSpace(src.CDPWalletSecret) != "" {
		dst.CDPWalletSecret = strings.TrimSpace(src.CDPWalletSecret)
	}
	if strings.TrimSpace(src.FireflySignerURL) != "" {
		dst.FireflySignerURL = strings.TrimSpace(src.FireflySignerURL)
	}
	if strings.TrimSpace(src.FireflyAuthToken) != "" {
		dst.FireflyAuthToken = strings.TrimSpace(src.FireflyAuthToken)
	}
	if strings.TrimSpace(src.FireflyKeystoreDir) != "" {
		dst.FireflyKeystoreDir = strings.TrimSpace(src.FireflyKeystoreDir)
	}
	if strings.TrimSpace(src.FireflyKeystorePassphrase) != "" {
		dst.FireflyKeystorePassphrase = strings.TrimSpace(src.FireflyKeystorePassphrase)
	}
	if strings.TrimSpace(src.DefaultNetwork) != "" {
		dst.DefaultNetwork = strings.TrimSpace(src.DefaultNetwork)
	}
	if strings.TrimSpace(src.DefaultAsset) != "" {
		dst.DefaultAsset = strings.TrimSpace(src.DefaultAsset)
	}
	if strings.TrimSpace(src.CustodyMode) != "" {
		dst.CustodyMode = strings.TrimSpace(src.CustodyMode)
	}
}

func mergeOAuthSettings(dst *models.OAuthProviderSettings, src models.OAuthProviderSettings) {
	if strings.TrimSpace(src.ClientID) != "" {
		dst.ClientID = strings.TrimSpace(src.ClientID)
	}
	if strings.TrimSpace(src.ClientSecret) != "" {
		dst.ClientSecret = strings.TrimSpace(src.ClientSecret)
	}
	if strings.TrimSpace(src.RedirectBase) != "" {
		dst.RedirectBase = strings.TrimSpace(src.RedirectBase)
	}
}

func mergeStripeSettings(dst *models.StripeIntegrationSettings, src models.StripeIntegrationSettings) {
	if strings.TrimSpace(src.SecretKey) != "" {
		dst.SecretKey = strings.TrimSpace(src.SecretKey)
	}
	if strings.TrimSpace(src.PublishableKey) != "" {
		dst.PublishableKey = strings.TrimSpace(src.PublishableKey)
	}
	if strings.TrimSpace(src.WebhookSecret) != "" {
		dst.WebhookSecret = strings.TrimSpace(src.WebhookSecret)
	}
	if strings.TrimSpace(src.OnrampReturnURL) != "" {
		dst.OnrampReturnURL = strings.TrimSpace(src.OnrampReturnURL)
	}
	if strings.TrimSpace(src.OnrampRefreshURL) != "" {
		dst.OnrampRefreshURL = strings.TrimSpace(src.OnrampRefreshURL)
	}
	if src.OnrampMinUSD > 0 {
		dst.OnrampMinUSD = src.OnrampMinUSD
	}
	if src.OnrampDefaultUSD > 0 {
		dst.OnrampDefaultUSD = src.OnrampDefaultUSD
	}
	if strings.TrimSpace(src.ConnectReturnURL) != "" {
		dst.ConnectReturnURL = strings.TrimSpace(src.ConnectReturnURL)
	}
	if strings.TrimSpace(src.ConnectRefreshURL) != "" {
		dst.ConnectRefreshURL = strings.TrimSpace(src.ConnectRefreshURL)
	}
	if strings.TrimSpace(src.ConnectWebhookSecret) != "" {
		dst.ConnectWebhookSecret = strings.TrimSpace(src.ConnectWebhookSecret)
	}
}

func mergeX402Settings(dst *models.X402IntegrationSettings, src models.X402IntegrationSettings) {
	if strings.TrimSpace(src.Mode) != "" {
		dst.Mode = strings.TrimSpace(src.Mode)
	}
	if strings.TrimSpace(src.FacilitatorURL) != "" {
		dst.FacilitatorURL = strings.TrimSpace(src.FacilitatorURL)
	}
	if strings.TrimSpace(src.FacilitatorAPIKey) != "" {
		dst.FacilitatorAPIKey = strings.TrimSpace(src.FacilitatorAPIKey)
	}
}

func mergeN8NSettings(dst *models.N8NIntegrationSettings, src models.N8NIntegrationSettings, cfg config.Config) {
	if strings.TrimSpace(src.BaseURL) != "" {
		dst.BaseURL = strings.TrimSpace(src.BaseURL)
	}
	if strings.TrimSpace(src.APIKey) != "" {
		dst.APIKey = strings.TrimSpace(src.APIKey)
	}
	if src.TimeoutSeconds > 0 {
		dst.TimeoutSeconds = src.TimeoutSeconds
	}
	if dst.TimeoutSeconds <= 0 {
		dst.TimeoutSeconds = cfg.N8NTimeoutSeconds
	}
	if dst.TimeoutSeconds <= 0 {
		dst.TimeoutSeconds = 12
	}
}

func (a *App) currentX402Service() *x402Service {
	resolved := a.resolvedIntegrations()
	cfg := a.cfg
	cfg.X402Mode = resolved.X402.Mode
	cfg.X402FacilitatorURL = resolved.X402.FacilitatorURL
	cfg.X402FacilitatorAPIKey = resolved.X402.FacilitatorAPIKey
	return newX402Service(cfg)
}

func (a *App) currentStripeOnrampService() *stripeOnrampService {
	resolved := a.resolvedIntegrations()
	cfg := a.cfg
	cfg.StripeSecretKey = resolved.Stripe.SecretKey
	cfg.StripeWebhookSecret = resolved.Stripe.WebhookSecret
	cfg.StripeOnrampReturnURL = resolved.Stripe.OnrampReturnURL
	cfg.StripeOnrampRefreshURL = resolved.Stripe.OnrampRefreshURL
	cfg.StripeOnrampMinUSD = resolved.Stripe.OnrampMinUSD
	cfg.StripeOnrampDefaultUSD = resolved.Stripe.OnrampDefaultUSD
	return newStripeOnrampService(cfg)
}

func (a *App) currentStripeConnectService() *stripeConnectService {
	resolved := a.resolvedIntegrations()
	cfg := a.cfg
	cfg.StripeSecretKey = resolved.Stripe.SecretKey
	cfg.StripeConnectReturnURL = resolved.Stripe.ConnectReturnURL
	cfg.StripeConnectRefreshURL = resolved.Stripe.ConnectRefreshURL
	cfg.StripeConnectWebhookSecret = resolved.Stripe.ConnectWebhookSecret
	return newStripeConnectService(cfg)
}

func (a *App) currentN8NService() *n8nService {
	resolved := a.resolvedIntegrations()
	cfg := a.cfg
	cfg.N8NBaseURL = resolved.N8N.BaseURL
	cfg.N8NAPIKey = resolved.N8N.APIKey
	cfg.N8NTimeoutSeconds = resolved.N8N.TimeoutSeconds
	return newN8NService(cfg)
}

func (a *App) runtimePublicConfig() map[string]interface{} {
	resolved := a.resolvedIntegrations()
	return map[string]interface{}{
		"stripe": map[string]interface{}{
			"publishableKey": strings.TrimSpace(resolved.Stripe.PublishableKey),
		},
		"n8n": map[string]interface{}{
			"url": strings.TrimSpace(resolved.N8N.BaseURL),
		},
		"oauth": map[string]interface{}{
			"googleConfigured": strings.TrimSpace(resolved.Google.ClientID) != "" && strings.TrimSpace(resolved.Google.ClientSecret) != "",
			"githubConfigured": strings.TrimSpace(resolved.GitHub.ClientID) != "" && strings.TrimSpace(resolved.GitHub.ClientSecret) != "",
		},
		"wallet": map[string]interface{}{
			"provider":                 strings.TrimSpace(resolved.Wallet.Provider),
			"activeProvider":           a.resolveWalletProviderName(resolved.Wallet),
			"managedAutoPayEnabled":    resolved.Wallet.ManagedAutoPayEnabled,
			"legacyPaymentModeEnabled": resolved.Wallet.LegacyPaymentModeEnabled,
			"externalWalletsEnabled":   resolved.Wallet.ExternalWalletsEnabled,
			"cdpEnabled":               resolved.Wallet.CDPEnabled,
			"fireflyEnabled":           resolved.Wallet.FireflyEnabled,
			"defaultNetwork":           strings.TrimSpace(resolved.Wallet.DefaultNetwork),
			"defaultAsset":             strings.TrimSpace(resolved.Wallet.DefaultAsset),
			"custodyMode":              strings.TrimSpace(resolved.Wallet.CustodyMode),
		},
	}
}

func (a *App) integrationSettingsResponse() map[string]interface{} {
	stored, hasStored := a.storedPlatformIntegrationSettings()
	resolved := a.resolvedIntegrations()
	googleSource := sectionSource(hasStored && sectionHasOAuthValues(stored.Google), strings.TrimSpace(a.cfg.GoogleClientID) != "" || strings.TrimSpace(a.cfg.GoogleClientSecret) != "")
	githubSource := sectionSource(hasStored && sectionHasOAuthValues(stored.GitHub), strings.TrimSpace(a.cfg.GitHubClientID) != "" || strings.TrimSpace(a.cfg.GitHubClientSecret) != "")
	stripeSource := sectionSource(hasStored && sectionHasStripeValues(stored.Stripe), strings.TrimSpace(a.cfg.StripeSecretKey) != "" || strings.TrimSpace(a.cfg.StripeOnrampReturnURL) != "" || strings.TrimSpace(a.cfg.StripeConnectReturnURL) != "")
	x402Source := sectionSource(hasStored && sectionHasX402Values(stored.X402), strings.TrimSpace(a.cfg.X402Mode) != "" || strings.TrimSpace(a.cfg.X402FacilitatorURL) != "")
	walletSource := sectionSource(hasStored && sectionHasWalletValues(stored.Wallet), strings.TrimSpace(a.cfg.WalletProvider) != "" || strings.TrimSpace(a.cfg.CDPAPIKeyID) != "" || strings.TrimSpace(a.cfg.FireflySignerURL) != "")
	n8nSource := sectionSource(hasStored && sectionHasN8NValues(stored.N8N), strings.TrimSpace(a.cfg.N8NBaseURL) != "" || strings.TrimSpace(a.cfg.N8NAPIKey) != "")

	return map[string]interface{}{
		"settings": map[string]interface{}{
			"google": map[string]interface{}{
				"clientId":     resolved.Google.ClientID,
				"redirectBase": resolved.Google.RedirectBase,
				"clientSecret": secretField(resolved.Google.ClientSecret),
			},
			"github": map[string]interface{}{
				"clientId":     resolved.GitHub.ClientID,
				"redirectBase": resolved.GitHub.RedirectBase,
				"clientSecret": secretField(resolved.GitHub.ClientSecret),
			},
			"stripe": map[string]interface{}{
				"publishableKey":       resolved.Stripe.PublishableKey,
				"secretKey":            secretField(resolved.Stripe.SecretKey),
				"webhookSecret":        secretField(resolved.Stripe.WebhookSecret),
				"onrampReturnUrl":      resolved.Stripe.OnrampReturnURL,
				"onrampRefreshUrl":     resolved.Stripe.OnrampRefreshURL,
				"onrampMinUsd":         resolved.Stripe.OnrampMinUSD,
				"onrampDefaultUsd":     resolved.Stripe.OnrampDefaultUSD,
				"connectReturnUrl":     resolved.Stripe.ConnectReturnURL,
				"connectRefreshUrl":    resolved.Stripe.ConnectRefreshURL,
				"connectWebhookSecret": secretField(resolved.Stripe.ConnectWebhookSecret),
			},
			"wallet": map[string]interface{}{
				"provider":                  resolved.Wallet.Provider,
				"activeProvider":            a.resolveWalletProviderName(resolved.Wallet),
				"managedAutoPayEnabled":     resolved.Wallet.ManagedAutoPayEnabled,
				"legacyPaymentModeEnabled":  resolved.Wallet.LegacyPaymentModeEnabled,
				"externalWalletsEnabled":    resolved.Wallet.ExternalWalletsEnabled,
				"cdpEnabled":                resolved.Wallet.CDPEnabled,
				"fireflyEnabled":            resolved.Wallet.FireflyEnabled,
				"cdpApiKeyId":               resolved.Wallet.CDPAPIKeyID,
				"cdpApiKeySecret":           secretField(resolved.Wallet.CDPAPIKeySecret),
				"cdpWalletSecret":           secretField(resolved.Wallet.CDPWalletSecret),
				"fireflySignerUrl":          resolved.Wallet.FireflySignerURL,
				"fireflyAuthToken":          secretField(resolved.Wallet.FireflyAuthToken),
				"fireflyKeystoreDir":        resolved.Wallet.FireflyKeystoreDir,
				"fireflyKeystorePassphrase": secretField(resolved.Wallet.FireflyKeystorePassphrase),
				"defaultNetwork":            resolved.Wallet.DefaultNetwork,
				"defaultAsset":              resolved.Wallet.DefaultAsset,
				"custodyMode":               resolved.Wallet.CustodyMode,
			},
			"x402": map[string]interface{}{
				"mode":              resolved.X402.Mode,
				"facilitatorUrl":    resolved.X402.FacilitatorURL,
				"facilitatorApiKey": secretField(resolved.X402.FacilitatorAPIKey),
			},
			"n8n": map[string]interface{}{
				"baseUrl":        resolved.N8N.BaseURL,
				"apiKey":         secretField(resolved.N8N.APIKey),
				"timeoutSeconds": resolved.N8N.TimeoutSeconds,
			},
		},
		"status": map[string]interface{}{
			"google": integrationStatus{
				Configured: strings.TrimSpace(resolved.Google.ClientID) != "" && strings.TrimSpace(resolved.Google.ClientSecret) != "",
				Source:     googleSource,
			},
			"github": integrationStatus{
				Configured: strings.TrimSpace(resolved.GitHub.ClientID) != "" && strings.TrimSpace(resolved.GitHub.ClientSecret) != "",
				Source:     githubSource,
			},
			"stripe": integrationStatus{
				Configured: strings.TrimSpace(resolved.Stripe.SecretKey) != "",
				Source:     stripeSource,
			},
			"wallet": integrationStatus{
				Configured: strings.TrimSpace(a.resolveWalletProviderName(resolved.Wallet)) != "",
				Source:     walletSource,
				Notes:      a.walletStatusNote(resolved.Wallet),
			},
			"x402": integrationStatus{
				Configured: strings.TrimSpace(resolved.X402.Mode) != "" && (strings.EqualFold(strings.TrimSpace(resolved.X402.Mode), "disabled") || strings.TrimSpace(resolved.X402.FacilitatorURL) != ""),
				Source:     x402Source,
			},
			"n8n": integrationStatus{
				Configured: strings.TrimSpace(resolved.N8N.BaseURL) != "",
				Source:     n8nSource,
			},
		},
		"runtime": a.runtimePublicConfig(),
	}
}

func secretField(value string) integrationSecretField {
	trimmed := strings.TrimSpace(value)
	return integrationSecretField{
		Set:    trimmed != "",
		Masked: maskSecret(trimmed),
	}
}

func maskSecret(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	if len(value) <= 6 {
		return strings.Repeat("*", len(value))
	}
	return value[:3] + strings.Repeat("*", len(value)-6) + value[len(value)-3:]
}

func sectionSource(hasUI bool, hasEnv bool) string {
	switch {
	case hasUI:
		return "ui"
	case hasEnv:
		return "env"
	default:
		return "none"
	}
}

func sectionHasOAuthValues(settings models.OAuthProviderSettings) bool {
	return strings.TrimSpace(settings.ClientID) != "" || strings.TrimSpace(settings.ClientSecret) != "" || strings.TrimSpace(settings.RedirectBase) != ""
}

func sectionHasStripeValues(settings models.StripeIntegrationSettings) bool {
	return strings.TrimSpace(settings.SecretKey) != "" ||
		strings.TrimSpace(settings.PublishableKey) != "" ||
		strings.TrimSpace(settings.WebhookSecret) != "" ||
		strings.TrimSpace(settings.OnrampReturnURL) != "" ||
		strings.TrimSpace(settings.ConnectReturnURL) != ""
}

func sectionHasX402Values(settings models.X402IntegrationSettings) bool {
	return strings.TrimSpace(settings.Mode) != "" || strings.TrimSpace(settings.FacilitatorURL) != "" || strings.TrimSpace(settings.FacilitatorAPIKey) != ""
}

func sectionHasWalletValues(settings models.WalletIntegrationSettings) bool {
	return strings.TrimSpace(settings.Provider) != "" ||
		settings.ManagedAutoPayEnabled ||
		settings.LegacyPaymentModeEnabled ||
		settings.ExternalWalletsEnabled ||
		settings.CDPEnabled ||
		settings.FireflyEnabled ||
		strings.TrimSpace(settings.CDPAPIKeyID) != "" ||
		strings.TrimSpace(settings.CDPAPIKeySecret) != "" ||
		strings.TrimSpace(settings.CDPWalletSecret) != "" ||
		strings.TrimSpace(settings.FireflySignerURL) != "" ||
		strings.TrimSpace(settings.FireflyAuthToken) != "" ||
		strings.TrimSpace(settings.FireflyKeystoreDir) != "" ||
		strings.TrimSpace(settings.FireflyKeystorePassphrase) != "" ||
		strings.TrimSpace(settings.DefaultNetwork) != "" ||
		strings.TrimSpace(settings.DefaultAsset) != "" ||
		strings.TrimSpace(settings.CustodyMode) != ""
}

func (a *App) walletStatusNote(settings models.WalletIntegrationSettings) string {
	active := a.resolveWalletProviderName(settings)
	if active == "" {
		return "No managed wallet backend selected."
	}
	if !settings.ManagedAutoPayEnabled {
		return "Managed wallet auto-pay is disabled."
	}
	switch active {
	case "firefly":
		if strings.TrimSpace(settings.FireflySignerURL) == "" {
			return "FireFly selected without signer URL."
		}
	case "cdp":
		if strings.TrimSpace(settings.CDPAPIKeyID) == "" {
			return "CDP selected without API credentials."
		}
	}
	return "Active provider: " + active
}

func (a *App) resolveWalletProviderName(settings models.WalletIntegrationSettings) string {
	preferred := strings.ToLower(strings.TrimSpace(settings.Provider))
	switch preferred {
	case "firefly":
		if settings.FireflyEnabled {
			return "firefly"
		}
	case "cdp":
		if settings.CDPEnabled {
			return "cdp"
		}
	}
	if settings.CDPEnabled {
		return "cdp"
	}
	if settings.FireflyEnabled {
		return "firefly"
	}
	return preferred
}

func sectionHasN8NValues(settings models.N8NIntegrationSettings) bool {
	return strings.TrimSpace(settings.BaseURL) != "" || strings.TrimSpace(settings.APIKey) != "" || settings.TimeoutSeconds > 0
}

func normalizeWalletSettings(settings models.WalletIntegrationSettings) models.WalletIntegrationSettings {
	if !settings.ManagedAutoPayEnabled && !settings.LegacyPaymentModeEnabled && !settings.CDPEnabled && !settings.FireflyEnabled && !settings.ExternalWalletsEnabled {
		settings.ManagedAutoPayEnabled = true
		settings.LegacyPaymentModeEnabled = true
		switch strings.ToLower(strings.TrimSpace(settings.Provider)) {
		case "firefly":
			settings.FireflyEnabled = true
		default:
			settings.CDPEnabled = true
		}
	}
	if strings.EqualFold(strings.TrimSpace(settings.Provider), "firefly") && !settings.CDPEnabled && !settings.FireflyEnabled {
		settings.FireflyEnabled = true
	}
	if strings.EqualFold(strings.TrimSpace(settings.Provider), "cdp") && !settings.CDPEnabled && !settings.FireflyEnabled {
		settings.CDPEnabled = true
	}
	return settings
}
