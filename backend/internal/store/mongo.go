package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const mongoTimeout = 8 * time.Second

type MongoStore struct {
	cfg           config.Config
	client        *mongo.Client
	db            *mongo.Database
	users         *mongo.Collection
	tenants       *mongo.Collection
	servers       *mongo.Collection
	entitlements  *mongo.Collection
	hubs          *mongo.Collection
	hubRoutes     *mongo.Collection
	connections   *mongo.Collection
	security      *mongo.Collection
	audit         *mongo.Collection
	x402          *mongo.Collection
	paymentPol    *mongo.Collection
	topups        *mongo.Collection
	feePol        *mongo.Collection
	ledger        *mongo.Collection
	payoutProf    *mongo.Collection
	payouts       *mongo.Collection
	deployTasks   *mongo.Collection
	agents        *mongo.Collection
	userSettings  *mongo.Collection
	oauthAccounts *mongo.Collection
	oauthClients  *mongo.Collection
	oauthCodes    *mongo.Collection
}

func NewMongoStore(cfg config.Config) (*MongoStore, error) {
	uri := strings.TrimSpace(cfg.MongoURI)
	if uri == "" {
		return nil, fmt.Errorf("mongo uri is required")
	}
	dbName := strings.TrimSpace(cfg.MongoDBName)
	if dbName == "" {
		dbName = "mcp_marketplace"
	}
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	db := client.Database(dbName)
	s := &MongoStore{
		cfg:           cfg,
		client:        client,
		db:            db,
		users:         db.Collection("users"),
		tenants:       db.Collection("tenants"),
		servers:       db.Collection("servers"),
		entitlements:  db.Collection("entitlements"),
		hubs:          db.Collection("hubs"),
		hubRoutes:     db.Collection("hub_routes"),
		connections:   db.Collection("connections"),
		security:      db.Collection("security_events"),
		audit:         db.Collection("audit_logs"),
		x402:          db.Collection("x402_intents"),
		paymentPol:    db.Collection("payment_policies"),
		topups:        db.Collection("wallet_topups"),
		feePol:        db.Collection("payment_fee_policies"),
		ledger:        db.Collection("ledger_entries"),
		payoutProf:    db.Collection("seller_payout_profiles"),
		payouts:       db.Collection("payout_records"),
		deployTasks:   db.Collection("deploy_tasks"),
		agents:        db.Collection("local_agents"),
		userSettings:  db.Collection("user_settings"),
		oauthAccounts: db.Collection("oauth_accounts"),
		oauthClients:  db.Collection("oauth_clients"),
		oauthCodes:    db.Collection("oauth_auth_codes"),
	}
	if err := s.ensureIndexes(); err != nil {
		return nil, err
	}
	if err := s.bootstrapSuperAdmin(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *MongoStore) ensureIndexes() error {
	type idx struct {
		col   *mongo.Collection
		model []mongo.IndexModel
	}
	ttlDesc := -1
	all := []idx{
		{
			col: s.users,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}}},
			},
		},
		{
			col: s.tenants,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			col: s.servers,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "status", Value: 1}}},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "deploymentStatus", Value: 1}}},
				{Keys: bson.D{{Key: "status", Value: 1}, {Key: "updatedAt", Value: ttlDesc}}},
			},
		},
		{
			col: s.entitlements,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "userId", Value: 1}, {Key: "serverId", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "userId", Value: 1}, {Key: "status", Value: 1}}},
			},
		},
		{
			col: s.hubs,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			col: s.hubRoutes,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "hubId", Value: 1}, {Key: "serverId", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			col: s.connections,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "userId", Value: 1}}},
				{Keys: bson.D{{Key: "tokenExpiresAt", Value: 1}}},
			},
		},
		{
			col: s.security,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "createdAt", Value: ttlDesc}}},
			},
		},
		{
			col: s.audit,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "createdAt", Value: ttlDesc}}},
			},
		},
		{
			col: s.x402,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "userId", Value: 1}, {Key: "createdAt", Value: ttlDesc}}},
				{Keys: bson.D{{Key: "status", Value: 1}}},
				{Keys: bson.D{{Key: "paymentIdentifier", Value: 1}}},
			},
		},
		{
			col: s.paymentPol,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			col: s.topups,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "userId", Value: 1}, {Key: "createdAt", Value: ttlDesc}}},
				{Keys: bson.D{{Key: "provider", Value: 1}, {Key: "providerSessionId", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			col: s.feePol,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "scope", Value: 1}, {Key: "tenantId", Value: 1}, {Key: "serverId", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			col: s.ledger,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "transactionId", Value: 1}, {Key: "createdAt", Value: ttlDesc}}},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "createdAt", Value: ttlDesc}}},
				{Keys: bson.D{{Key: "intentId", Value: 1}}},
			},
		},
		{
			col: s.payoutProf,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "stripeAccountId", Value: 1}}},
			},
		},
		{
			col: s.payouts,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "createdAt", Value: ttlDesc}}},
				{Keys: bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: ttlDesc}}},
			},
		},
		{
			col: s.deployTasks,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "serverId", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "status", Value: 1}, {Key: "nextAttemptAt", Value: 1}}},
			},
		},
		{
			col: s.agents,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "userId", Value: 1}, {Key: "deviceId", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			col: s.userSettings,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			col: s.oauthAccounts,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "provider", Value: 1}, {Key: "providerId", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "userId", Value: 1}}},
			},
		},
		{
			col: s.oauthClients,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "clientId", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			col: s.oauthCodes,
			model: []mongo.IndexModel{
				{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "code", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "clientId", Value: 1}}},
				{Keys: bson.D{{Key: "userId", Value: 1}}},
				// TTL index for auto-expiration of auth codes
				{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
			},
		},
	}
	for _, entry := range all {
		ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
		_, err := entry.col.Indexes().CreateMany(ctx, entry.model)
		cancel()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MongoStore) bootstrapSuperAdmin() error {
	if strings.TrimSpace(s.cfg.SuperAdminEmail) == "" {
		return nil
	}
	if _, exists := s.GetUserByEmail(s.cfg.SuperAdminEmail); exists {
		return nil
	}
	now := time.Now().UTC()
	if _, ok := s.GetTenantByID("tenant_platform"); !ok {
		ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
		_, err := s.tenants.InsertOne(ctx, models.Tenant{
			ID:          "tenant_platform",
			Name:        "Platform",
			Slug:        "platform",
			OwnerUserID: "user_admin",
			PlanTier:    "enterprise",
			Status:      "active",
			CreatedAt:   now,
		})
		cancel()
		if err != nil && !isDuplicateKey(err) {
			return err
		}
	}
	adminHash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.SuperAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	adminUser := models.User{
		ID:           "user_admin",
		TenantID:     "tenant_platform",
		Email:        strings.ToLower(strings.TrimSpace(s.cfg.SuperAdminEmail)),
		Name:         "Owner Admin",
		Locale:       "en-US",
		Timezone:     "America/Los_Angeles",
		Role:         models.RoleAdmin,
		PasswordHash: string(adminHash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	_, err = s.users.InsertOne(ctx, adminUser)
	cancel()
	if err != nil && !isDuplicateKey(err) {
		return err
	}
	s.UpsertUserSettings(models.UserSettings{
		UserID: adminUser.ID,
		Preferences: models.UserPreferences{
			Theme:          "system",
			Language:       "en",
			Timezone:       adminUser.Timezone,
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
	})
	return nil
}

func newPrefixedID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, primitive.NewObjectID().Hex())
}

func isDuplicateKey(err error) bool {
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		return writeErr.HasErrorCode(11000)
	}
	var bulkErr mongo.BulkWriteException
	if errors.As(err, &bulkErr) {
		return bulkErr.HasErrorCode(11000)
	}
	return false
}

