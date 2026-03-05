package config

import "os"

type Config struct {
	Port               string
	JWTSecret          string
	BaseURL            string
	SuperAdminEmail    string
	SuperAdminPassword string
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
	return Config{
		Port:               port,
		JWTSecret:          secret,
		BaseURL:            baseURL,
		SuperAdminEmail:    superAdminEmail,
		SuperAdminPassword: superAdminPassword,
	}
}
