package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yourorg/mcp-marketplace/backend/internal/auth"
	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/store"
)

type App struct {
	cfg            config.Config
	store          store.Store
	jwt            *auth.JWTManager
	oauth          *oauthState
	x402           *x402Service
	stripeOnramp   *stripeOnrampService
	allowedOrigins map[string]struct{}
	rateLimiter    *ipRateLimiter
	authLimiter    *ipRateLimiter
}

func NewRouter(cfg config.Config, st store.Store, jwt *auth.JWTManager) http.Handler {
	allowedOrigins := map[string]struct{}{}
	for _, origin := range cfg.CORSAllowedOrigins {
		allowedOrigins[origin] = struct{}{}
	}
	app := &App{
		cfg:            cfg,
		store:          st,
		jwt:            jwt,
		oauth:          newOAuthState(),
		x402:           newX402Service(cfg),
		stripeOnramp:   newStripeOnrampService(cfg),
		allowedOrigins: allowedOrigins,
		rateLimiter:    newIPRateLimiter(cfg.RateLimitPerMinute, cfg.RateLimitPerMinute/4),
		authLimiter:    newIPRateLimiter(30, 10),
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(app.rateLimit)
	r.Use(app.cors)
	r.Use(middleware.Heartbeat("/healthz"))
	r.Use(app.securityHeaders)

	r.Get("/health", app.health)
	r.Post("/auth/signup", app.signup)
	r.Post("/auth/login", app.login)
	r.Get("/.well-known/oauth-protected-resource", app.oauthProtectedResourceMetadata)
	r.Get("/.well-known/oauth-authorization-server", app.oauthAuthorizationServerMetadata)
	r.Post("/webhooks/stripe/onramp", app.handleStripeOnrampWebhook)
	r.Post("/oauth/register", app.oauthRegisterClient)
	r.With(app.authenticate).Get("/oauth/authorize", app.oauthAuthorize)
	r.Post("/oauth/token", app.oauthToken)
	r.With(app.authenticate).Get("/mcp/hub/{tenantID}/{userID}", app.mcpHub)
	r.With(app.authenticate).Post("/mcp/hub/{tenantID}/{userID}", app.mcpHub)

	r.Route("/v1", func(v1 chi.Router) {
		v1.Get("/marketplace/servers", app.listMarketplaceServers)
		v1.Get("/marketplace/servers/{slug}", app.getMarketplaceServer)

		v1.Group(func(prv chi.Router) {
			prv.Use(app.authenticate)
			prv.Get("/me", app.me)
			prv.Post("/marketplace/servers/{slug}/install", app.installMarketplaceServer)
			prv.Get("/settings/profile", app.getUserProfile)
			prv.Put("/settings/profile", app.updateUserProfile)
			prv.Put("/settings/security/password", app.changeUserPassword)
			prv.Get("/settings/security/mfa", app.getMFAStatus)
			prv.Post("/settings/security/mfa/totp/setup", app.setupTOTP)
			prv.Post("/settings/security/mfa/totp/verify", app.verifyTOTP)
			prv.Post("/settings/security/mfa/totp/disable", app.disableTOTP)
			prv.Get("/settings/preferences", app.getUserPreferences)
			prv.Put("/settings/preferences", app.updateUserPreferences)
			prv.Get("/settings/notifications", app.getUserNotifications)
			prv.Put("/settings/notifications", app.updateUserNotifications)
			prv.Get("/buyer/connections", app.listBuyerConnections)
			prv.Post("/buyer/connections", app.createBuyerConnection)
			prv.Post("/buyer/connections/{id}/rotate", app.rotateBuyerConnectionToken)
			prv.Post("/buyer/connections/{id}/revoke", app.revokeBuyerConnection)
			prv.Get("/buyer/entitlements", app.listBuyerEntitlements)
			prv.Get("/buyer/billing", app.getBuyerBilling)
			prv.Get("/buyer/invoices", app.listBuyerInvoices)
			prv.Get("/buyer/hub", app.getBuyerHub)
			prv.Get("/buyer/payments/controls", app.buyerPaymentControls)
			prv.Put("/buyer/payments/controls", app.buyerPaymentControls)
			prv.Get("/buyer/payments/topups", app.listBuyerWalletTopUps)
			prv.Post("/buyer/payments/topups/stripe/session", app.createBuyerStripeTopUpSession)
			prv.Get("/buyer/local-agents", app.listLocalAgents)
			prv.Post("/buyer/local-agents", app.upsertLocalAgent)
			prv.Get("/billing/x402/intents", app.listX402Intents)
			prv.Post("/billing/x402/intents", app.createX402Intent)
			prv.Post("/billing/x402/intents/{id}/settle", app.settleX402Intent)

			prv.Group(func(m chi.Router) {
				m.Use(app.requireRole("merchant", "admin"))
				m.Get("/merchant/servers", app.listMerchantServers)
				m.Post("/merchant/servers", app.createMerchantServer)
				m.Put("/merchant/servers/{id}", app.updateMerchantServer)
				m.Get("/merchant/revenue", app.merchantRevenue)
				m.Get("/merchant/servers/{id}/observability", app.serverObservability)
				m.Get("/merchant/servers/{id}/auth", app.serverAuthConfig)
				m.Get("/merchant/servers/{id}/pricing", app.serverPricing)
				m.Get("/merchant/servers/{id}/deployments", app.serverDeployments)
				m.Get("/merchant/servers/{id}/builder", app.serverBuilder)
				m.Get("/merchant/payments/overview", app.merchantPaymentsOverview)
				m.Get("/merchant/servers/{id}/payments/config", app.merchantServerPaymentConfig)
				m.Put("/merchant/servers/{id}/payments/config", app.merchantServerPaymentConfig)
			})

			prv.Group(func(a chi.Router) {
				a.Use(app.requireRole("admin"))
				a.Get("/admin/tenants", app.listTenants)
				a.Get("/admin/security-events", app.listSecurityEvents)
				a.Get("/admin/audit-logs", app.listAuditLogs)
				a.Get("/admin/client-compatibility", app.clientCompatibility)
				a.Get("/admin/payments/overview", app.adminPaymentsOverview)
				a.Post("/admin/entitlements", app.adminGrantEntitlement)
			})
		})
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