func decodeAll[T any](cur *mongo.Cursor) ([]T, error) {
	defer cur.Close(context.Background())
	out := make([]T, 0)
	for cur.Next(context.Background()) {
		var item T
		if err := cur.Decode(&item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, cur.Err()
}

func (s *MongoStore) GetUserByEmail(email string) (models.User, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return models.User{}, false
	}
	var u models.User
	err := s.users.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		return models.User{}, false
	}
	return u, true
}

func (s *MongoStore) GetUserByID(id string) (models.User, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var u models.User
	err := s.users.FindOne(ctx, bson.M{"id": id}).Decode(&u)
	if err != nil {
		return models.User{}, false
	}
	return u, true
}

func (s *MongoStore) CreateUser(user models.User) (models.User, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	now := time.Now().UTC()
	user.ID = newPrefixedID("user")
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.CreatedAt = now
	user.UpdatedAt = now
	if strings.TrimSpace(user.Locale) == "" {
		user.Locale = "en-US"
	}
	if strings.TrimSpace(user.Timezone) == "" {
		user.Timezone = "America/Los_Angeles"
	}
	if _, err := s.users.InsertOne(ctx, user); err != nil {
		return models.User{}, false
	}
	defaultLanding := "/buyer/dashboard"
	if user.Role == models.RoleMerchant {
		defaultLanding = "/merchant/onboarding"
	}
	if user.Role == models.RoleAdmin {
		defaultLanding = "/admin/tenants"
	}
	s.UpsertUserSettings(models.UserSettings{
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
		UpdatedAt: now,
	})
	return user, true
}

func (s *MongoStore) UpdateUser(user models.User) bool {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	current, ok := s.GetUserByID(user.ID)
	if !ok {
		return false
	}
	newEmail := strings.ToLower(strings.TrimSpace(user.Email))
	if newEmail == "" {
		newEmail = current.Email
	}
	user.Email = newEmail
	user.UpdatedAt = time.Now().UTC()
	res, err := s.users.ReplaceOne(ctx, bson.M{"id": user.ID}, user)
	if err != nil {
		return false
	}
	return res.MatchedCount > 0
}

func (s *MongoStore) GetOAuthAccount(provider models.OAuthProvider, providerID string) (models.OAuthAccount, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var account models.OAuthAccount
	err := s.oauthAccounts.FindOne(ctx, bson.M{"provider": provider, "providerId": providerID}).Decode(&account)
	if err != nil {
		return models.OAuthAccount{}, false
	}
	return account, true
}

