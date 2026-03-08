package main

import (
	"log"
	"net/http"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/auth"
	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	api "github.com/yourorg/mcp-marketplace/backend/internal/http"
	"github.com/yourorg/mcp-marketplace/backend/internal/store"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}
	mongoStore, err := store.NewMongoStore(cfg)
	if err != nil {
		log.Printf("warning: mongodb connection failed: %v", err)
		log.Fatalf("startup aborted: mongodb is required and in-memory fallback is disabled")
	}
	log.Printf("using mongo store (%s)", cfg.MongoDBName)
	var st store.Store = mongoStore
	jwt, err := auth.NewJWTManager(cfg)
	if err != nil {
		log.Fatalf("failed to initialize jwt manager: %v", err)
	}
	router := api.NewRouter(cfg, st, jwt)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("backend listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
