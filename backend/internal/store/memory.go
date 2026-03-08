package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type MemoryStore struct {
	mu           sync.RWMutex
	persistPath  string
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
	paymentPol   map[string]models.PaymentPolicy
	topups       map[string]models.WalletTopUp
	feePol       map[string]models.PaymentFeePolicy
	ledger       map[string]models.LedgerEntry
	payoutProf   map[string]models.SellerPayoutProfile
	payouts      map[string]models.PayoutRecord
	deployTasks  map[string]models.DeployTask
	deployBySrv  map[string]string
	agents       map[string]models.LocalAgent
	userSettings map[string]models.UserSettings
	seq          int
}

type diskState struct {
	Users        map[string]models.User                `json:"users"`
	UsersByEmail map[string]string                     `json:"usersByEmail"`
	Tenants      map[string]models.Tenant              `json:"tenants"`
	Servers      map[string]models.Server              `json:"servers"`
	Entitlements map[string]models.Entitlement         `json:"entitlements"`
	Hubs         map[string]models.HubProfile          `json:"hubs"`
	HubRoutes    map[string][]models.HubRoute          `json:"hubRoutes"`
	Connections  map[string]models.Connection          `json:"connections"`
	Security     map[string]models.SecurityEvent       `json:"security"`
	Audit        map[string]models.AuditLog            `json:"audit"`
	X402         map[string]models.X402Intent          `json:"x402"`
	PaymentPol   map[string]models.PaymentPolicy       `json:"paymentPolicies"`
	TopUps       map[string]models.WalletTopUp         `json:"walletTopups"`
	FeePol       map[string]models.PaymentFeePolicy    `json:"paymentFeePolicies"`
	Ledger       map[string]models.LedgerEntry         `json:"ledgerEntries"`
	PayoutProf   map[string]models.SellerPayoutProfile `json:"sellerPayoutProfiles"`
	Payouts      map[string]models.PayoutRecord        `json:"payoutRecords"`
	DeployTasks  map[string]models.DeployTask          `json:"deployTasks"`
	DeployBySrv  map[string]string                     `json:"deployTasksByServer"`
	Agents       map[string]models.LocalAgent          `json:"agents"`
	UserSettings map[string]models.UserSettings        `json:"userSettings"`
	Seq          int                                   `json:"seq"`
}

func (s *MemoryStore) snapshotLocked() diskState {
	return diskState{
		Users:        s.users,
		UsersByEmail: s.usersByEmail,
		Tenants:      s.tenants,
		Servers:      s.servers,
		Entitlements: s.entitlements,
		Hubs:         s.hubs,
		HubRoutes:    s.hubRoutes,
		Connections:  s.connections,
		Security:     s.security,
		Audit:        s.audit,
		X402:         s.x402,
		PaymentPol:   s.paymentPol,
		TopUps:       s.topups,
		FeePol:       s.feePol,
		Ledger:       s.ledger,
		PayoutProf:   s.payoutProf,
		Payouts:      s.payouts,
		DeployTasks:  s.deployTasks,
		DeployBySrv:  s.deployBySrv,
		Agents:       s.agents,
		UserSettings: s.userSettings,
		Seq:          s.seq,
	}
}