func (s *MongoStore) GetOAuthAccountsByUserID(userID string) []models.OAuthAccount {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.oauthAccounts.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return []models.OAuthAccount{}
	}
	accounts, err := decodeAll[models.OAuthAccount](cur)
	if err != nil {
		return []models.OAuthAccount{}
	}
	return accounts
}

func (s *MongoStore) CreateOAuthAccount(account models.OAuthAccount) models.OAuthAccount {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	account.ID = newPrefixedID("oauth")
	account.CreatedAt = time.Now().UTC()
	account.UpdatedAt = account.CreatedAt
	_, _ = s.oauthAccounts.InsertOne(ctx, account)
	return account
}

func (s *MongoStore) UpdateOAuthAccount(account models.OAuthAccount) bool {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	account.UpdatedAt = time.Now().UTC()
	res, err := s.oauthAccounts.ReplaceOne(ctx, bson.M{"id": account.ID}, account)
	if err != nil {
		return false
	}
	return res.MatchedCount > 0
}

func (s *MongoStore) DeleteOAuthAccount(provider models.OAuthProvider, providerID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	res, err := s.oauthAccounts.DeleteOne(ctx, bson.M{"provider": provider, "providerId": providerID})
	if err != nil {
		return false
	}
	return res.DeletedCount > 0
}

func (s *MongoStore) CreateTenant(tenant models.Tenant) models.Tenant {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	tenant.ID = newPrefixedID("tenant")
	tenant.CreatedAt = time.Now().UTC()
	_, _ = s.tenants.InsertOne(ctx, tenant)
	return tenant
}

func (s *MongoStore) ListMarketplaceServers() []models.Server {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.servers.Find(ctx, bson.M{"status": models.ServerStatusPublished, "deploymentStatus": models.ServerDeploymentDeployed})
	if err != nil {
		return []models.Server{}
	}
	items, err := decodeAll[models.Server](cur)
	if err != nil {
		return []models.Server{}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].InstallCount > items[j].InstallCount
	})
	return items
}

func (s *MongoStore) GetServerBySlug(slug string) (models.Server, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var server models.Server
	err := s.servers.FindOne(ctx, bson.M{"slug": slug, "status": models.ServerStatusPublished, "deploymentStatus": models.ServerDeploymentDeployed}).Decode(&server)
	if err != nil {
		return models.Server{}, false
	}
	return server, true
}

func (s *MongoStore) GetServerByID(id string) (models.Server, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var server models.Server
	err := s.servers.FindOne(ctx, bson.M{"id": id}).Decode(&server)
	if err != nil {
		return models.Server{}, false
	}
	return server, true
}

