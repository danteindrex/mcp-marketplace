package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type MemoryStore struct {
	mu           sync.RWMutex
	users        map[string]models.User
	usersByEmail map[string]string
	tenants      map[string]models.Tenant
	servers      map[string]models.Server
	entitlements map[string]models.Entitlement
	hubs         map[string]models.HubProfile
	hubRoutes    map[string][]models.HubRoute
	connections  map[string]models.Connection
	security     map[string]models.SecurityEvent
	audit        map[string]models.AuditLog
	x402         map[string]models.X402Intent
	agents       map[string]models.LocalAgent
	seq          int
}

func NewMemoryStore() *MemoryStore {
	now := time.Now().UTC()
	s := &MemoryStore{
		users:        map[string]models.User{},
		usersByEmail: map[string]string{},
		tenants:      map[string]models.Tenant{},
		servers:      map[string]models.Server{},
		entitlements: map[string]models.Entitlement{},
		hubs:         map[string]models.HubProfile{},
		hubRoutes:    map[string][]models.HubRoute{},
		connections:  map[string]models.Connection{},
		security:     map[string]models.SecurityEvent{},
		audit:        map[string]models.AuditLog{},
		x402:         map[string]models.X402Intent{},
		agents:       map[string]models.LocalAgent{},
	}

	tAdmin := models.Tenant{ID: "tenant_platform", Name: "Platform", Slug: "platform", OwnerUserID: "user_admin", PlanTier: "enterprise", Status: "active", CreatedAt: now}
	tSeller := models.Tenant{ID: "tenant_dataflow", Name: "DataFlow Inc", Slug: "dataflow", OwnerUserID: "user_merchant", PlanTier: "professional", Status: "active", CreatedAt: now}
	tBuyer := models.Tenant{ID: "tenant_acme", Name: "Acme Corp", Slug: "acme", OwnerUserID: "user_buyer", PlanTier: "professional", Status: "active", CreatedAt: now}
	s.tenants[tAdmin.ID] = tAdmin
	s.tenants[tSeller.ID] = tSeller
	s.tenants[tBuyer.ID] = tBuyer

	uAdmin := models.User{ID: "user_admin", TenantID: tAdmin.ID, Email: "admin@platform.local", Name: "Owner Admin", Role: models.RoleAdmin, CreatedAt: now}
	uMerchant := models.User{ID: "user_merchant", TenantID: tSeller.ID, Email: "merchant@dataflow.local", Name: "Seller One", Role: models.RoleMerchant, CreatedAt: now}
	uBuyer := models.User{ID: "user_buyer", TenantID: tBuyer.ID, Email: "buyer@acme.local", Name: "Buyer One", Role: models.RoleBuyer, CreatedAt: now}
	s.putUser(uAdmin)
	s.putUser(uMerchant)
	s.putUser(uBuyer)

	server1 := models.Server{
		ID: "srv_postgres", TenantID: tSeller.ID, Author: "DataFlow Inc", Name: "PostgreSQL Assistant", Slug: "postgresql-assistant",
		Description: "Postgres operations", Category: "data", Version: "2.1.0", DockerImage: "dataflow/postgresql-assistant:2.1.0",
		CanonicalResourceURI: "https://mcp.marketplace.local/resource/srv_postgres", RequiredScopes: []string{"db:read", "db:write"}, PricingType: "subscription", PricingAmount: 29,
		Verified: true, Featured: true, InstallCount: 2300, Rating: 4.8, Status: "published", SupportsCloud: true, SupportsLocal: true, CreatedAt: now, UpdatedAt: now,
	}
	server2 := models.Server{
		ID: "srv_doc", TenantID: tSeller.ID, Author: "DataFlow Inc", Name: "Document Analyzer", Slug: "document-analyzer",
		Description: "AI extraction", Category: "ai", Version: "1.5.2", DockerImage: "docai/document-analyzer:1.5.2",
		CanonicalResourceURI: "https://mcp.marketplace.local/resource/srv_doc", RequiredScopes: []string{"documents:read", "ai:inference"}, PricingType: "x402", PricingAmount: 0.02,
		Verified: true, Featured: false, InstallCount: 1800, Rating: 4.6, Status: "published", SupportsCloud: true, SupportsLocal: true, CreatedAt: now, UpdatedAt: now,
	}
	s.servers[server1.ID] = server1
	s.servers[server2.ID] = server2

	e := models.Entitlement{
		ID: "ent_buyer_postgres", TenantID: tBuyer.ID, UserID: uBuyer.ID, ServerID: server1.ID, AllowedScopes: []string{"db:read", "db:write"},
		CloudAllowed: true, LocalAllowed: true, Status: "active", CreatedAt: now, UpdatedAt: now,
	}
	s.entitlements[e.ID] = e

	h := models.HubProfile{
		ID: "hub_buyer", TenantID: tBuyer.ID, UserID: uBuyer.ID, HubURL: "https://mcp.marketplace.local/hub/tenant_acme/user_buyer", CatalogVersion: 1,
		CatalogHash: hashCatalog([]string{"srv_postgres"}), Status: "active", CreatedAt: now, UpdatedAt: now,
	}
	s.hubs[keyHub(tBuyer.ID, uBuyer.ID)] = h
	s.hubRoutes[h.ID] = []models.HubRoute{{ID: "route1", HubID: h.ID, ServerID: server1.ID, ToolName: "query_postgres", UpstreamType: "cloud", Priority: 1, Enabled: true, UpdatedAt: now}}

	s.security["sec_1"] = models.SecurityEvent{ID: "sec_1", TenantID: tBuyer.ID, Type: "token_reuse", Severity: "high", Description: "Refresh token reuse detected", Actor: uBuyer.ID, TargetID: server1.ID, Resolved: true, CreatedAt: now}
	s.audit["aud_1"] = models.AuditLog{ID: "aud_1", TenantID: tBuyer.ID, ActorID: uBuyer.ID, Action: "server.connect", TargetType: "server", TargetID: server1.ID, Outcome: "success", Metadata: map[string]interface{}{"client": "vscode"}, CreatedAt: now}
	return s
}