func (s *MemoryStore) persistLocked() {
	if s.persistPath == "" {
		return
	}
	dir := filepath.Dir(s.persistPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	blob, err := json.MarshalIndent(s.snapshotLocked(), "", "  ")
	if err != nil {
		return
	}
	tmp := s.persistPath + ".tmp"
	if err := os.WriteFile(tmp, blob, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, s.persistPath)
}

func (s *MemoryStore) loadFromDisk() bool {
	if s.persistPath == "" {
		return false
	}
	blob, err := os.ReadFile(s.persistPath)
	if err != nil {
		return false
	}
	var state diskState
	if err := json.Unmarshal(blob, &state); err != nil {
		return false
	}
	if state.Users == nil {
		return false
	}
	if state.UsersByEmail == nil {
		state.UsersByEmail = map[string]string{}
	}
	if state.Tenants == nil {
		state.Tenants = map[string]models.Tenant{}
	}
	if state.Servers == nil {
		state.Servers = map[string]models.Server{}
	}
	if state.Entitlements == nil {
		state.Entitlements = map[string]models.Entitlement{}
	}
	if state.Hubs == nil {
		state.Hubs = map[string]models.HubProfile{}
	}
	if state.HubRoutes == nil {
		state.HubRoutes = map[string][]models.HubRoute{}
	}
	if state.Connections == nil {
		state.Connections = map[string]models.Connection{}
	}
	if state.Security == nil {
		state.Security = map[string]models.SecurityEvent{}
	}
	if state.Audit == nil {
		state.Audit = map[string]models.AuditLog{}
	}
	if state.X402 == nil {
		state.X402 = map[string]models.X402Intent{}
	}
	if state.PaymentPol == nil {
		state.PaymentPol = map[string]models.PaymentPolicy{}
	}
	if state.TopUps == nil {
		state.TopUps = map[string]models.WalletTopUp{}
	}
	if state.FeePol == nil {
		state.FeePol = map[string]models.PaymentFeePolicy{}
	}
	if state.Ledger == nil {
		state.Ledger = map[string]models.LedgerEntry{}
	}
	if state.PayoutProf == nil {
		state.PayoutProf = map[string]models.SellerPayoutProfile{}
	}
	if state.Payouts == nil {
		state.Payouts = map[string]models.PayoutRecord{}
	}
	if state.DeployTasks == nil {
		state.DeployTasks = map[string]models.DeployTask{}
	}
	if state.DeployBySrv == nil {
		state.DeployBySrv = map[string]string{}
	}
	if len(state.DeployBySrv) == 0 && len(state.DeployTasks) > 0 {
		for id, task := range state.DeployTasks {
			if strings.TrimSpace(task.ServerID) == "" {
				continue
			}
			state.DeployBySrv[task.ServerID] = id
		}
	}
	if state.Agents == nil {
		state.Agents = map[string]models.LocalAgent{}
	}
	if state.UserSettings == nil {
		state.UserSettings = map[string]models.UserSettings{}
	}

	s.users = state.Users
	s.usersByEmail = state.UsersByEmail
	s.tenants = state.Tenants
	s.servers = state.Servers
	s.entitlements = state.Entitlements
	s.hubs = state.Hubs
	s.hubRoutes = state.HubRoutes
	s.connections = state.Connections
	s.security = state.Security
	s.audit = state.Audit
	s.x402 = state.X402
	s.paymentPol = state.PaymentPol
	s.topups = state.TopUps
	s.feePol = state.FeePol
	s.ledger = state.Ledger
	s.payoutProf = state.PayoutProf
	s.payouts = state.Payouts
	s.deployTasks = state.DeployTasks
	s.deployBySrv = state.DeployBySrv
	s.agents = state.Agents
	s.userSettings = state.UserSettings
	s.seq = state.Seq
	return true
}

func NewMemoryStore(cfg config.Config) *MemoryStore {
	now := time.Now().UTC()
	s := &MemoryStore{
		persistPath:  cfg.DataFilePath,
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
		paymentPol:   map[string]models.PaymentPolicy{},
		topups:       map[string]models.WalletTopUp{},
		feePol:       map[string]models.PaymentFeePolicy{},
		ledger:       map[string]models.LedgerEntry{},
		payoutProf:   map[string]models.SellerPayoutProfile{},
		payouts:      map[string]models.PayoutRecord{},
		deployTasks:  map[string]models.DeployTask{},
		deployBySrv:  map[string]string{},
		agents:       map[string]models.LocalAgent{},
		userSettings: map[string]models.UserSettings{},
	}
	if s.loadFromDisk() {
		// Remove legacy seeded catalog records from older demo builds.
		delete(s.servers, "srv_postgres")
		delete(s.servers, "srv_doc")
		delete(s.tenants, "tenant_catalog")
		for id, ent := range s.entitlements {
			if ent.ServerID == "srv_postgres" || ent.ServerID == "srv_doc" {
				delete(s.entitlements, id)
			}
		}
		s.persistLocked()
		// Ensure bootstrap admin exists for initial operator access.
		if _, exists := s.GetUserByEmail(cfg.SuperAdminEmail); exists {
			return s
		}
	}

	tAdmin := models.Tenant{ID: "tenant_platform", Name: "Platform", Slug: "platform", OwnerUserID: "user_admin", PlanTier: "enterprise", Status: "active", CreatedAt: now}
	s.tenants[tAdmin.ID] = tAdmin

	adminHash, _ := bcrypt.GenerateFromPassword([]byte(cfg.SuperAdminPassword), bcrypt.DefaultCost)
	uAdmin := models.User{
		ID:           "user_admin",
		TenantID:     tAdmin.ID,
		Email:        cfg.SuperAdminEmail,
		Name:         "Owner Admin",
		Locale:       "en-US",
		Timezone:     "America/Los_Angeles",
		Role:         models.RoleAdmin,
		PasswordHash: string(adminHash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.putUser(uAdmin)
	s.userSettings[uAdmin.ID] = models.UserSettings{
		UserID: uAdmin.ID,
		Preferences: models.UserPreferences{
			Theme:          "system",
			Language:       "en",
			Timezone:       uAdmin.Timezone,
			DefaultLanding: "/admin/tenants",
			CompactMode:    false,
		},
		Notifications: models.NotificationSettings{
			ProductUpdates: true,
			SecurityAlerts: true,
			BillingAlerts:  true,
			MarketingEmail: false,
			WeeklyDigest:   true,
		},
		UpdatedAt: now,
	}
	s.persistLocked()
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

func feePolicyKey(scope, tenantID, serverID string) string {
	return strings.ToLower(strings.TrimSpace(scope)) + ":" + strings.TrimSpace(tenantID) + ":" + strings.TrimSpace(serverID)
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

func (s *MemoryStore) CreateUser(user models.User) (models.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	email := strings.ToLower(user.Email)
	if _, exists := s.usersByEmail[email]; exists {
		return models.User{}, false
	}
	user.ID = s.next("user")
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = user.CreatedAt
	if strings.TrimSpace(user.Locale) == "" {
		user.Locale = "en-US"
	}
	if strings.TrimSpace(user.Timezone) == "" {
		user.Timezone = "America/Los_Angeles"
	}
	s.putUser(user)
	defaultLanding := "/buyer/dashboard"
	if user.Role == models.RoleMerchant {
		defaultLanding = "/merchant/onboarding"
	}
	if user.Role == models.RoleAdmin {
		defaultLanding = "/admin/tenants"
	}
	s.userSettings[user.ID] = models.UserSettings{
		UserID: user.ID,
		Preferences: models.UserPreferences{
			Theme:          "system",
			Language:       "en",
			Timezone:       user.Timezone,
			DefaultLanding: defaultLanding,
			CompactMode:    false,
		},
		Notifications: models.NotificationSettings{
			ProductUpdates: true,
			SecurityAlerts: true,
			BillingAlerts:  true,
			MarketingEmail: false,
			WeeklyDigest:   true,
		},
		UpdatedAt: user.CreatedAt,
	}
	s.persistLocked()
	return user, true
}

func (s *MemoryStore) UpdateUser(user models.User) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.users[user.ID]
	if !ok {
		return false
	}
	newEmail := strings.ToLower(strings.TrimSpace(user.Email))
	if newEmail == "" {
		newEmail = current.Email
	}
	if existingID, exists := s.usersByEmail[newEmail]; exists && existingID != user.ID {
		return false
	}
	oldEmail := strings.ToLower(strings.TrimSpace(current.Email))
	if oldEmail != newEmail {
		delete(s.usersByEmail, oldEmail)
		s.usersByEmail[newEmail] = user.ID
	}
	user.Email = newEmail
	user.UpdatedAt = time.Now().UTC()
	s.users[user.ID] = user
	s.persistLocked()
	return true
}

func (s *MemoryStore) CreateTenant(tenant models.Tenant) models.Tenant {
	s.mu.Lock()
	defer s.mu.Unlock()
	tenant.ID = s.next("tenant")
	tenant.CreatedAt = time.Now().UTC()
	s.tenants[tenant.ID] = tenant
	s.persistLocked()
	return tenant
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
	if strings.TrimSpace(server.Status) == "" {
		server.Status = models.ServerStatusDraft
	}
	if strings.TrimSpace(server.DeploymentStatus) == "" {
		server.DeploymentStatus = models.ServerDeploymentPending
	}
	server.CreatedAt = time.Now().UTC()
	server.UpdatedAt = server.CreatedAt
	s.servers[server.ID] = server
	s.persistLocked()
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
	s.persistLocked()
	return true
}

func (s *MemoryStore) UpsertDeployTaskByServer(task models.DeployTask) models.DeployTask {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if existingID, exists := s.deployBySrv[task.ServerID]; exists {
		if existing, ok := s.deployTasks[existingID]; ok {
			if strings.TrimSpace(task.ID) == "" {
				task.ID = existing.ID
			}
			if task.CreatedAt.IsZero() {
				task.CreatedAt = existing.CreatedAt
			}
		}
	}
	if strings.TrimSpace(task.ID) == "" {
		task.ID = s.next("dpl")
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	if task.MaxAttempts <= 0 {
		task.MaxAttempts = 6
	}
	if task.NextAttemptAt.IsZero() {
		task.NextAttemptAt = now
	}
	if strings.TrimSpace(task.Status) == "" {
		task.Status = models.DeployTaskStatusPending
	}
	if task.Status == models.DeployTaskStatusPending {
		task.CompletedAt = time.Time{}
	}
	task.UpdatedAt = now
	s.deployTasks[task.ID] = task
	s.deployBySrv[task.ServerID] = task.ID
	s.persistLocked()
	return task
}

func (s *MemoryStore) GetDeployTaskByServer(serverID string) (models.DeployTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if id, ok := s.deployBySrv[serverID]; ok {
		if task, exists := s.deployTasks[id]; exists {
			return task, true
		}
	}
	for _, task := range s.deployTasks {
		if task.ServerID == serverID {
			return task, true
		}
	}
	return models.DeployTask{}, false
}

func (s *MemoryStore) ListDueDeployTasks(now time.Time, limit int) []models.DeployTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]models.DeployTask, 0)
	for _, task := range s.deployTasks {
		if task.Status != models.DeployTaskStatusPending {
			continue
		}
		if !task.NextAttemptAt.IsZero() && task.NextAttemptAt.After(now) {
			continue
		}
		out = append(out, task)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.NextAttemptAt.Equal(b.NextAttemptAt) {
			return a.CreatedAt.Before(b.CreatedAt)
		}
		if a.NextAttemptAt.IsZero() {
			return true
		}
		if b.NextAttemptAt.IsZero() {
			return false
		}
		return a.NextAttemptAt.Before(b.NextAttemptAt)
	})
	if limit > 0 && len(out) > limit {
		return out[:limit]
	}
	return out
}

func (s *MemoryStore) UpdateDeployTask(task models.DeployTask) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.deployTasks[task.ID]; !ok {
		return false
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = time.Now().UTC()
	}
	s.deployTasks[task.ID] = task
	if strings.TrimSpace(task.ServerID) != "" {
		s.deployBySrv[task.ServerID] = task.ID
	}
	s.persistLocked()
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
	s.persistLocked()
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
	s.persistLocked()
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
	s.persistLocked()
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
	s.persistLocked()
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
	s.persistLocked()
}

