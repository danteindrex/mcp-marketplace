package store

import "github.com/yourorg/mcp-marketplace/backend/internal/models"

type Store interface {
	GetUserByEmail(email string) (models.User, bool)
	GetUserByID(id string) (models.User, bool)
	ListMarketplaceServers() []models.Server
	GetServerBySlug(slug string) (models.Server, bool)
	GetServerByID(id string) (models.Server, bool)
	ListMerchantServers(tenantID string) []models.Server
	CreateServer(server models.Server) models.Server
	UpdateServer(server models.Server) bool

	ListEntitlements(tenantID, userID string) []models.Entitlement
	GrantEntitlement(e models.Entitlement) models.Entitlement

	GetHubProfile(tenantID, userID string) (models.HubProfile, bool)
	UpsertHubProfile(profile models.HubProfile) models.HubProfile
	ListHubRoutes(hubID string) []models.HubRoute
	ReplaceHubRoutes(hubID string, routes []models.HubRoute)

	ListConnections(tenantID, userID string) []models.Connection
	UpsertConnection(connection models.Connection) models.Connection

	ListTenants() []models.Tenant
	GetTenantByID(id string) (models.Tenant, bool)
	ListSecurityEvents() []models.SecurityEvent
	ListAuditLogs() []models.AuditLog
	AddAuditLog(log models.AuditLog)
	CreateX402Intent(intent models.X402Intent) models.X402Intent
	SettleX402Intent(id string) (models.X402Intent, bool)
	ListX402Intents(tenantID, userID string) []models.X402Intent
	ListLocalAgents(tenantID, userID string) []models.LocalAgent
	UpsertLocalAgent(agent models.LocalAgent) models.LocalAgent
}
