package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                           string
	JWTPrivateKeyPEM               string
	JWTPublicKeyPEM                string
	JWTKeyID                       string
	BaseURL                        string
	N8NBaseURL                     string
	N8NAPIKey                      string
	N8NTimeoutSeconds              int
	X402FacilitatorURL             string
	X402FacilitatorAPIKey          string
	X402Mode                       string
	SupportedPayMethods            []string
	StripeSecretKey                string
	StripeWebhookSecret            string
	StripeOnrampReturnURL          string
	StripeOnrampRefreshURL         string
	StripeOnrampMinUSD             float64
	StripeOnrampDefaultUSD         float64
	StripeConnectReturnURL         string
	StripeConnectRefreshURL        string
	StripeConnectWebhookSecret     string
	WalletProvider                 string
	WalletManagedAutoPayEnabled    bool
	WalletLegacyPaymentModeEnabled bool
	WalletExternalWalletsEnabled   bool
	WalletCDPEnabled               bool
	WalletFireflyEnabled           bool
	CDPAPIKeyID                    string
	CDPAPIKeySecret                string
	CDPWalletSecret                string
	FireflySignerURL               string
	FireflyAuthToken               string
	FireflyKeystoreDir             string
	FireflyKeystorePassphrase      string
	WalletDefaultNetwork           string
	WalletDefaultAsset             string
	WalletCustodyMode              string
	PlatformFeeBps                 int
	PlatformMinFeeUSDC             float64
	PlatformMaxFeeUSDC             float64
	PlatformHoldDays               int
	MongoURI                       string
	MongoDBName                    string
	MongoRequired                  bool
	SuperAdminEmail                string
	SuperAdminPassword             string
	DataFilePath                   string
	CORSAllowedOrigins             []string
	RateLimitPerMinute             int
	TrustProxyHeaders              bool
	AllowInsecureDefaults          bool
	GoogleClientID                 string
	GoogleClientSecret             string
	GitHubClientID                 string
	GitHubClientSecret             string
	OAuthRedirectBase              string
	MCPSDKEnabled                  bool
	DockerRuntimeSocket            string
	DockerRuntimeNetwork           string
	DockerRuntimePublicHost        string
}