func (s *MemoryStore) CreateX402Intent(intent models.X402Intent) models.X402Intent {
	s.mu.Lock()
	defer s.mu.Unlock()
	intent.ID = s.next("x402")
	intent.CreatedAt = time.Now().UTC()
	intent.Status = "payment_required"
	intent.Challenge = "PAYMENT-REQUIRED"
	if intent.Quantity <= 0 {
		intent.Quantity = 1
	}
	if intent.RemainingQuantity < 0 {
		intent.RemainingQuantity = 0
	}
	s.x402[intent.ID] = intent
	s.persistLocked()
	return intent
}

func (s *MemoryStore) GetX402Intent(id string) (models.X402Intent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	intent, ok := s.x402[id]
	return intent, ok
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
	if intent.Quantity <= 0 {
		intent.Quantity = 1
	}
	if intent.RemainingQuantity <= 0 {
		intent.RemainingQuantity = intent.Quantity
	}
	s.x402[id] = intent
	s.persistLocked()
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

func (s *MemoryStore) ListAllX402Intents() []models.X402Intent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.X402Intent, 0, len(s.x402))
	for _, intent := range s.x402 {
		out = append(out, intent)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *MemoryStore) UpdateX402Intent(intent models.X402Intent) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.x402[intent.ID]; !ok {
		return false
	}
	s.x402[intent.ID] = intent
	s.persistLocked()
	return true
}

