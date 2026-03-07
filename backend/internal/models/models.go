package models

import "time"

type Role string

const (
	RoleBuyer    Role = "buyer"
	RoleMerchant Role = "merchant"
	RoleAdmin    Role = "admin"
)

const (
	ServerStatusDraft     = "draft"
	ServerStatusPublished = "published"
	ServerStatusArchived  = "archived"
)

const (
	ServerDeploymentPending  = "not_deployed"
	ServerDeploymentQueued   = "deploy_queued"
	ServerDeploymentDeployed = "deployed"
)

const (
	DeployTaskStatusPending    = "pending"
	DeployTaskStatusProcessing = "processing"
	DeployTaskStatusCompleted  = "completed"
	DeployTaskStatusFailed     = "failed"
)

type User struct {
	ID            string    `json:"id" bson:"id"`
	TenantID      string    `json:"tenantId" bson:"tenantId"`
	Email         string    `json:"email" bson:"email"`
	Name          string    `json:"name" bson:"name"`
	Phone         string    `json:"phone,omitempty" bson:"phone,omitempty"`
	AvatarURL     string    `json:"avatarUrl,omitempty" bson:"avatarUrl,omitempty"`
	Locale        string    `json:"locale,omitempty" bson:"locale,omitempty"`
	Timezone      string    `json:"timezone,omitempty" bson:"timezone,omitempty"`
	Role          Role      `json:"role" bson:"role"`
	PasswordHash  string    `json:"-" bson:"passwordHash"`
	MFAEnabled    bool      `json:"mfaEnabled,omitempty" bson:"mfaEnabled,omitempty"`
	MFATOTPSecret string    `json:"-" bson:"mfaTotpSecret,omitempty"`
	CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" bson:"updatedAt"`
}

type Tenant struct {
	ID          string    `json:"id" bson:"id"`
	Name        string    `json:"name" bson:"name"`
	Slug        string    `json:"slug" bson:"slug"`
	OwnerUserID string    `json:"ownerUserId" bson:"ownerUserId"`
	PlanTier    string    `json:"planTier" bson:"planTier"`
	Status      string    `json:"status" bson:"status"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}

type Server struct {
	ID                   string    `json:"id" bson:"id"`
	TenantID             string    `json:"tenantId" bson:"tenantId"`
	Author               string    `json:"author" bson:"author"`
	Name                 string    `json:"name" bson:"name"`
	Slug                 string    `json:"slug" bson:"slug"`
	Description          string    `json:"description" bson:"description"`
	Category             string    `json:"category" bson:"category"`
	Version              string    `json:"version" bson:"version"`
	DockerImage          string    `json:"dockerImage" bson:"dockerImage"`
	CanonicalResourceURI string    `json:"canonicalResourceUri" bson:"canonicalResourceUri"`
	RequiredScopes       []string  `json:"requiredScopes" bson:"requiredScopes"`
	PricingType          string    `json:"pricingType" bson:"pricingType"`
	PricingAmount        float64   `json:"pricingAmount" bson:"pricingAmount"`
	Verified             bool      `json:"verified" bson:"verified"`
	Featured             bool      `json:"featured" bson:"featured"`
	InstallCount         int       `json:"installCount" bson:"installCount"`
	Rating               float64   `json:"rating" bson:"rating"`
	Status               string    `json:"status" bson:"status"`
	DeploymentStatus     string    `json:"deploymentStatus,omitempty" bson:"deploymentStatus,omitempty"`
	DeploymentTarget     string    `json:"deploymentTarget,omitempty" bson:"deploymentTarget,omitempty"`
	DeployedBy           string    `json:"deployedBy,omitempty" bson:"deployedBy,omitempty"`
	DeployedAt           time.Time `json:"deployedAt,omitempty" bson:"deployedAt,omitempty"`
	N8nWorkflowID        string    `json:"n8nWorkflowId,omitempty" bson:"n8nWorkflowId,omitempty"`
	N8nWorkflowURL       string    `json:"n8nWorkflowUrl,omitempty" bson:"n8nWorkflowUrl,omitempty"`
	PublishedAt          time.Time `json:"publishedAt,omitempty" bson:"publishedAt,omitempty"`
	SupportsLocal        bool      `json:"supportsLocal" bson:"supportsLocal"`
	SupportsCloud        bool      `json:"supportsCloud" bson:"supportsCloud"`
	PaymentMethods       []string  `json:"paymentMethods,omitempty" bson:"paymentMethods,omitempty"`
	PaymentAddress       string    `json:"paymentAddress,omitempty" bson:"paymentAddress,omitempty"`
	PerCallCapUSDC       float64   `json:"perCallCapUsdc,omitempty" bson:"perCallCapUsdc,omitempty"`
	DailyCapUSDC         float64   `json:"dailyCapUsdc,omitempty" bson:"dailyCapUsdc,omitempty"`
	MonthlyCapUSDC       float64   `json:"monthlyCapUsdc,omitempty" bson:"monthlyCapUsdc,omitempty"`
	UpdatedAt            time.Time `json:"updatedAt" bson:"updatedAt"`
	CreatedAt            time.Time `json:"createdAt" bson:"createdAt"`
}

type DeployTask struct {
	ID                 string    `json:"id" bson:"id"`
	TenantID           string    `json:"tenantId" bson:"tenantId"`
	ServerID           string    `json:"serverId" bson:"serverId"`
	RequestedBy        string    `json:"requestedBy" bson:"requestedBy"`
	PreferredWorkflowID string   `json:"preferredWorkflowId,omitempty" bson:"preferredWorkflowId,omitempty"`
	DeploymentTarget   string    `json:"deploymentTarget,omitempty" bson:"deploymentTarget,omitempty"`
	Status             string    `json:"status" bson:"status"`
	AttemptCount       int       `json:"attemptCount" bson:"attemptCount"`
	MaxAttempts        int       `json:"maxAttempts" bson:"maxAttempts"`
	NextAttemptAt      time.Time `json:"nextAttemptAt" bson:"nextAttemptAt"`
	LastError          string    `json:"lastError,omitempty" bson:"lastError,omitempty"`
	CreatedAt          time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt" bson:"updatedAt"`
	CompletedAt        time.Time `json:"completedAt,omitempty" bson:"completedAt,omitempty"`
}