func (s *MongoStore) ListMerchantServers(tenantID string) []models.Server {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.servers.Find(ctx, bson.M{"tenantId": tenantID})
	if err != nil {
		return []models.Server{}
	}
	items, err := decodeAll[models.Server](cur)
	if err != nil {
		return []models.Server{}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	return items
}

func (s *MongoStore) CreateServer(server models.Server) models.Server {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	now := time.Now().UTC()
	server.ID = newPrefixedID("srv")
	server.Status = models.ServerStatusDraft
	server.DeploymentStatus = models.ServerDeploymentPending
	server.PublishedAt = time.Time{}
	server.DeployedAt = time.Time{}
	server.CreatedAt = now
	server.UpdatedAt = now
	_, _ = s.servers.InsertOne(ctx, server)
	return server
}

func (s *MongoStore) UpdateServer(server models.Server) bool {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	server.UpdatedAt = time.Now().UTC()
	res, err := s.servers.ReplaceOne(ctx, bson.M{"id": server.ID}, server)
	if err != nil {
		return false
	}
	return res.MatchedCount > 0
}

func (s *MongoStore) UpsertDeployTaskByServer(task models.DeployTask) models.DeployTask {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	now := time.Now().UTC()
	if strings.TrimSpace(task.ID) == "" {
		task.ID = newPrefixedID("dpl")
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
	task.UpdatedAt = now
	if task.Status == models.DeployTaskStatusPending {
		task.CompletedAt = time.Time{}
	}

	filter := bson.M{"serverId": strings.TrimSpace(task.ServerID)}
	update := bson.M{
		"$set": bson.M{
			"tenantId":            task.TenantID,
			"serverId":            task.ServerID,
			"requestedBy":         task.RequestedBy,
			"preferredWorkflowId": task.PreferredWorkflowID,
			"deploymentTarget":    task.DeploymentTarget,
			"status":              task.Status,
			"attemptCount":        task.AttemptCount,
			"maxAttempts":         task.MaxAttempts,
			"nextAttemptAt":       task.NextAttemptAt,
			"lastError":           task.LastError,
			"updatedAt":           task.UpdatedAt,
			"completedAt":         task.CompletedAt,
		},
		"$setOnInsert": bson.M{
			"id":        task.ID,
			"createdAt": task.CreatedAt,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var out models.DeployTask
	if err := s.deployTasks.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out); err != nil {
		return task
	}
	return out
}

func (s *MongoStore) GetDeployTaskByServer(serverID string) (models.DeployTask, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	var task models.DeployTask
	err := s.deployTasks.FindOne(ctx, bson.M{"serverId": strings.TrimSpace(serverID)}).Decode(&task)
	if err != nil {
		return models.DeployTask{}, false
	}
	return task, true
}

func (s *MongoStore) ListDueDeployTasks(now time.Time, limit int) []models.DeployTask {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	filter := bson.M{
		"status":        models.DeployTaskStatusPending,
		"nextAttemptAt": bson.M{"$lte": now},
	}
	opts := options.Find().SetSort(bson.D{{Key: "nextAttemptAt", Value: 1}, {Key: "createdAt", Value: 1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cur, err := s.deployTasks.Find(ctx, filter, opts)
	if err != nil {
		return []models.DeployTask{}
	}
	items, err := decodeAll[models.DeployTask](cur)
	if err != nil {
		return []models.DeployTask{}
	}
	return items
}

func (s *MongoStore) UpdateDeployTask(task models.DeployTask) bool {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = time.Now().UTC()
	}
	res, err := s.deployTasks.ReplaceOne(ctx, bson.M{"id": task.ID}, task)
	if err != nil {
		return false
	}
	return res.MatchedCount > 0
}

func (s *MongoStore) ListEntitlements(tenantID, userID string) []models.Entitlement {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.entitlements.Find(ctx, bson.M{
		"tenantId": tenantID,
		"userId":   userID,
		"status":   "active",
	})
	if err != nil {
		return []models.Entitlement{}
	}
	items, err := decodeAll[models.Entitlement](cur)
	if err != nil {
		return []models.Entitlement{}
	}
	return items
}

func (s *MongoStore) GrantEntitlement(e models.Entitlement) models.Entitlement {
	now := time.Now().UTC()
	e.Status = "active"
	e.UpdatedAt = now
	e.CreatedAt = now
	e.ID = newPrefixedID("ent")
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	filter := bson.M{
		"tenantId": e.TenantID,
		"userId":   e.UserID,
		"serverId": e.ServerID,
	}
	update := bson.M{
		"$set": bson.M{
			"allowedScopes": e.AllowedScopes,
			"cloudAllowed":  e.CloudAllowed,
			"localAllowed":  e.LocalAllowed,
			"status":        "active",
			"updatedAt":     now,
		},
		"$setOnInsert": bson.M{
			"id":        e.ID,
			"tenantId":  e.TenantID,
			"userId":    e.UserID,
			"serverId":  e.ServerID,
			"createdAt": now,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var out models.Entitlement
	if err := s.entitlements.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out); err != nil {
		return e
	}
	_ = s.rebuildHubRoutes(out.TenantID, out.UserID)
	return out
}

func (s *MongoStore) GetHubProfile(tenantID, userID string) (models.HubProfile, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var hub models.HubProfile
	err := s.hubs.FindOne(ctx, bson.M{"tenantId": tenantID, "userId": userID}).Decode(&hub)
	if err != nil {
		return models.HubProfile{}, false
	}
	return hub, true
}

func (s *MongoStore) UpsertHubProfile(profile models.HubProfile) models.HubProfile {
	now := time.Now().UTC()
	if profile.ID == "" {
		profile.ID = newPrefixedID("hub")
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	filter := bson.M{"tenantId": profile.TenantID, "userId": profile.UserID}
	update := bson.M{
		"$set": bson.M{
			"id":             profile.ID,
			"tenantId":       profile.TenantID,
			"userId":         profile.UserID,
			"hubUrl":         profile.HubURL,
			"catalogHash":    profile.CatalogHash,
			"catalogVersion": profile.CatalogVersion,
			"status":         profile.Status,
			"updatedAt":      profile.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"createdAt": profile.CreatedAt,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var out models.HubProfile
	if err := s.hubs.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out); err != nil {
		return profile
	}
	return out
}

func (s *MongoStore) ListHubRoutes(hubID string) []models.HubRoute {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.hubRoutes.Find(ctx, bson.M{"hubId": hubID})
	if err != nil {
		return []models.HubRoute{}
	}
	items, err := decodeAll[models.HubRoute](cur)
	if err != nil {
		return []models.HubRoute{}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Priority < items[j].Priority
	})
	return items
}

func (s *MongoStore) ReplaceHubRoutes(hubID string, routes []models.HubRoute) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	_, _ = s.hubRoutes.DeleteMany(ctx, bson.M{"hubId": hubID})
	if len(routes) == 0 {
		return
	}
	now := time.Now().UTC()
	docs := make([]interface{}, 0, len(routes))
	for i := range routes {
		if routes[i].ID == "" {
			routes[i].ID = newPrefixedID("route")
		}
		if routes[i].UpdatedAt.IsZero() {
			routes[i].UpdatedAt = now
		}
		docs = append(docs, routes[i])
	}
	_, _ = s.hubRoutes.InsertMany(ctx, docs)
}

func (s *MongoStore) rebuildHubRoutes(tenantID, userID string) error {
	hub, ok := s.GetHubProfile(tenantID, userID)
	if !ok {
		hub = s.UpsertHubProfile(models.HubProfile{
			TenantID: tenantID,
			UserID:   userID,
			HubURL:   s.cfg.BaseURL + "/mcp/hub/" + tenantID + "/" + userID,
			Status:   "active",
		})
	}
	entitlements := s.ListEntitlements(tenantID, userID)
	serverIDs := make([]string, 0, len(entitlements))
	routes := make([]models.HubRoute, 0, len(entitlements))
	now := time.Now().UTC()
	for _, e := range entitlements {
		srv, ok := s.GetServerByID(e.ServerID)
		if !ok {
			continue
		}
		serverIDs = append(serverIDs, srv.ID)
		routes = append(routes, models.HubRoute{
			ID:           newPrefixedID("route"),
			HubID:        hub.ID,
			ServerID:     srv.ID,
			ToolName:     "invoke_" + strings.ReplaceAll(srv.Slug, "-", "_"),
			UpstreamType: "cloud",
			Priority:     1,
			Enabled:      true,
			UpdatedAt:    now,
		})
	}
	hub.CatalogVersion++
	hub.CatalogHash = hashCatalog(serverIDs)
	hub.UpdatedAt = now
	s.UpsertHubProfile(hub)
	s.ReplaceHubRoutes(hub.ID, routes)
	return nil
}

func (s *MongoStore) ListConnections(tenantID, userID string) []models.Connection {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.connections.Find(ctx, bson.M{"tenantId": tenantID, "userId": userID})
	if err != nil {
		return []models.Connection{}
	}
	items, err := decodeAll[models.Connection](cur)
	if err != nil {
		return []models.Connection{}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].LastUsedAt.After(items[j].LastUsedAt)
	})
	return items
}

func (s *MongoStore) UpsertConnection(connection models.Connection) models.Connection {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	now := time.Now().UTC()
	if connection.ID == "" {
		connection.ID = newPrefixedID("conn")
		connection.CreatedAt = now
	}
	if connection.CreatedAt.IsZero() {
		connection.CreatedAt = now
	}
	connection.LastUsedAt = now
	_, _ = s.connections.ReplaceOne(
		ctx,
		bson.M{"id": connection.ID},
		connection,
		options.Replace().SetUpsert(true),
	)
	return connection
}

func (s *MongoStore) ListTenants() []models.Tenant {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.tenants.Find(ctx, bson.M{})
	if err != nil {
		return []models.Tenant{}
	}
	items, err := decodeAll[models.Tenant](cur)
	if err != nil {
		return []models.Tenant{}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *MongoStore) GetTenantByID(id string) (models.Tenant, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var tenant models.Tenant
	err := s.tenants.FindOne(ctx, bson.M{"id": id}).Decode(&tenant)
	if err != nil {
		return models.Tenant{}, false
	}
	return tenant, true
}

func (s *MongoStore) ListSecurityEvents() []models.SecurityEvent {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.security.Find(ctx, bson.M{})
	if err != nil {
		return []models.SecurityEvent{}
	}
	items, err := decodeAll[models.SecurityEvent](cur)
	if err != nil {
		return []models.SecurityEvent{}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *MongoStore) ListAuditLogs() []models.AuditLog {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.audit.Find(ctx, bson.M{})
	if err != nil {
		return []models.AuditLog{}
	}
	items, err := decodeAll[models.AuditLog](cur)
	if err != nil {
		return []models.AuditLog{}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *MongoStore) AddAuditLog(log models.AuditLog) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	if log.ID == "" {
		log.ID = newPrefixedID("aud")
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	_, _ = s.audit.InsertOne(ctx, log)
}

func (s *MongoStore) CreateX402Intent(intent models.X402Intent) models.X402Intent {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	intent.ID = newPrefixedID("x402")
	intent.CreatedAt = time.Now().UTC()
	intent.Status = "payment_required"
	if strings.TrimSpace(intent.Challenge) == "" {
		intent.Challenge = "PAYMENT-REQUIRED"
	}
	if intent.Quantity <= 0 {
		intent.Quantity = 1
	}
	if intent.RemainingQuantity < 0 {
		intent.RemainingQuantity = 0
	}
	_, _ = s.x402.InsertOne(ctx, intent)
	return intent
}

func (s *MongoStore) GetX402Intent(id string) (models.X402Intent, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var intent models.X402Intent
	err := s.x402.FindOne(ctx, bson.M{"id": id}).Decode(&intent)
	if err != nil {
		return models.X402Intent{}, false
	}
	return intent, true
}

func (s *MongoStore) SettleX402Intent(id string) (models.X402Intent, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var existing models.X402Intent
	if err := s.x402.FindOne(ctx, bson.M{"id": id}).Decode(&existing); err != nil {
		return models.X402Intent{}, false
	}
	qty := existing.Quantity
	if qty <= 0 {
		qty = 1
	}
	remaining := existing.RemainingQuantity
	if remaining <= 0 {
		remaining = qty
	}
	update := bson.M{
		"$set": bson.M{
			"status":            "settled",
			"settledAt":         time.Now().UTC(),
			"quantity":          qty,
			"remainingQuantity": remaining,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var intent models.X402Intent
	if err := s.x402.FindOneAndUpdate(ctx, bson.M{"id": id}, update, opts).Decode(&intent); err != nil {
		return models.X402Intent{}, false
	}
	return intent, true
}

func (s *MongoStore) ListX402Intents(tenantID, userID string) []models.X402Intent {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.x402.Find(ctx, bson.M{"tenantId": tenantID, "userId": userID})
	if err != nil {
		return []models.X402Intent{}
	}
	items, err := decodeAll[models.X402Intent](cur)
	if err != nil {
		return []models.X402Intent{}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *MongoStore) ListAllX402Intents() []models.X402Intent {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.x402.Find(ctx, bson.M{})
	if err != nil {
		return []models.X402Intent{}
	}
	items, err := decodeAll[models.X402Intent](cur)
	if err != nil {
		return []models.X402Intent{}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *MongoStore) UpdateX402Intent(intent models.X402Intent) bool {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	res, err := s.x402.ReplaceOne(ctx, bson.M{"id": intent.ID}, intent)
	if err != nil {
		return false
	}
	return res.MatchedCount > 0
}

func (s *MongoStore) GetPaymentPolicy(tenantID, userID string) (models.PaymentPolicy, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var policy models.PaymentPolicy
	err := s.paymentPol.FindOne(ctx, bson.M{"tenantId": tenantID, "userId": userID}).Decode(&policy)
	if err != nil {
		return models.PaymentPolicy{}, false
	}
	return policy, true
}

func (s *MongoStore) UpsertPaymentPolicy(policy models.PaymentPolicy) models.PaymentPolicy {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	policy.UpdatedAt = time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"tenantId":            policy.TenantID,
			"userId":              policy.UserID,
			"monthlySpendCapUsdc": policy.MonthlySpendCapUSDC,
			"dailySpendCapUsdc":   policy.DailySpendCapUSDC,
			"perCallCapUsdc":      policy.PerCallCapUSDC,
			"allowedMethods":      policy.AllowedMethods,
			"siwxWallet":          policy.SIWXWallet,
			"walletBalanceUsdc":   policy.WalletBalanceUSDC,
			"minimumBalanceUsdc":  policy.MinimumBalanceUSDC,
			"hardStopOnLowFunds":  policy.HardStopOnLowFunds,
			"autoTopUpEnabled":    policy.AutoTopUpEnabled,
			"autoTopUpAmountUsd":  policy.AutoTopUpAmountUSD,
			"autoTopUpTriggerUsd": policy.AutoTopUpTriggerUSD,
			"fundingMethod":       policy.FundingMethod,
			"walletAddress":       policy.WalletAddress,
			"lastTopUpAt":         policy.LastTopUpAt,
			"updatedAt":           policy.UpdatedAt,
		},
	}
	_, _ = s.paymentPol.UpdateOne(ctx, bson.M{"tenantId": policy.TenantID, "userId": policy.UserID}, update, options.Update().SetUpsert(true))
	return policy
}

func (s *MongoStore) CreateWalletTopUp(item models.WalletTopUp) models.WalletTopUp {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	if strings.TrimSpace(item.ID) == "" {
		item.ID = newPrefixedID("topup")
	}
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	if strings.TrimSpace(item.Status) == "" {
		item.Status = "pending"
	}
	_, _ = s.topups.InsertOne(ctx, item)
	return item
}

func (s *MongoStore) GetWalletTopUp(id string) (models.WalletTopUp, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var item models.WalletTopUp
	err := s.topups.FindOne(ctx, bson.M{"id": id}).Decode(&item)
	if err != nil {
		return models.WalletTopUp{}, false
	}
	return item, true
}

func (s *MongoStore) GetWalletTopUpByProviderSession(provider, providerSessionID string) (models.WalletTopUp, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var item models.WalletTopUp
	err := s.topups.FindOne(ctx, bson.M{
		"provider":          strings.ToLower(strings.TrimSpace(provider)),
		"providerSessionId": strings.TrimSpace(providerSessionID),
	}).Decode(&item)
	if err != nil {
		return models.WalletTopUp{}, false
	}
	return item, true
}

func (s *MongoStore) UpdateWalletTopUp(item models.WalletTopUp) bool {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	item.UpdatedAt = time.Now().UTC()
	res, err := s.topups.ReplaceOne(ctx, bson.M{"id": item.ID}, item)
	if err != nil {
		return false
	}
	return res.MatchedCount > 0
}

func (s *MongoStore) ListWalletTopUps(tenantID, userID string, limit int) []models.WalletTopUp {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cur, err := s.topups.Find(ctx, bson.M{"tenantId": tenantID, "userId": userID}, opts)
	if err != nil {
		return []models.WalletTopUp{}
	}
	items, err := decodeAll[models.WalletTopUp](cur)
	if err != nil {
		return []models.WalletTopUp{}
	}
	return items
}

func (s *MongoStore) UpsertPaymentFeePolicy(policy models.PaymentFeePolicy) models.PaymentFeePolicy {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	now := time.Now().UTC()
	if strings.TrimSpace(policy.ID) == "" {
		policy.ID = newPrefixedID("feepol")
	}
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}
	policy.UpdatedAt = now
	filter := bson.M{
		"scope":    strings.ToLower(strings.TrimSpace(policy.Scope)),
		"tenantId": strings.TrimSpace(policy.TenantID),
		"serverId": strings.TrimSpace(policy.ServerID),
	}
	_, _ = s.feePol.ReplaceOne(ctx, filter, policy, options.Replace().SetUpsert(true))
	return policy
}

func (s *MongoStore) GetPaymentFeePolicy(scope, tenantID, serverID string) (models.PaymentFeePolicy, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var policy models.PaymentFeePolicy
	err := s.feePol.FindOne(ctx, bson.M{
		"scope":    strings.ToLower(strings.TrimSpace(scope)),
		"tenantId": strings.TrimSpace(tenantID),
		"serverId": strings.TrimSpace(serverID),
	}).Decode(&policy)
	if err != nil {
		return models.PaymentFeePolicy{}, false
	}
	return policy, true
}

func (s *MongoStore) ListPaymentFeePolicies() []models.PaymentFeePolicy {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.feePol.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}))
	if err != nil {
		return []models.PaymentFeePolicy{}
	}
	items, err := decodeAll[models.PaymentFeePolicy](cur)
	if err != nil {
		return []models.PaymentFeePolicy{}
	}
	return items
}

func (s *MongoStore) CreateLedgerEntries(entries []models.LedgerEntry) []models.LedgerEntry {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	now := time.Now().UTC()
	docs := make([]interface{}, 0, len(entries))
	out := make([]models.LedgerEntry, 0, len(entries))
	for _, entry := range entries {
		if strings.TrimSpace(entry.ID) == "" {
			entry.ID = newPrefixedID("led")
		}
		if entry.CreatedAt.IsZero() {
			entry.CreatedAt = now
		}
		docs = append(docs, entry)
		out = append(out, entry)
	}
	if len(docs) > 0 {
		_, _ = s.ledger.InsertMany(ctx, docs)
	}
	return out
}

func (s *MongoStore) ListLedgerEntries(tenantID string, limit int) []models.LedgerEntry {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	filter := bson.M{}
	if strings.TrimSpace(tenantID) != "" {
		filter["tenantId"] = strings.TrimSpace(tenantID)
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cur, err := s.ledger.Find(ctx, filter, opts)
	if err != nil {
		return []models.LedgerEntry{}
	}
	items, err := decodeAll[models.LedgerEntry](cur)
	if err != nil {
		return []models.LedgerEntry{}
	}
	return items
}

func (s *MongoStore) UpsertSellerPayoutProfile(profile models.SellerPayoutProfile) models.SellerPayoutProfile {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	now := time.Now().UTC()
	if strings.TrimSpace(profile.ID) == "" {
		profile.ID = newPrefixedID("payout_profile")
	}
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now
	_, _ = s.payoutProf.ReplaceOne(ctx, bson.M{"tenantId": profile.TenantID}, profile, options.Replace().SetUpsert(true))
	return profile
}

func (s *MongoStore) GetSellerPayoutProfile(tenantID string) (models.SellerPayoutProfile, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var profile models.SellerPayoutProfile
	err := s.payoutProf.FindOne(ctx, bson.M{"tenantId": strings.TrimSpace(tenantID)}).Decode(&profile)
	if err != nil {
		return models.SellerPayoutProfile{}, false
	}
	return profile, true
}

func (s *MongoStore) ListSellerPayoutProfiles() []models.SellerPayoutProfile {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.payoutProf.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}))
	if err != nil {
		return []models.SellerPayoutProfile{}
	}
	items, err := decodeAll[models.SellerPayoutProfile](cur)
	if err != nil {
		return []models.SellerPayoutProfile{}
	}
	return items
}

func (s *MongoStore) CreatePayoutRecord(record models.PayoutRecord) models.PayoutRecord {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	now := time.Now().UTC()
	if strings.TrimSpace(record.ID) == "" {
		record.ID = newPrefixedID("payout")
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	record.UpdatedAt = now
	_, _ = s.payouts.InsertOne(ctx, record)
	return record
}

func (s *MongoStore) UpdatePayoutRecord(record models.PayoutRecord) bool {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	record.UpdatedAt = time.Now().UTC()
	res, err := s.payouts.ReplaceOne(ctx, bson.M{"id": record.ID}, record)
	if err != nil {
		return false
	}
	return res.MatchedCount > 0
}

func (s *MongoStore) ListPayoutRecords(tenantID string, limit int) []models.PayoutRecord {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	filter := bson.M{}
	if strings.TrimSpace(tenantID) != "" {
		filter["tenantId"] = strings.TrimSpace(tenantID)
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cur, err := s.payouts.Find(ctx, filter, opts)
	if err != nil {
		return []models.PayoutRecord{}
	}
	items, err := decodeAll[models.PayoutRecord](cur)
	if err != nil {
		return []models.PayoutRecord{}
	}
	return items
}

func (s *MongoStore) ListLocalAgents(tenantID, userID string) []models.LocalAgent {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := s.agents.Find(ctx, bson.M{"tenantId": tenantID, "userId": userID})
	if err != nil {
		return []models.LocalAgent{}
	}
	items, err := decodeAll[models.LocalAgent](cur)
	if err != nil {
		return []models.LocalAgent{}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].LastSeenAt.After(items[j].LastSeenAt)
	})
	return items
}

func (s *MongoStore) UpsertLocalAgent(agent models.LocalAgent) models.LocalAgent {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	agent.LastSeenAt = time.Now().UTC()
	filter := bson.M{}
	if strings.TrimSpace(agent.ID) != "" {
		filter["id"] = agent.ID
	} else {
		agent.ID = newPrefixedID("agent")
		filter["tenantId"] = agent.TenantID
		filter["userId"] = agent.UserID
		filter["deviceId"] = agent.DeviceID
	}
	update := bson.M{
		"$set": bson.M{
			"id":           agent.ID,
			"tenantId":     agent.TenantID,
			"userId":       agent.UserID,
			"deviceId":     agent.DeviceID,
			"version":      agent.Version,
			"tunnelStatus": agent.TunnelStatus,
			"lastSeenAt":   agent.LastSeenAt,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var out models.LocalAgent
	if err := s.agents.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out); err != nil {
		return agent
	}
	return out
}

func (s *MongoStore) GetUserSettings(userID string) (models.UserSettings, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var settings models.UserSettings
	err := s.userSettings.FindOne(ctx, bson.M{"userId": userID}).Decode(&settings)
	if err != nil {
		return models.UserSettings{}, false
	}
	return settings, true
}

func (s *MongoStore) UpsertUserSettings(settings models.UserSettings) models.UserSettings {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	settings.UpdatedAt = time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"userId":        settings.UserID,
			"preferences":   settings.Preferences,
			"notifications": settings.Notifications,
			"updatedAt":     settings.UpdatedAt,
		},
	}
	_, _ = s.userSettings.UpdateOne(ctx, bson.M{"userId": settings.UserID}, update, options.Update().SetUpsert(true))
	return settings
}

// OAuthClient methods
func (s *MongoStore) CreateOAuthClient(client models.OAuthClient) models.OAuthClient {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	now := time.Now().UTC()
	client.ID = newPrefixedID("oauth_client")
	client.CreatedAt = now
	client.UpdatedAt = now
	if client.ClientIDIssuedAt.IsZero() {
		client.ClientIDIssuedAt = now
	}
	_, _ = s.oauthClients.InsertOne(ctx, client)
	return client
}

func (s *MongoStore) GetOAuthClient(clientID string) (models.OAuthClient, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var client models.OAuthClient
	err := s.oauthClients.FindOne(ctx, bson.M{"clientId": clientID}).Decode(&client)
	if err != nil {
		return models.OAuthClient{}, false
	}
	return client, true
}

func (s *MongoStore) ListOAuthClients() []models.OAuthClient {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cur, err := s.oauthClients.Find(ctx, bson.M{}, opts)
	if err != nil {
		return []models.OAuthClient{}
	}
	clients, err := decodeAll[models.OAuthClient](cur)
	if err != nil {
		return []models.OAuthClient{}
	}
	return clients
}

func (s *MongoStore) DeleteOAuthClient(clientID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	res, err := s.oauthClients.DeleteOne(ctx, bson.M{"clientId": clientID})
	if err != nil {
		return false
	}
	return res.DeletedCount > 0
}

// OAuthAuthCode methods (with TTL)
func (s *MongoStore) CreateOAuthAuthCode(code models.OAuthAuthCode) models.OAuthAuthCode {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	now := time.Now().UTC()
	code.ID = newPrefixedID("oauth_code")
	code.CreatedAt = now
	if code.ExpiresAt.IsZero() {
		code.ExpiresAt = now.Add(5 * time.Minute)
	}
	code.Consumed = false
	_, _ = s.oauthCodes.InsertOne(ctx, code)
	return code
}

func (s *MongoStore) GetOAuthAuthCode(code string) (models.OAuthAuthCode, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var ac models.OAuthAuthCode
	err := s.oauthCodes.FindOne(ctx, bson.M{"code": code}).Decode(&ac)
	if err != nil {
		return models.OAuthAuthCode{}, false
	}
	// Check if expired or consumed
	if ac.Consumed || time.Now().UTC().After(ac.ExpiresAt) {
		return models.OAuthAuthCode{}, false
	}
	return ac, true
}

func (s *MongoStore) ConsumeOAuthAuthCode(code string) (models.OAuthAuthCode, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	// First, find and atomically mark as consumed
	filter := bson.M{
		"code":      code,
		"consumed":  false,
		"expiresAt": bson.M{"$gt": time.Now().UTC()},
	}
	update := bson.M{
		"$set": bson.M{"consumed": true},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var ac models.OAuthAuthCode
	err := s.oauthCodes.FindOneAndUpdate(ctx, filter, update, opts).Decode(&ac)
	if err != nil {
		return models.OAuthAuthCode{}, false
	}
	return ac, true
}

func (s *MongoStore) StoreType() string {
	return "mongo"
}

func (s *MongoStore) Health(ctx context.Context) error {
	return s.client.Ping(ctx, nil)
}