func (s *MemoryStore) GetPaymentPolicy(tenantID, userID string) (models.PaymentPolicy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pol, ok := s.paymentPol[keyHub(tenantID, userID)]
	return pol, ok
}

func (s *MemoryStore) UpsertPaymentPolicy(policy models.PaymentPolicy) models.PaymentPolicy {
	s.mu.Lock()
	defer s.mu.Unlock()
	policy.UpdatedAt = time.Now().UTC()
	s.paymentPol[keyHub(policy.TenantID, policy.UserID)] = policy
	s.persistLocked()
	return policy
}

func (s *MemoryStore) CreateWalletTopUp(item models.WalletTopUp) models.WalletTopUp {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(item.ID) == "" {
		item.ID = s.next("topup")
	}
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	if strings.TrimSpace(item.Status) == "" {
		item.Status = "pending"
	}
	s.topups[item.ID] = item
	s.persistLocked()
	return item
}

func (s *MemoryStore) GetWalletTopUp(id string) (models.WalletTopUp, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.topups[id]
	return item, ok
}

func (s *MemoryStore) GetWalletTopUpByProviderSession(provider, providerSessionID string) (models.WalletTopUp, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	wantProvider := strings.ToLower(strings.TrimSpace(provider))
	wantSession := strings.TrimSpace(providerSessionID)
	for _, item := range s.topups {
		if strings.ToLower(strings.TrimSpace(item.Provider)) == wantProvider && strings.TrimSpace(item.ProviderSessionID) == wantSession {
			return item, true
		}
	}
	return models.WalletTopUp{}, false
}