type Entitlement struct {
	ID            string    `json:"id" bson:"id"`
	TenantID      string    `json:"tenantId" bson:"tenantId"`
	UserID        string    `json:"userId" bson:"userId"`
	ServerID      string    `json:"serverId" bson:"serverId"`
	AllowedScopes []string  `json:"allowedScopes" bson:"allowedScopes"`
	CloudAllowed  bool      `json:"cloudAllowed" bson:"cloudAllowed"`
	LocalAllowed  bool      `json:"localAllowed" bson:"localAllowed"`
	Status        string    `json:"status" bson:"status"`
	CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" bson:"updatedAt"`
}

type Connection struct {
	ID             string    `json:"id" bson:"id"`
	TenantID       string    `json:"tenantId" bson:"tenantId"`
	UserID         string    `json:"userId" bson:"userId"`
	ServerID       string    `json:"serverId,omitempty" bson:"serverId,omitempty"`
	ServerSlug     string    `json:"serverSlug,omitempty" bson:"serverSlug,omitempty"`
	ServerName     string    `json:"serverName,omitempty" bson:"serverName,omitempty"`
	HubID          string    `json:"hubId" bson:"hubId"`
	Client         string    `json:"client" bson:"client"`
	Status         string    `json:"status" bson:"status"`
	Resource       string    `json:"resource" bson:"resource"`
	GrantedScopes  []string  `json:"grantedScopes" bson:"grantedScopes"`
	TokenExpiresAt time.Time `json:"tokenExpiresAt" bson:"tokenExpiresAt"`
	LastUsedAt     time.Time `json:"lastUsedAt" bson:"lastUsedAt"`
	CreatedAt      time.Time `json:"createdAt" bson:"createdAt"`
	CatalogVersion int       `json:"catalogVersion" bson:"catalogVersion"`
}

type HubProfile struct {
	ID             string    `json:"id" bson:"id"`
	TenantID       string    `json:"tenantId" bson:"tenantId"`
	UserID         string    `json:"userId" bson:"userId"`
	HubURL         string    `json:"hubUrl" bson:"hubUrl"`
	CatalogHash    string    `json:"catalogHash" bson:"catalogHash"`
	CatalogVersion int       `json:"catalogVersion" bson:"catalogVersion"`
	Status         string    `json:"status" bson:"status"`
	CreatedAt      time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt" bson:"updatedAt"`
}