func Load() Config {
	allowInsecureDefaults := parseBool(os.Getenv("ALLOW_INSECURE_DEFAULTS"), false)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	jwtPrivateKeyPEM := strings.TrimSpace(os.Getenv("JWT_PRIVATE_KEY_PEM"))
	jwtPublicKeyPEM := strings.TrimSpace(os.Getenv("JWT_PUBLIC_KEY_PEM"))
	jwtKeyID := strings.TrimSpace(os.Getenv("JWT_KEY_ID"))
	if jwtKeyID == "" {
		jwtKeyID = "default-rs256-key"
	}
	baseURL := strings.TrimSpace(os.Getenv("BASE_URL"))
	n8nBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("N8N_BASE_URL")), "/")
	n8nAPIKey := strings.TrimSpace(os.Getenv("N8N_API_KEY"))
	n8nTimeoutSeconds := parsePositiveInt(os.Getenv("N8N_TIMEOUT_SECONDS"), 12)
	x402Mode := strings.ToLower(trimASCII(os.Getenv("X402_MODE")))
	// Default to facilitator mode for real payment validation
	// Set X402_MODE="disabled" to bypass payment verification for development/testing
	if x402Mode == "" {
		x402Mode = "facilitator"
	}
	facilitatorURL := strings.TrimSpace(os.Getenv("X402_FACILITATOR_URL"))
	facilitatorAPIKey := strings.TrimSpace(os.Getenv("X402_FACILITATOR_API_KEY"))
	payMethodsCSV := os.Getenv("SUPPORTED_PAYMENT_METHODS")
	if payMethodsCSV == "" {
		payMethodsCSV = "x402_wallet,wallet_balance"
	}
	supportedPayMethods := splitAndTrim(payMethodsCSV)
	stripeSecretKey := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	stripeWebhookSecret := strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
	stripeOnrampReturnURL := strings.TrimSpace(os.Getenv("STRIPE_ONRAMP_RETURN_URL"))
	stripeOnrampRefreshURL := strings.TrimSpace(os.Getenv("STRIPE_ONRAMP_REFRESH_URL"))
	stripeOnrampMinUSD := parsePositiveFloat(os.Getenv("STRIPE_ONRAMP_MIN_USD"), 10.0)
	stripeOnrampDefaultUSD := parsePositiveFloat(os.Getenv("STRIPE_ONRAMP_DEFAULT_USD"), 50.0)
	stripeConnectReturnURL := strings.TrimSpace(os.Getenv("STRIPE_CONNECT_RETURN_URL"))
	stripeConnectRefreshURL := strings.TrimSpace(os.Getenv("STRIPE_CONNECT_REFRESH_URL"))
	stripeConnectWebhookSecret := strings.TrimSpace(os.Getenv("STRIPE_CONNECT_WEBHOOK_SECRET"))
	walletProvider := strings.ToLower(strings.TrimSpace(os.Getenv("WALLET_PROVIDER")))
	if walletProvider == "" {
		walletProvider = "cdp"
	}
	walletManagedAutoPayEnabled := parseBool(os.Getenv("WALLET_MANAGED_AUTOPAY_ENABLED"), true)
	walletLegacyPaymentModeEnabled := parseBool(os.Getenv("WALLET_LEGACY_PAYMENT_MODE_ENABLED"), true)
	walletExternalWalletsEnabled := parseBool(os.Getenv("WALLET_EXTERNAL_WALLETS_ENABLED"), false)
	walletCDPEnabled := parseBool(os.Getenv("WALLET_CDP_ENABLED"), true)
	walletFireflyEnabled := parseBool(os.Getenv("WALLET_FIREFLY_ENABLED"), false)
	cdpAPIKeyID := strings.TrimSpace(os.Getenv("CDP_API_KEY_ID"))
	cdpAPIKeySecret := strings.TrimSpace(os.Getenv("CDP_API_KEY_SECRET"))
	cdpWalletSecret := strings.TrimSpace(os.Getenv("CDP_WALLET_SECRET"))
	fireflySignerURL := strings.TrimRight(strings.TrimSpace(os.Getenv("FIREFLY_SIGNER_URL")), "/")
	fireflyAuthToken := strings.TrimSpace(os.Getenv("FIREFLY_SIGNER_AUTH_TOKEN"))
	fireflyKeystoreDir := strings.TrimSpace(os.Getenv("FIREFLY_KEYSTORE_DIR"))
	if fireflyKeystoreDir == "" {
		fireflyKeystoreDir = "./data/firefly/keystore"
	}
	fireflyKeystorePassphrase := strings.TrimSpace(os.Getenv("FIREFLY_KEYSTORE_PASSPHRASE"))
	walletDefaultNetwork := strings.TrimSpace(os.Getenv("WALLET_DEFAULT_NETWORK"))
	if walletDefaultNetwork == "" {
		walletDefaultNetwork = "base"
	}
	walletDefaultAsset := strings.TrimSpace(os.Getenv("WALLET_DEFAULT_ASSET"))
	if walletDefaultAsset == "" {
		walletDefaultAsset = "USDC"
	}
	walletCustodyMode := strings.TrimSpace(os.Getenv("WALLET_CUSTODY_MODE"))
	if walletCustodyMode == "" {
		walletCustodyMode = "provider_managed"
	}
	platformFeeBps := parsePositiveInt(os.Getenv("PLATFORM_FEE_BPS"), 1000)
	platformMinFeeUSDC := parseNonNegativeFloat(os.Getenv("PLATFORM_MIN_FEE_USDC"), 0)
	platformMaxFeeUSDC := parseNonNegativeFloat(os.Getenv("PLATFORM_MAX_FEE_USDC"), 0)
	platformHoldDays := parseNonNegativeInt(os.Getenv("PLATFORM_HOLD_DAYS"), 0)
	mongoURI := strings.TrimSpace(os.Getenv("MONGO_URI"))
	mongoDBName := strings.TrimSpace(os.Getenv("MONGO_DB_NAME"))
	if mongoDBName == "" {
		mongoDBName = "mcp_marketplace"
	}
	mongoRequired := parseBool(os.Getenv("MONGO_REQUIRED"), !allowInsecureDefaults)
	superAdminEmail := os.Getenv("SUPER_ADMIN_EMAIL")
	if superAdminEmail == "" {
		superAdminEmail = "admin@platform.local"
	}
	superAdminPassword := os.Getenv("SUPER_ADMIN_PASSWORD")
	if superAdminPassword == "" && allowInsecureDefaults {
		superAdminPassword = "change-admin-password"
	}
	dataFilePath := os.Getenv("DATA_FILE_PATH")
	if dataFilePath == "" {
		dataFilePath = "./data/store.json"
	}

	originsCSV := os.Getenv("CORS_ALLOWED_ORIGINS")
	corsOrigins := []string{}
	for _, part := range splitAndTrim(originsCSV) {
		if part != "" {
			corsOrigins = append(corsOrigins, part)
		}
	}

	rateLimit := parsePositiveInt(os.Getenv("RATE_LIMIT_PER_MINUTE"), 240)
	trustProxyHeaders := parseBool(os.Getenv("TRUST_PROXY_HEADERS"), false)

	googleClientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	googleClientSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	githubClientID := strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID"))
	githubClientSecret := strings.TrimSpace(os.Getenv("GITHUB_CLIENT_SECRET"))
	oauthRedirectBase := strings.TrimSpace(os.Getenv("OAUTH_REDIRECT_BASE"))
	mcpSDKEnabled := parseBool(os.Getenv("MCP_SDK_ENABLED"), false)
	dockerRuntimeSocket := strings.TrimSpace(os.Getenv("DOCKER_RUNTIME_SOCKET"))
	if dockerRuntimeSocket == "" {
		dockerRuntimeSocket = "/var/run/docker.sock"
	}
	dockerRuntimeNetwork := strings.TrimSpace(os.Getenv("DOCKER_RUNTIME_NETWORK"))
	if dockerRuntimeNetwork == "" {
		dockerRuntimeNetwork = "infra_default"
	}
	dockerRuntimePublicHost := strings.TrimSpace(os.Getenv("DOCKER_RUNTIME_PUBLIC_HOST"))
	if dockerRuntimePublicHost == "" {
		dockerRuntimePublicHost = "localhost"
	}

	return Config{
		Port:                           port,
		JWTPrivateKeyPEM:               jwtPrivateKeyPEM,
		JWTPublicKeyPEM:                jwtPublicKeyPEM,
		JWTKeyID:                       jwtKeyID,
		BaseURL:                        baseURL,
		N8NBaseURL:                     n8nBaseURL,
		N8NAPIKey:                      n8nAPIKey,
		N8NTimeoutSeconds:              n8nTimeoutSeconds,
		X402FacilitatorURL:             facilitatorURL,
		X402FacilitatorAPIKey:          facilitatorAPIKey,
		X402Mode:                       x402Mode,
		SupportedPayMethods:            supportedPayMethods,
		StripeSecretKey:                stripeSecretKey,
		StripeWebhookSecret:            stripeWebhookSecret,
		StripeOnrampReturnURL:          stripeOnrampReturnURL,
		StripeOnrampRefreshURL:         stripeOnrampRefreshURL,
		StripeOnrampMinUSD:             stripeOnrampMinUSD,
		StripeOnrampDefaultUSD:         stripeOnrampDefaultUSD,
		StripeConnectReturnURL:         stripeConnectReturnURL,
		StripeConnectRefreshURL:        stripeConnectRefreshURL,
		StripeConnectWebhookSecret:     stripeConnectWebhookSecret,
		WalletProvider:                 walletProvider,
		WalletManagedAutoPayEnabled:    walletManagedAutoPayEnabled,
		WalletLegacyPaymentModeEnabled: walletLegacyPaymentModeEnabled,
		WalletExternalWalletsEnabled:   walletExternalWalletsEnabled,
		WalletCDPEnabled:               walletCDPEnabled,
		WalletFireflyEnabled:           walletFireflyEnabled,
		CDPAPIKeyID:                    cdpAPIKeyID,
		CDPAPIKeySecret:                cdpAPIKeySecret,
		CDPWalletSecret:                cdpWalletSecret,
		FireflySignerURL:               fireflySignerURL,
		FireflyAuthToken:               fireflyAuthToken,
		FireflyKeystoreDir:             fireflyKeystoreDir,
		FireflyKeystorePassphrase:      fireflyKeystorePassphrase,
		WalletDefaultNetwork:           walletDefaultNetwork,
		WalletDefaultAsset:             walletDefaultAsset,
		WalletCustodyMode:              walletCustodyMode,
		PlatformFeeBps:                 platformFeeBps,
		PlatformMinFeeUSDC:             platformMinFeeUSDC,
		PlatformMaxFeeUSDC:             platformMaxFeeUSDC,
		PlatformHoldDays:               platformHoldDays,
		MongoURI:                       mongoURI,
		MongoDBName:                    mongoDBName,
		MongoRequired:                  mongoRequired,
		SuperAdminEmail:                superAdminEmail,
		SuperAdminPassword:             superAdminPassword,
		DataFilePath:                   dataFilePath,
		CORSAllowedOrigins:             corsOrigins,
		RateLimitPerMinute:             rateLimit,
		TrustProxyHeaders:              trustProxyHeaders,
		AllowInsecureDefaults:          allowInsecureDefaults,
		GoogleClientID:                 googleClientID,
		GoogleClientSecret:             googleClientSecret,
		GitHubClientID:                 githubClientID,
		GitHubClientSecret:             githubClientSecret,
		OAuthRedirectBase:              oauthRedirectBase,
		MCPSDKEnabled:                  mcpSDKEnabled,
		DockerRuntimeSocket:            dockerRuntimeSocket,
		DockerRuntimeNetwork:           dockerRuntimeNetwork,
		DockerRuntimePublicHost:        dockerRuntimePublicHost,
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Port) == "" {
		return fmt.Errorf("PORT must be set")
	}
	if strings.TrimSpace(c.BaseURL) == "" {
		return fmt.Errorf("BASE_URL must be set")
	}
	if len(c.CORSAllowedOrigins) == 0 {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS must be set")
	}
	if strings.TrimSpace(c.GoogleClientID) != "" || strings.TrimSpace(c.GoogleClientSecret) != "" ||
		strings.TrimSpace(c.GitHubClientID) != "" || strings.TrimSpace(c.GitHubClientSecret) != "" {
		if strings.TrimSpace(c.OAuthRedirectBase) == "" {
			return fmt.Errorf("OAUTH_REDIRECT_BASE must be set when Google or GitHub OAuth is configured")
		}
	}
	if strings.TrimSpace(c.SuperAdminEmail) == "" {
		return fmt.Errorf("SUPER_ADMIN_EMAIL must be set")
	}
	switch strings.ToLower(strings.TrimSpace(c.X402Mode)) {
	case "", "facilitator", "disabled", "test":
	default:
		return fmt.Errorf("X402_MODE must be one of: facilitator, disabled, test")
	}
	if c.PlatformFeeBps < 0 || c.PlatformFeeBps > 10000 {
		return fmt.Errorf("PLATFORM_FEE_BPS must be between 0 and 10000")
	}
	if strings.TrimSpace(c.MongoURI) == "" {
		return fmt.Errorf("MONGO_URI must be set (in-memory fallback is disabled)")
	}
	hasJWTKeys := strings.TrimSpace(c.JWTPrivateKeyPEM) != "" && strings.TrimSpace(c.JWTPublicKeyPEM) != ""
	if c.AllowInsecureDefaults {
		return nil
	}
	if !hasJWTKeys {
		return fmt.Errorf("JWT_PRIVATE_KEY_PEM and JWT_PUBLIC_KEY_PEM must be set")
	}
	if strings.TrimSpace(c.SuperAdminPassword) == "" || c.SuperAdminPassword == "change-admin-password" {
		return fmt.Errorf("SUPER_ADMIN_PASSWORD must be set to a secure value")
	}
	if strings.TrimSpace(c.StripeSecretKey) != "" {
		if strings.TrimSpace(c.StripeWebhookSecret) == "" {
			return fmt.Errorf("STRIPE_WEBHOOK_SECRET must be set when STRIPE_SECRET_KEY is set")
		}
		if strings.TrimSpace(c.StripeConnectWebhookSecret) == "" {
			return fmt.Errorf("STRIPE_CONNECT_WEBHOOK_SECRET must be set when STRIPE_SECRET_KEY is set")
		}
	}
	if strings.TrimSpace(c.WalletProvider) != "" && !c.AllowInsecureDefaults && strings.EqualFold(strings.TrimSpace(c.WalletProvider), "cdp") {
		if strings.TrimSpace(c.CDPAPIKeyID) == "" || strings.TrimSpace(c.CDPAPIKeySecret) == "" || strings.TrimSpace(c.CDPWalletSecret) == "" {
			return fmt.Errorf("CDP_API_KEY_ID, CDP_API_KEY_SECRET, and CDP_WALLET_SECRET must be set when WALLET_PROVIDER=cdp")
		}
	}
	if strings.TrimSpace(c.WalletProvider) != "" && !c.AllowInsecureDefaults && strings.EqualFold(strings.TrimSpace(c.WalletProvider), "firefly") {
		if strings.TrimSpace(c.FireflySignerURL) == "" || strings.TrimSpace(c.FireflyKeystoreDir) == "" || strings.TrimSpace(c.FireflyKeystorePassphrase) == "" {
			return fmt.Errorf("FIREFLY_SIGNER_URL, FIREFLY_KEYSTORE_DIR, and FIREFLY_KEYSTORE_PASSPHRASE must be set when WALLET_PROVIDER=firefly")
		}
	}
	return nil
}