func (s *MemoryStore) next(prefix string) string {
	s.seq++
	return fmt.Sprintf("%s_%d", prefix, s.seq)
}

func (s *MemoryStore) putUser(user models.User) {
	s.users[user.ID] = user
	s.usersByEmail[strings.ToLower(user.Email)] = user.ID
}

func keyHub(tenantID, userID string) string {
	return tenantID + ":" + userID
}

func hashCatalog(items []string) string {
	sorted := make([]string, len(items))
	copy(sorted, items)
	sort.Strings(sorted)
	h := sha256.Sum256([]byte(strings.Join(sorted, ",")))
	return hex.EncodeToString(h[:])
}

func (s *MemoryStore) GetUserByEmail(email string) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.usersByEmail[strings.ToLower(email)]
	if !ok {
		return models.User{}, false
	}
	u, ok := s.users[id]
	return u, ok
}

func (s *MemoryStore) GetUserByID(id string) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *MemoryStore) ListMarketplaceServers() []models.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Server, 0, len(s.servers))
	for _, server := range s.servers {
		if server.Status == "published" {
			out = append(out, server)
		}
	}
	return out
}

func (s *MemoryStore) GetServerBySlug(slug string) (models.Server, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, server := range s.servers {
		if server.Slug == slug {
			return server, true
		}
	}
	return models.Server{}, false
}

func (s *MemoryStore) GetServerByID(id string) (models.Server, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	server, ok := s.servers[id]
	return server, ok
}

func (s *MemoryStore) ListMerchantServers(tenantID string) []models.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Server, 0)
	for _, server := range s.servers {
		if server.TenantID == tenantID {
			out = append(out, server)
		}
	}
	return out
}

func (s *MemoryStore) CreateServer(server models.Server) models.Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	server.ID = s.next("srv")
	server.CreatedAt = time.Now().UTC()
	server.UpdatedAt = server.CreatedAt
	s.servers[server.ID] = server
	return server
}

func (s *MemoryStore) UpdateServer(server models.Server) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.servers[server.ID]; !ok {
		return false
	}
	server.UpdatedAt = time.Now().UTC()
	s.servers[server.ID] = server
	return true
}

func (s *MemoryStore) ListEntitlements(tenantID, userID string) []models.Entitlement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Entitlement, 0)
	for _, e := range s.entitlements {
		if e.TenantID == tenantID && e.UserID == userID && e.Status == "active" {
			out = append(out, e)
		}
	}
	return out
}

func (s *MemoryStore) GrantEntitlement(e models.Entitlement) models.Entitlement {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	e.ID = s.next("ent")
	e.CreatedAt = now
	e.UpdatedAt = now
	e.Status = "active"
	s.entitlements[e.ID] = e
	s.refreshHubRoutesLocked(e.TenantID, e.UserID)
	return e
}

func (s *MemoryStore) GetHubProfile(tenantID, userID string) (models.HubProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h, ok := s.hubs[keyHub(tenantID, userID)]
	return h, ok
}

func (s *MemoryStore) UpsertHubProfile(profile models.HubProfile) models.HubProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	if profile.ID == "" {
		profile.ID = s.next("hub")
		profile.CreatedAt = time.Now().UTC()
	}
	profile.UpdatedAt = time.Now().UTC()
	s.hubs[keyHub(profile.TenantID, profile.UserID)] = profile
	return profile
}