type HubRoute struct {
	ID           string    `json:"id" bson:"id"`
	HubID        string    `json:"hubId" bson:"hubId"`
	ServerID     string    `json:"serverId" bson:"serverId"`
	ToolName     string    `json:"toolName" bson:"toolName"`
	UpstreamType string    `json:"upstreamType" bson:"upstreamType"`
	Priority     int       `json:"priority" bson:"priority"`
	Enabled      bool      `json:"enabled" bson:"enabled"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
}

type SecurityEvent struct {
	ID          string    `json:"id" bson:"id"`
	TenantID    string    `json:"tenantId" bson:"tenantId"`
	Type        string    `json:"type" bson:"type"`
	Severity    string    `json:"severity" bson:"severity"`
	Description string    `json:"description" bson:"description"`
	Actor       string    `json:"actor" bson:"actor"`
	TargetID    string    `json:"targetId" bson:"targetId"`
	Resolved    bool      `json:"resolved" bson:"resolved"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}

type AuditLog struct {
	ID         string                 `json:"id" bson:"id"`
	TenantID   string                 `json:"tenantId" bson:"tenantId"`
	ActorID    string                 `json:"actorId" bson:"actorId"`
	Action     string                 `json:"action" bson:"action"`
	TargetType string                 `json:"targetType" bson:"targetType"`
	TargetID   string                 `json:"targetId" bson:"targetId"`
	Outcome    string                 `json:"outcome" bson:"outcome"`
	Metadata   map[string]interface{} `json:"metadata" bson:"metadata"`
	CreatedAt  time.Time              `json:"createdAt" bson:"createdAt"`
}

type X402Intent struct {
	ID                 string    `json:"id" bson:"id"`
	TenantID           string    `json:"tenantId" bson:"tenantId"`
	UserID             string    `json:"userId" bson:"userId"`
	ServerID           string    `json:"serverId" bson:"serverId"`
	ToolName           string    `json:"toolName" bson:"toolName"`
	AmountUSDC         float64   `json:"amountUsdc" bson:"amountUsdc"`
	Network            string    `json:"network" bson:"network"`
	Asset              string    `json:"asset" bson:"asset"`
	Status             string    `json:"status" bson:"status"`
	Challenge          string    `json:"challenge" bson:"challenge"`
	PaymentMethod      string    `json:"paymentMethod,omitempty" bson:"paymentMethod,omitempty"`
	PaymentIdentifier  string    `json:"paymentIdentifier,omitempty" bson:"paymentIdentifier,omitempty"`
	IdempotencyKey     string    `json:"idempotencyKey,omitempty" bson:"idempotencyKey,omitempty"`
	VerificationStatus string    `json:"verificationStatus,omitempty" bson:"verificationStatus,omitempty"`
	VerificationNote   string    `json:"verificationNote,omitempty" bson:"verificationNote,omitempty"`
	FacilitatorTx      string    `json:"facilitatorTx,omitempty" bson:"facilitatorTx,omitempty"`
	Quantity           int       `json:"quantity,omitempty" bson:"quantity,omitempty"`
	RemainingQuantity  int       `json:"remainingQuantity,omitempty" bson:"remainingQuantity,omitempty"`
	Resource           string    `json:"resource,omitempty" bson:"resource,omitempty"`
	X402Version        string    `json:"x402Version,omitempty" bson:"x402Version,omitempty"`
	RequestFingerprint string    `json:"requestFingerprint,omitempty" bson:"requestFingerprint,omitempty"`
	PlatformFeeBps     int       `json:"platformFeeBps,omitempty" bson:"platformFeeBps,omitempty"`
	PlatformFeeUSDC    float64   `json:"platformFeeUsdc,omitempty" bson:"platformFeeUsdc,omitempty"`
	SellerNetUSDC      float64   `json:"sellerNetUsdc,omitempty" bson:"sellerNetUsdc,omitempty"`
	AccountingPosted   bool      `json:"accountingPosted,omitempty" bson:"accountingPosted,omitempty"`
	AccountingRef      string    `json:"accountingRef,omitempty" bson:"accountingRef,omitempty"`
	CreatedAt          time.Time `json:"createdAt" bson:"createdAt"`
	SettledAt          time.Time `json:"settledAt,omitempty" bson:"settledAt,omitempty"`
}

