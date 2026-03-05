package config

import "os"

type Config struct {
	Port      string
	JWTSecret string
	BaseURL   string
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
	return Config{Port: port, JWTSecret: secret, BaseURL: baseURL}
}