func splitAndTrim(csv string) []string {
	out := []string{}
	current := ""
	for _, r := range csv {
		if r == ',' {
			out = append(out, trimASCII(current))
			current = ""
			continue
		}
		current += string(r)
	}
	out = append(out, trimASCII(current))
	return out
}

func trimASCII(v string) string {
	start := 0
	end := len(v)
	for start < end && (v[start] == ' ' || v[start] == '\t' || v[start] == '\n' || v[start] == '\r') {
		start++
	}
	for end > start && (v[end-1] == ' ' || v[end-1] == '\t' || v[end-1] == '\n' || v[end-1] == '\r') {
		end--
	}
	return v[start:end]
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n := 0
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c < '0' || c > '9' {
			return fallback
		}
		n = n*10 + int(c-'0')
	}
	if n <= 0 {
		return fallback
	}
	return n
}

func parseBool(raw string, fallback bool) bool {
	if raw == "" {
		return fallback
	}
	switch strings.ToLower(trimASCII(raw)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func parsePositiveFloat(raw string, fallback float64) float64 {
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(trimASCII(raw), 64)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func parseNonNegativeFloat(raw string, fallback float64) float64 {
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(trimASCII(raw), 64)
	if err != nil || v < 0 {
		return fallback
	}
	return v
}

func parseNonNegativeInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(trimASCII(raw))
	if err != nil || n < 0 {
		return fallback
	}
	return n
}