func (s *MemoryStore) UpdateWalletTopUp(item models.WalletTopUp) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.topups[item.ID]; !ok {
		return false
	}
	item.UpdatedAt = time.Now().UTC()
	s.topups[item.ID] = item
	s.persistLocked()
	return true
}

func (s *MemoryStore) ListWalletTopUps(tenantID, userID string, limit int) []models.WalletTopUp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.WalletTopUp, 0)
	for _, item := range s.topups {
		if item.TenantID == tenantID && item.UserID == userID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	if limit > 0 && len(out) > limit {
		return out[:limit]
	}
	return out
}

func (s *MemoryStore) UpsertPaymentFeePolicy(policy models.PaymentFeePolicy) models.PaymentFeePolicy {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if strings.TrimSpace(policy.ID) == "" {
		policy.ID = s.next("feepol")
	}
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}
	policy.UpdatedAt = now
	key := feePolicyKey(policy.Scope, policy.TenantID, policy.ServerID)
	s.feePol[key] = policy
	s.persistLocked()
	return policy
}

func (s *MemoryStore) GetPaymentFeePolicy(scope, tenantID, serverID string) (models.PaymentFeePolicy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.feePol[feePolicyKey(scope, tenantID, serverID)]
	return item, ok
}

func (s *MemoryStore) ListPaymentFeePolicies() []models.PaymentFeePolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.PaymentFeePolicy, 0, len(s.feePol))
	for _, item := range s.feePol {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func (s *MemoryStore) CreateLedgerEntries(entries []models.LedgerEntry) []models.LedgerEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	out := make([]models.LedgerEntry, 0, len(entries))
	for _, entry := range entries {
		if strings.TrimSpace(entry.ID) == "" {
			entry.ID = s.next("led")
		}
		if entry.CreatedAt.IsZero() {
			entry.CreatedAt = now
		}
		s.ledger[entry.ID] = entry
		out = append(out, entry)
	}
	s.persistLocked()
	return out
}

