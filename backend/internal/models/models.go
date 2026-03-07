package models

import "time"

type Role string

const (
	RoleBuyer    Role = "buyer"
	RoleMerchant Role = "merchant"
	RoleAdmin    Role = "admin"
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
	Metadata           map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt" bson:"updatedAt"`
	FulfilledAt        time.Time              `json:"fulfilledAt,omitempty" bson:"fulfilledAt,omitempty"`
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
