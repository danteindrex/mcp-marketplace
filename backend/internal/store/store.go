package store

import (
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type Store interface {
	GetUserByEmail(email string) (models.User, bool)
	GetUserByID(id string) (models.User, bool)
	CreateUser(user models.User) (models.User, bool)
	UpdateUser(user models.User) bool
	CreateTenant(tenant models.Tenant) models.Tenant
	ListMarketplaceServers() []models.Server
	GetServerBySlug(slug string) (models.Server, bool)
	GetServerByID(id string) (models.Server, bool)
	ListMerchantServers(tenantID string) []models.Server
	CreateServer(server models.Server) models.Server
	UpdateServer(server models.Server) bool
	UpsertDeployTaskByServer(task models.DeployTask) models.DeployTask
	GetDeployTaskByServer(serverID string) (models.DeployTask, bool)
	ListDueDeployTasks(now time.Time, limit int) []models.DeployTask
	UpdateDeployTask(task models.DeployTask) bool

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
	GetX402Intent(id string) (models.X402Intent, bool)
	SettleX402Intent(id string) (models.X402Intent, bool)
	ListX402Intents(tenantID, userID string) []models.X402Intent
	ListAllX402Intents() []models.X402Intent
	UpdateX402Intent(intent models.X402Intent) bool
	GetPaymentPolicy(tenantID, userID string) (models.PaymentPolicy, bool)
	UpsertPaymentPolicy(policy models.PaymentPolicy) models.PaymentPolicy
	CreateWalletTopUp(item models.WalletTopUp) models.WalletTopUp
	GetWalletTopUp(id string) (models.WalletTopUp, bool)
	GetWalletTopUpByProviderSession(provider, providerSessionID string) (models.WalletTopUp, bool)
	UpdateWalletTopUp(item models.WalletTopUp) bool
	ListWalletTopUps(tenantID, userID string, limit int) []models.WalletTopUp
	UpsertPaymentFeePolicy(policy models.PaymentFeePolicy) models.PaymentFeePolicy
	GetPaymentFeePolicy(scope, tenantID, serverID string) (models.PaymentFeePolicy, bool)
	ListPaymentFeePolicies() []models.PaymentFeePolicy
	CreateLedgerEntries(entries []models.LedgerEntry) []models.LedgerEntry
	ListLedgerEntries(tenantID string, limit int) []models.LedgerEntry
	UpsertSellerPayoutProfile(profile models.SellerPayoutProfile) models.SellerPayoutProfile
	GetSellerPayoutProfile(tenantID string) (models.SellerPayoutProfile, bool)
	ListSellerPayoutProfiles() []models.SellerPayoutProfile
	CreatePayoutRecord(record models.PayoutRecord) models.PayoutRecord
	UpdatePayoutRecord(record models.PayoutRecord) bool
	ListPayoutRecords(tenantID string, limit int) []models.PayoutRecord
	ListLocalAgents(tenantID, userID string) []models.LocalAgent
	UpsertLocalAgent(agent models.LocalAgent) models.LocalAgent
	GetUserSettings(userID string) (models.UserSettings, bool)
	UpsertUserSettings(settings models.UserSettings) models.UserSettings
}