func (s *MemoryStore) ListLedgerEntries(tenantID string, limit int) []models.LedgerEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.LedgerEntry, 0)
	for _, entry := range s.ledger {
		if strings.TrimSpace(tenantID) == "" || entry.TenantID == tenantID {
			out = append(out, entry)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	if limit > 0 && len(out) > limit {
		return out[:limit]
	}
	return out
}

func (s *MemoryStore) UpsertSellerPayoutProfile(profile models.SellerPayoutProfile) models.SellerPayoutProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if strings.TrimSpace(profile.ID) == "" {
		profile.ID = s.next("payout_profile")
	}
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now
	s.payoutProf[profile.TenantID] = profile
	s.persistLocked()
	return profile
}

func (s *MemoryStore) GetSellerPayoutProfile(tenantID string) (models.SellerPayoutProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.payoutProf[tenantID]
	return item, ok
}

func (s *MemoryStore) ListSellerPayoutProfiles() []models.SellerPayoutProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.SellerPayoutProfile, 0, len(s.payoutProf))
	for _, item := range s.payoutProf {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func (s *MemoryStore) CreatePayoutRecord(record models.PayoutRecord) models.PayoutRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if strings.TrimSpace(record.ID) == "" {
		record.ID = s.next("payout")
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	record.UpdatedAt = now
	s.payouts[record.ID] = record
	s.persistLocked()
	return record
}

func (s *MemoryStore) UpdatePayoutRecord(record models.PayoutRecord) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.payouts[record.ID]; !ok {
		return false
	}
	record.UpdatedAt = time.Now().UTC()
	s.payouts[record.ID] = record
	s.persistLocked()
	return true
}

func (s *MemoryStore) ListPayoutRecords(tenantID string, limit int) []models.PayoutRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.PayoutRecord, 0)
	for _, item := range s.payouts {
		if strings.TrimSpace(tenantID) == "" || item.TenantID == tenantID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	if limit > 0 && len(out) > limit {
		return out[:limit]
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
	s.persistLocked()
	return agent
}

func (s *MemoryStore) GetUserSettings(userID string) (models.UserSettings, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	settings, ok := s.userSettings[userID]
	return settings, ok
}

func (s *MemoryStore) UpsertUserSettings(settings models.UserSettings) models.UserSettings {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings.UpdatedAt = time.Now().UTC()
	s.userSettings[settings.UserID] = settings
	s.persistLocked()
	return settings
}

func (s *MemoryStore) StoreType() string {
	return "memory"
}

func (s *MemoryStore) Health(context.Context) error {
	return nil
}