func (s *MemoryStore) ListHubRoutes(hubID string) []models.HubRoute {
	s.mu.RLock()
	defer s.mu.RUnlock()
	routes := s.hubRoutes[hubID]
	out := make([]models.HubRoute, len(routes))
	copy(out, routes)
	return out
}

func (s *MemoryStore) ReplaceHubRoutes(hubID string, routes []models.HubRoute) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hubRoutes[hubID] = routes
}

func (s *MemoryStore) refreshHubRoutesLocked(tenantID, userID string) {
	hk := keyHub(tenantID, userID)
	hub, ok := s.hubs[hk]
	if !ok {
		hub = models.HubProfile{ID: s.next("hub"), TenantID: tenantID, UserID: userID, HubURL: fmt.Sprintf("https://mcp.marketplace.local/hub/%s/%s", tenantID, userID), Status: "active", CreatedAt: time.Now().UTC()}
	}
	serverIDs := make([]string, 0)
	routes := make([]models.HubRoute, 0)
	for _, e := range s.entitlements {
		if e.TenantID == tenantID && e.UserID == userID && e.Status == "active" {
			srv, ok := s.servers[e.ServerID]
			if !ok {
				continue
			}
			serverIDs = append(serverIDs, srv.ID)
			routes = append(routes, models.HubRoute{ID: s.next("route"), HubID: hub.ID, ServerID: srv.ID, ToolName: "invoke_" + strings.ReplaceAll(srv.Slug, "-", "_"), UpstreamType: "cloud", Priority: 1, Enabled: true, UpdatedAt: time.Now().UTC()})
		}
	}
	hub.CatalogVersion++
	hub.CatalogHash = hashCatalog(serverIDs)
	hub.UpdatedAt = time.Now().UTC()
	s.hubs[hk] = hub
	s.hubRoutes[hub.ID] = routes
}

func (s *MemoryStore) ListConnections(tenantID, userID string) []models.Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Connection, 0)
	for _, c := range s.connections {
		if c.TenantID == tenantID && c.UserID == userID {
			out = append(out, c)
		}
	}
	return out
}

func (s *MemoryStore) UpsertConnection(connection models.Connection) models.Connection {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if connection.ID == "" {
		connection.ID = s.next("conn")
		connection.CreatedAt = now
	}
	connection.LastUsedAt = now
	s.connections[connection.ID] = connection
	return connection
}

func (s *MemoryStore) ListTenants() []models.Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Tenant, 0, len(s.tenants))
	for _, tenant := range s.tenants {
		out = append(out, tenant)
	}
	return out
}

func (s *MemoryStore) GetTenantByID(id string) (models.Tenant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tenants[id]
	return t, ok
}

func (s *MemoryStore) ListSecurityEvents() []models.SecurityEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.SecurityEvent, 0, len(s.security))
	for _, e := range s.security {
		out = append(out, e)
	}
	return out
}

func (s *MemoryStore) ListAuditLogs() []models.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.AuditLog, 0, len(s.audit))
	for _, log := range s.audit {
		out = append(out, log)
	}
	return out
}

func (s *MemoryStore) AddAuditLog(log models.AuditLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if log.ID == "" {
		log.ID = s.next("aud")
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	s.audit[log.ID] = log
}

func (s *MemoryStore) CreateX402Intent(intent models.X402Intent) models.X402Intent {
	s.mu.Lock()
	defer s.mu.Unlock()
	intent.ID = s.next("x402")
	intent.CreatedAt = time.Now().UTC()
	intent.Status = "payment_required"
	intent.Challenge = "PAYMENT-REQUIRED"
	s.x402[intent.ID] = intent
	return intent
}

func (s *MemoryStore) SettleX402Intent(id string) (models.X402Intent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	intent, ok := s.x402[id]
	if !ok {
		return models.X402Intent{}, false
	}
	intent.Status = "settled"
	intent.SettledAt = time.Now().UTC()
	s.x402[id] = intent
	return intent, true
}

func (s *MemoryStore) ListX402Intents(tenantID, userID string) []models.X402Intent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.X402Intent, 0)
	for _, intent := range s.x402 {
		if intent.TenantID == tenantID && intent.UserID == userID {
			out = append(out, intent)
		}
	}
	return out
}

func (s *MemoryStore) ListLocalAgents(tenantID, userID string) []models.LocalAgent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.LocalAgent, 0)
	for _, agent := range s.agents {
		if agent.TenantID == tenantID && agent.UserID == userID {
			out = append(out, agent)
		}
	}
	return out
}

func (s *MemoryStore) UpsertLocalAgent(agent models.LocalAgent) models.LocalAgent {
	s.mu.Lock()
	defer s.mu.Unlock()
	if agent.ID == "" {
		agent.ID = s.next("agent")
	}
	agent.LastSeenAt = time.Now().UTC()
	s.agents[agent.ID] = agent
	return agent
}