type PaymentPolicy struct {
	TenantID            string    `json:"tenantId" bson:"tenantId"`
	UserID              string    `json:"userId" bson:"userId"`
	MonthlySpendCapUSDC float64   `json:"monthlySpendCapUsdc,omitempty" bson:"monthlySpendCapUsdc,omitempty"`
	DailySpendCapUSDC   float64   `json:"dailySpendCapUsdc,omitempty" bson:"dailySpendCapUsdc,omitempty"`
	PerCallCapUSDC      float64   `json:"perCallCapUsdc,omitempty" bson:"perCallCapUsdc,omitempty"`
	AllowedMethods      []string  `json:"allowedMethods,omitempty" bson:"allowedMethods,omitempty"`
	SIWXWallet          string    `json:"siwxWallet,omitempty" bson:"siwxWallet,omitempty"`
	WalletBalanceUSDC   float64   `json:"walletBalanceUsdc,omitempty" bson:"walletBalanceUsdc,omitempty"`
	MinimumBalanceUSDC  float64   `json:"minimumBalanceUsdc,omitempty" bson:"minimumBalanceUsdc,omitempty"`
	HardStopOnLowFunds  bool      `json:"hardStopOnLowFunds,omitempty" bson:"hardStopOnLowFunds,omitempty"`
	AutoTopUpEnabled    bool      `json:"autoTopUpEnabled,omitempty" bson:"autoTopUpEnabled,omitempty"`
	AutoTopUpAmountUSD  float64   `json:"autoTopUpAmountUsd,omitempty" bson:"autoTopUpAmountUsd,omitempty"`
	AutoTopUpTriggerUSD float64   `json:"autoTopUpTriggerUsd,omitempty" bson:"autoTopUpTriggerUsd,omitempty"`
	FundingMethod       string    `json:"fundingMethod,omitempty" bson:"fundingMethod,omitempty"`
	WalletAddress       string    `json:"walletAddress,omitempty" bson:"walletAddress,omitempty"`
	LastTopUpAt         time.Time `json:"lastTopUpAt,omitempty" bson:"lastTopUpAt,omitempty"`
	UpdatedAt           time.Time `json:"updatedAt" bson:"updatedAt"`
}

type WalletTopUp struct {
	ID                 string                 `json:"id" bson:"id"`
	TenantID           string                 `json:"tenantId" bson:"tenantId"`
	UserID             string                 `json:"userId" bson:"userId"`
	Provider           string                 `json:"provider" bson:"provider"`
	ProviderSessionID  string                 `json:"providerSessionId,omitempty" bson:"providerSessionId,omitempty"`
	ProviderEventID    string                 `json:"providerEventId,omitempty" bson:"providerEventId,omitempty"`
	Status             string                 `json:"status" bson:"status"`
	SourceCurrency     string                 `json:"sourceCurrency,omitempty" bson:"sourceCurrency,omitempty"`
	SourceAmount       float64                `json:"sourceAmount,omitempty" bson:"sourceAmount,omitempty"`
	DestinationAsset   string                 `json:"destinationAsset,omitempty" bson:"destinationAsset,omitempty"`
	DestinationNetwork string                 `json:"destinationNetwork,omitempty" bson:"destinationNetwork,omitempty"`
	DestinationAmount  float64                `json:"destinationAmount,omitempty" bson:"destinationAmount,omitempty"`
	PaymentMethod      string                 `json:"paymentMethod,omitempty" bson:"paymentMethod,omitempty"`
	WalletAddress      string                 `json:"walletAddress,omitempty" bson:"walletAddress,omitempty"`
	HostedURL          string                 `json:"hostedUrl,omitempty" bson:"hostedUrl,omitempty"`
	FailureReason      string                 `json:"failureReason,omitempty" bson:"failureReason,omitempty"`
	EstimatedFeeUSD    float64                `json:"estimatedFeeUsd,omitempty" bson:"estimatedFeeUsd,omitempty"`
	AccountingPosted   bool                   `json:"accountingPosted,omitempty" bson:"accountingPosted,omitempty"`
	AccountingRef      string                 `json:"accountingRef,omitempty" bson:"accountingRef,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt" bson:"updatedAt"`
	FulfilledAt        time.Time              `json:"fulfilledAt,omitempty" bson:"fulfilledAt,omitempty"`
}

