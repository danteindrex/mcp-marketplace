package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                       string
	JWTSecret                  string
	BaseURL                    string
	N8NBaseURL                 string
	N8NAPIKey                  string
	N8NTimeoutSeconds          int
	X402FacilitatorURL         string
	X402FacilitatorAPIKey      string
	X402Mode                   string
	SupportedPayMethods        []string
	StripeSecretKey            string
	StripeWebhookSecret        string
	StripeOnrampReturnURL      string
	StripeOnrampRefreshURL     string
	StripeOnrampMinUSD         float64
	StripeOnrampDefaultUSD     float64
	StripeConnectReturnURL     string
	StripeConnectRefreshURL    string
	StripeConnectWebhookSecret string
	PlatformFeeBps             int
	PlatformMinFeeUSDC         float64
	PlatformMaxFeeUSDC         float64
	PlatformHoldDays           int
	MongoURI                   string
	MongoDBName                string
	MongoRequired              bool
	SuperAdminEmail            string
	SuperAdminPassword         string
	DataFilePath               string
	CORSAllowedOrigins         []string
	RateLimitPerMinute         int
	TrustProxyHeaders          bool
	AllowInsecureDefaults      bool
}

func Load() Config {
	allowInsecureDefaults := parseBool(os.Getenv("ALLOW_INSECURE_DEFAULTS"), false)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" && allowInsecureDefaults {
		secret = "change-me-in-production"
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}
	n8nBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("N8N_BASE_URL")), "/")
	n8nAPIKey := strings.TrimSpace(os.Getenv("N8N_API_KEY"))
	n8nTimeoutSeconds := parsePositiveInt(os.Getenv("N8N_TIMEOUT_SECONDS"), 12)
	x402Mode := strings.ToLower(trimASCII(os.Getenv("X402_MODE")))
	if x402Mode == "" {
		x402Mode = "mock"
	}
	facilitatorURL := strings.TrimSpace(os.Getenv("X402_FACILITATOR_URL"))
	facilitatorAPIKey := strings.TrimSpace(os.Getenv("X402_FACILITATOR_API_KEY"))
	payMethodsCSV := os.Getenv("SUPPORTED_PAYMENT_METHODS")
	if payMethodsCSV == "" {
		payMethodsCSV = "x402_wallet,wallet_balance,coinbase_commerce,stripe"
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
	if originsCSV == "" {
		originsCSV = "http://localhost:3000,http://127.0.0.1:3000"
	}
	corsOrigins := []string{}
	for _, part := range splitAndTrim(originsCSV) {
		if part != "" {
			corsOrigins = append(corsOrigins, part)
		}
	}
	if len(corsOrigins) == 0 {
		corsOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}

	rateLimit := parsePositiveInt(os.Getenv("RATE_LIMIT_PER_MINUTE"), 240)
	trustProxyHeaders := parseBool(os.Getenv("TRUST_PROXY_HEADERS"), false)
	return Config{
		Port:                       port,
		JWTSecret:                  secret,
		BaseURL:                    baseURL,
		N8NBaseURL:                 n8nBaseURL,
		N8NAPIKey:                  n8nAPIKey,
		N8NTimeoutSeconds:          n8nTimeoutSeconds,
		X402FacilitatorURL:         facilitatorURL,
		X402FacilitatorAPIKey:      facilitatorAPIKey,
		X402Mode:                   x402Mode,
		SupportedPayMethods:        supportedPayMethods,
		StripeSecretKey:            stripeSecretKey,
		StripeWebhookSecret:        stripeWebhookSecret,
		StripeOnrampReturnURL:      stripeOnrampReturnURL,
		StripeOnrampRefreshURL:     stripeOnrampRefreshURL,
		StripeOnrampMinUSD:         stripeOnrampMinUSD,
		StripeOnrampDefaultUSD:     stripeOnrampDefaultUSD,
		StripeConnectReturnURL:     stripeConnectReturnURL,
		StripeConnectRefreshURL:    stripeConnectRefreshURL,
		StripeConnectWebhookSecret: stripeConnectWebhookSecret,
		PlatformFeeBps:             platformFeeBps,
		PlatformMinFeeUSDC:         platformMinFeeUSDC,
		PlatformMaxFeeUSDC:         platformMaxFeeUSDC,
		PlatformHoldDays:           platformHoldDays,
		MongoURI:                   mongoURI,
		MongoDBName:                mongoDBName,
		MongoRequired:              mongoRequired,
		SuperAdminEmail:            superAdminEmail,
		SuperAdminPassword:         superAdminPassword,
		DataFilePath:               dataFilePath,
		CORSAllowedOrigins:         corsOrigins,
		RateLimitPerMinute:         rateLimit,
		TrustProxyHeaders:          trustProxyHeaders,
		AllowInsecureDefaults:      allowInsecureDefaults,
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Port) == "" {
		return fmt.Errorf("PORT must be set")
	}
	if strings.TrimSpace(c.SuperAdminEmail) == "" {
		return fmt.Errorf("SUPER_ADMIN_EMAIL must be set")
	}
	switch strings.ToLower(strings.TrimSpace(c.X402Mode)) {
	case "", "mock", "facilitator":
	default:
		return fmt.Errorf("X402_MODE must be one of: mock, facilitator")
	}
	if c.PlatformFeeBps < 0 || c.PlatformFeeBps > 10000 {
		return fmt.Errorf("PLATFORM_FEE_BPS must be between 0 and 10000")
	}
	if strings.TrimSpace(c.MongoURI) == "" {
		return fmt.Errorf("MONGO_URI must be set (in-memory fallback is disabled)")
	}
	if c.AllowInsecureDefaults {
		return nil
	}
	if strings.TrimSpace(c.JWTSecret) == "" || c.JWTSecret == "change-me-in-production" {
		return fmt.Errorf("JWT_SECRET must be set to a secure value")
	}
	if strings.TrimSpace(c.SuperAdminPassword) == "" || c.SuperAdminPassword == "change-admin-password" {
		return fmt.Errorf("SUPER_ADMIN_PASSWORD must be set to a secure value")
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
