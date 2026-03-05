package models

import "time"

type Role string

const (
	RoleBuyer    Role = "buyer"
	RoleMerchant Role = "merchant"
	RoleAdmin    Role = "admin"
)

type User struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Role         Role      `json:"role"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Tenant struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	OwnerUserID string    `json:"ownerUserId"`
	PlanTier    string    `json:"planTier"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Server struct {
	ID                   string    `json:"id"`
	TenantID             string    `json:"tenantId"`
	Author               string    `json:"author"`
	Name                 string    `json:"name"`
	Slug                 string    `json:"slug"`
	Description          string    `json:"description"`
	Category             string    `json:"category"`
	Version              string    `json:"version"`
	DockerImage          string    `json:"dockerImage"`
	CanonicalResourceURI string    `json:"canonicalResourceUri"`
	RequiredScopes       []string  `json:"requiredScopes"`
	PricingType          string    `json:"pricingType"`
	PricingAmount        float64   `json:"pricingAmount"`
	Verified             bool      `json:"verified"`
	Featured             bool      `json:"featured"`
	InstallCount         int       `json:"installCount"`
	Rating               float64   `json:"rating"`
	Status               string    `json:"status"`
	SupportsLocal        bool      `json:"supportsLocal"`
	SupportsCloud        bool      `json:"supportsCloud"`
	UpdatedAt            time.Time `json:"updatedAt"`
	CreatedAt            time.Time `json:"createdAt"`
}

type Entitlement struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenantId"`
	UserID        string    `json:"userId"`
	ServerID      string    `json:"serverId"`
	AllowedScopes []string  `json:"allowedScopes"`
	CloudAllowed  bool      `json:"cloudAllowed"`
	LocalAllowed  bool      `json:"localAllowed"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Connection struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenantId"`
	UserID         string    `json:"userId"`
	HubID          string    `json:"hubId"`
	Client         string    `json:"client"`
	Status         string    `json:"status"`
	Resource       string    `json:"resource"`
	GrantedScopes  []string  `json:"grantedScopes"`
	TokenExpiresAt time.Time `json:"tokenExpiresAt"`
	LastUsedAt     time.Time `json:"lastUsedAt"`
	CreatedAt      time.Time `json:"createdAt"`
	CatalogVersion int       `json:"catalogVersion"`
}

type HubProfile struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenantId"`
	UserID         string    `json:"userId"`
	HubURL         string    `json:"hubUrl"`
	CatalogHash    string    `json:"catalogHash"`
	CatalogVersion int       `json:"catalogVersion"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type HubRoute struct {
	ID           string    `json:"id"`
	HubID        string    `json:"hubId"`
	ServerID     string    `json:"serverId"`
	ToolName     string    `json:"toolName"`
	UpstreamType string    `json:"upstreamType"`
	Priority     int       `json:"priority"`
	Enabled      bool      `json:"enabled"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type SecurityEvent struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Actor       string    `json:"actor"`
	TargetID    string    `json:"targetId"`
	Resolved    bool      `json:"resolved"`
	CreatedAt   time.Time `json:"createdAt"`
}

type AuditLog struct {
	ID         string                 `json:"id"`
	TenantID   string                 `json:"tenantId"`
	ActorID    string                 `json:"actorId"`
	Action     string                 `json:"action"`
	TargetType string                 `json:"targetType"`
	TargetID   string                 `json:"targetId"`
	Outcome    string                 `json:"outcome"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"createdAt"`
}

type X402Intent struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenantId"`
	UserID     string    `json:"userId"`
	ServerID   string    `json:"serverId"`
	ToolName   string    `json:"toolName"`
	AmountUSDC float64   `json:"amountUsdc"`
	Network    string    `json:"network"`
	Asset      string    `json:"asset"`
	Status     string    `json:"status"`
	Challenge  string    `json:"challenge"`
	CreatedAt  time.Time `json:"createdAt"`
	SettledAt  time.Time `json:"settledAt,omitempty"`
}

type LocalAgent struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	UserID       string    `json:"userId"`
	DeviceID     string    `json:"deviceId"`
	Version      string    `json:"version"`
	TunnelStatus string    `json:"tunnelStatus"`
	LastSeenAt   time.Time `json:"lastSeenAt"`
}