type PaymentFeePolicy struct {
	ID             string    `json:"id" bson:"id"`
	Scope          string    `json:"scope" bson:"scope"` // global, tenant, server
	TenantID       string    `json:"tenantId,omitempty" bson:"tenantId,omitempty"`
	ServerID       string    `json:"serverId,omitempty" bson:"serverId,omitempty"`
	PlatformFeeBps int       `json:"platformFeeBps" bson:"platformFeeBps"`
	MinFeeUSDC     float64   `json:"minFeeUsdc,omitempty" bson:"minFeeUsdc,omitempty"`
	MaxFeeUSDC     float64   `json:"maxFeeUsdc,omitempty" bson:"maxFeeUsdc,omitempty"`
	HoldDays       int       `json:"holdDays,omitempty" bson:"holdDays,omitempty"`
	AutoPayouts    bool      `json:"autoPayouts,omitempty" bson:"autoPayouts,omitempty"`
	PayoutCadence  string    `json:"payoutCadence,omitempty" bson:"payoutCadence,omitempty"` // daily, weekly, manual
	Enabled        bool      `json:"enabled" bson:"enabled"`
	CreatedBy      string    `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	UpdatedAt      time.Time `json:"updatedAt" bson:"updatedAt"`
	CreatedAt      time.Time `json:"createdAt" bson:"createdAt"`
}

type LedgerEntry struct {
	ID            string                 `json:"id" bson:"id"`
	TransactionID string                 `json:"transactionId" bson:"transactionId"`
	TenantID      string                 `json:"tenantId,omitempty" bson:"tenantId,omitempty"`
	UserID        string                 `json:"userId,omitempty" bson:"userId,omitempty"`
	ServerID      string                 `json:"serverId,omitempty" bson:"serverId,omitempty"`
	IntentID      string                 `json:"intentId,omitempty" bson:"intentId,omitempty"`
	TopUpID       string                 `json:"topupId,omitempty" bson:"topupId,omitempty"`
	PayoutID      string                 `json:"payoutId,omitempty" bson:"payoutId,omitempty"`
	Account       string                 `json:"account" bson:"account"`
	EntryType     string                 `json:"entryType" bson:"entryType"` // debit, credit
	AmountUSDC    float64                `json:"amountUsdc" bson:"amountUsdc"`
	Category      string                 `json:"category,omitempty" bson:"category,omitempty"`
	Description   string                 `json:"description,omitempty" bson:"description,omitempty"`
	Reference     string                 `json:"reference,omitempty" bson:"reference,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"createdAt" bson:"createdAt"`
}

type SellerPayoutProfile struct {
	ID                  string    `json:"id" bson:"id"`
	TenantID            string    `json:"tenantId" bson:"tenantId"`
	PreferredMethod     string    `json:"preferredMethod" bson:"preferredMethod"` // stablecoin, stripe_connect
	StablecoinAddress   string    `json:"stablecoinAddress,omitempty" bson:"stablecoinAddress,omitempty"`
	StablecoinNetwork   string    `json:"stablecoinNetwork,omitempty" bson:"stablecoinNetwork,omitempty"`
	StripeAccountID     string    `json:"stripeAccountId,omitempty" bson:"stripeAccountId,omitempty"`
	StripeOnboardingURL string    `json:"stripeOnboardingUrl,omitempty" bson:"stripeOnboardingUrl,omitempty"`
	KYCStatus           string    `json:"kycStatus,omitempty" bson:"kycStatus,omitempty"` // pending, verified, restricted
	KYCDetailsSubmitted bool      `json:"kycDetailsSubmitted,omitempty" bson:"kycDetailsSubmitted,omitempty"`
	KYCChargesEnabled   bool      `json:"kycChargesEnabled,omitempty" bson:"kycChargesEnabled,omitempty"`
	KYCPayoutsEnabled   bool      `json:"kycPayoutsEnabled,omitempty" bson:"kycPayoutsEnabled,omitempty"`
	KYCCurrentlyDue     []string  `json:"kycCurrentlyDue,omitempty" bson:"kycCurrentlyDue,omitempty"`
	KYCEventuallyDue    []string  `json:"kycEventuallyDue,omitempty" bson:"kycEventuallyDue,omitempty"`
	KYCDisabledReason   string    `json:"kycDisabledReason,omitempty" bson:"kycDisabledReason,omitempty"`
	KYCLastCheckedAt    time.Time `json:"kycLastCheckedAt,omitempty" bson:"kycLastCheckedAt,omitempty"`
	PayoutBlocked       bool      `json:"payoutBlocked,omitempty" bson:"payoutBlocked,omitempty"`
	PayoutBlockReason   string    `json:"payoutBlockReason,omitempty" bson:"payoutBlockReason,omitempty"`
	TaxFormStatus       string    `json:"taxFormStatus,omitempty" bson:"taxFormStatus,omitempty"` // missing, submitted, verified
	RiskLevel           string    `json:"riskLevel,omitempty" bson:"riskLevel,omitempty"`         // low, medium, high
	MinPayoutUSDC       float64   `json:"minPayoutUsdc,omitempty" bson:"minPayoutUsdc,omitempty"`
	HoldDays            int       `json:"holdDays,omitempty" bson:"holdDays,omitempty"`
	UpdatedAt           time.Time `json:"updatedAt" bson:"updatedAt"`
	CreatedAt           time.Time `json:"createdAt" bson:"createdAt"`
}

