package store

import (
	"path/filepath"
	"testing"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func TestMemoryStorePersistsUsersToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	dataPath := filepath.Join(tmpDir, "store.json")

	cfg := config.Config{
		SuperAdminEmail:    "admin@platform.local",
		SuperAdminPassword: "AdminPass123!@",
		DataFilePath:       dataPath,
	}

	first := NewMemoryStore(cfg)
	tenant := first.CreateTenant(models.Tenant{
		Name:        "Acme",
		Slug:        "acme",
		OwnerUserID: "",
		PlanTier:    "pro",
		Status:      "active",
	})
	created, ok := first.CreateUser(models.User{
		TenantID:     tenant.ID,
		Email:        "persisted@acme.local",
		Name:         "Persisted User",
		Role:         models.RoleBuyer,
		PasswordHash: "hash",
	})
	if !ok {
		t.Fatalf("expected user creation to succeed")
	}

	second := NewMemoryStore(cfg)
	user, found := second.GetUserByEmail("persisted@acme.local")
	if !found {
		t.Fatalf("expected user to be loaded from persisted store")
	}
	if user.ID != created.ID {
		t.Fatalf("expected same user id after reload: got %s want %s", user.ID, created.ID)
	}
}
