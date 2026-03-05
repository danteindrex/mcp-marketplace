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
	cfg   config.Config
	store store.Store
	jwt   *auth.JWTManager
	oauth *oauthState
}

func NewRouter(cfg config.Config, st store.Store, jwt *auth.JWTManager) http.Handler {
	app := &App{cfg: cfg, store: st, jwt: jwt, oauth: newOAuthState()}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(app.cors)
	r.Use(middleware.Heartbeat("/healthz"))
	r.Use(app.securityHeaders)

	r.Get("/health", app.health)
	r.Post("/auth/signup", app.signup)
	r.Post("/auth/login", app.login)
	r.Get("/.well-known/oauth-protected-resource", app.oauthProtectedResourceMetadata)
	r.Get("/.well-known/oauth-authorization-server", app.oauthAuthorizationServerMetadata)
	r.Post("/oauth/register", app.oauthRegisterClient)
	r.Get("/oauth/authorize", app.oauthAuthorize)
	r.Post("/oauth/token", app.oauthToken)

	r.Route("/v1", func(v1 chi.Router) {
		v1.Get("/marketplace/servers", app.listMarketplaceServers)
		v1.Get("/marketplace/servers/{slug}", app.getMarketplaceServer)

		v1.Group(func(prv chi.Router) {
			prv.Use(app.authenticate)
			prv.Get("/me", app.me)
			prv.Get("/buyer/connections", app.listBuyerConnections)
			prv.Post("/buyer/connections", app.createBuyerConnection)
			prv.Post("/buyer/connections/{id}/rotate", app.rotateBuyerConnectionToken)
			prv.Post("/buyer/connections/{id}/revoke", app.revokeBuyerConnection)
			prv.Get("/buyer/entitlements", app.listBuyerEntitlements)
			prv.Get("/buyer/billing", app.getBuyerBilling)
			prv.Get("/buyer/invoices", app.listBuyerInvoices)
			prv.Get("/buyer/hub", app.getBuyerHub)
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
			})

			prv.Group(func(a chi.Router) {
				a.Use(app.requireRole("admin"))
				a.Get("/admin/tenants", app.listTenants)
				a.Get("/admin/security-events", app.listSecurityEvents)
				a.Get("/admin/audit-logs", app.listAuditLogs)
				a.Get("/admin/client-compatibility", app.clientCompatibility)
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