type PayoutRecord struct {
	ID                string                 `json:"id" bson:"id"`
	TenantID          string                 `json:"tenantId" bson:"tenantId"`
	Method            string                 `json:"method" bson:"method"` // stablecoin, stripe_connect
	Status            string                 `json:"status" bson:"status"` // pending, processing, completed, failed, blocked
	AmountUSDC        float64                `json:"amountUsdc" bson:"amountUsdc"`
	FeeUSDC           float64                `json:"feeUsdc,omitempty" bson:"feeUsdc,omitempty"`
	NetUSDC           float64                `json:"netUsdc,omitempty" bson:"netUsdc,omitempty"`
	StablecoinAddress string                 `json:"stablecoinAddress,omitempty" bson:"stablecoinAddress,omitempty"`
	StripeAccountID   string                 `json:"stripeAccountId,omitempty" bson:"stripeAccountId,omitempty"`
	ExternalRef       string                 `json:"externalRef,omitempty" bson:"externalRef,omitempty"`
	FailureReason     string                 `json:"failureReason,omitempty" bson:"failureReason,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"createdAt" bson:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt" bson:"updatedAt"`
	CompletedAt       time.Time              `json:"completedAt,omitempty" bson:"completedAt,omitempty"`
}

type LocalAgent struct {
	ID           string    `json:"id" bson:"id"`
	TenantID     string    `json:"tenantId" bson:"tenantId"`
	UserID       string    `json:"userId" bson:"userId"`
	DeviceID     string    `json:"deviceId" bson:"deviceId"`
	Version      string    `json:"version" bson:"version"`
	TunnelStatus string    `json:"tunnelStatus" bson:"tunnelStatus"`
	LastSeenAt   time.Time `json:"lastSeenAt" bson:"lastSeenAt"`
}

type UserPreferences struct {
	Theme          string `json:"theme" bson:"theme"`
	Language       string `json:"language" bson:"language"`
	Timezone       string `json:"timezone" bson:"timezone"`
	DefaultLanding string `json:"defaultLanding" bson:"defaultLanding"`
	CompactMode    bool   `json:"compactMode" bson:"compactMode"`
}

type NotificationSettings struct {
	ProductUpdates bool `json:"productUpdates" bson:"productUpdates"`
	SecurityAlerts bool `json:"securityAlerts" bson:"securityAlerts"`
	BillingAlerts  bool `json:"billingAlerts" bson:"billingAlerts"`
	MarketingEmail bool `json:"marketingEmail" bson:"marketingEmail"`
	WeeklyDigest   bool `json:"weeklyDigest" bson:"weeklyDigest"`
}

type UserSettings struct {
	UserID        string               `json:"userId" bson:"userId"`
	Preferences   UserPreferences      `json:"preferences" bson:"preferences"`
	Notifications NotificationSettings `json:"notifications" bson:"notifications"`
	UpdatedAt     time.Time            `json:"updatedAt" bson:"updatedAt"`
}
