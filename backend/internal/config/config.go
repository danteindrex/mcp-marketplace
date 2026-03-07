package config

import "os"

type Config struct {
	Port               string
	JWTSecret          string
	BaseURL            string
	SuperAdminEmail    string
	SuperAdminPassword string
	DataFilePath       string
	CORSAllowedOrigins []string
	RateLimitPerMinute int
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change-me-in-production"
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}
	superAdminEmail := os.Getenv("SUPER_ADMIN_EMAIL")
	if superAdminEmail == "" {
		superAdminEmail = "admin@platform.local"
	}
	superAdminPassword := os.Getenv("SUPER_ADMIN_PASSWORD")
	if superAdminPassword == "" {
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
	return Config{
		Port:               port,
		JWTSecret:          secret,
		BaseURL:            baseURL,
		SuperAdminEmail:    superAdminEmail,
		SuperAdminPassword: superAdminPassword,
		DataFilePath:       dataFilePath,
		CORSAllowedOrigins: corsOrigins,
		RateLimitPerMinute: rateLimit,
	}
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
