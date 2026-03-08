package config

import "testing"

func TestValidateRequiresStripeWebhookSecretsInStrictMode(t *testing.T) {
	cfg := Config{
		Port:                       "8080",
		JWTPrivateKeyPEM:           "private-key",
		JWTPublicKeyPEM:            "public-key",
		SuperAdminEmail:            "admin@platform.local",
		SuperAdminPassword:         "admin-pass",
		MongoURI:                   "mongodb://localhost:27017",
		StripeSecretKey:            "sk_test_123",
		StripeWebhookSecret:        "whsec_onramp",
		AllowInsecureDefaults:      false,
		StripeConnectWebhookSecret: "",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing Stripe Connect webhook secret to fail validation")
	}

	cfg.StripeConnectWebhookSecret = "whsec_connect"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected complete strict Stripe config to validate, got %v", err)
	}
}
