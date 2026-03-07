package main

import (
	"log"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/store"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration for seed: %v", err)
	}

	mongoStore, err := store.NewMongoStore(cfg)
	if err != nil {
		log.Printf("warning: mongodb connection failed during seed: %v", err)
		log.Fatalf("seed aborted: mongodb is required and in-memory fallback is disabled")
	}
	var st store.Store = mongoStore
	user, ok := st.GetUserByEmail(cfg.SuperAdminEmail)
	if !ok {
		log.Fatalf("super admin was not created")
	}

	log.Printf("super admin ready: email=%s tenant=%s user_id=%s", user.Email, user.TenantID, user.ID)
}
