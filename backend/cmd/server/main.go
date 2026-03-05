package main

import (
	"log"
	"net/http"

	"github.com/yourorg/mcp-marketplace/backend/internal/auth"
	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	api "github.com/yourorg/mcp-marketplace/backend/internal/http"
	"github.com/yourorg/mcp-marketplace/backend/internal/store"
)

func main() {
	cfg := config.Load()
	st := store.NewMemoryStore()
	jwt := auth.NewJWTManager(cfg.JWTSecret)
	router := api.NewRouter(cfg, st, jwt)

	log.Printf("backend listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
