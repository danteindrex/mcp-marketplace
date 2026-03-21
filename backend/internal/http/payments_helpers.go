package http

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type paymentMethodDescriptor struct {
	ID              string
	DisplayName     string
	Enabled         bool
	Configured      bool
	ProductionReady bool
	Readiness       string
	Integration     string
	Notes           string
	Docs            string
	Network         string
	Asset           string
}

func normalizePaymentMethods(in []string) []string {
	uniq := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		v := strings.ToLower(strings.TrimSpace(raw))
		if v == "" {
			continue
		}
		if _, ok := uniq[v]; ok {
			continue
		}
		uniq[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func (a *App) effectivePaymentPolicy(tenantID, userID string) models.PaymentPolicy {
	defaultMethods := a.defaultAllowedPaymentMethods()
	pol, ok := a.store.GetPaymentPolicy(tenantID, userID)
	if ok {
		return ensurePaymentPolicyDefaults(pol, defaultMethods)
	}
	return ensurePaymentPolicyDefaults(models.PaymentPolicy{
		TenantID:           tenantID,
		UserID:             userID,
		AllowedMethods:     defaultMethods,
		FundingMethod:      "stripe_onramp",
		HardStopOnLowFunds: true,
	}, defaultMethods)
}

func (a *App) paymentMethodsCatalog() []map[string]interface{} {
	supported := a.configuredPaymentMethods()
	out := make([]map[string]interface{}, 0, len(supported))
	for _, method := range supported {
		desc := a.describePaymentMethod(method)
		item := map[string]interface{}{
			"id":              desc.ID,
			"displayName":     desc.DisplayName,
			"enabled":         desc.Enabled,
			"configured":      desc.Configured,
			"productionReady": desc.ProductionReady,
			"readiness":       desc.Readiness,
			"integration":     desc.Integration,
			"notes":           desc.Notes,
			"docs":            desc.Docs,
		}
		if desc.Network != "" {
			item["network"] = desc.Network
		}
		if desc.Asset != "" {
			item["asset"] = desc.Asset
		}
		out = append(out, item)
	}
	return out
}

func (a *App) configuredPaymentMethods() []string {
	methods := normalizePaymentMethods(a.cfg.SupportedPayMethods)
	if len(methods) == 0 {
		return []string{"x402_wallet", "wallet_balance"}
	}
	return methods
}

func (a *App) defaultAllowedPaymentMethods() []string {
	configured := a.configuredPaymentMethods()
	defaults := make([]string, 0, len(configured))
	for _, method := range configured {
		if a.describePaymentMethod(method).Enabled {
			defaults = append(defaults, method)
		}
	}
	if len(defaults) == 0 {
		return []string{"x402_wallet", "wallet_balance"}
	}
	return defaults
}

func (a *App) isPaymentMethodEnabled(method string) bool {
	return a.describePaymentMethod(method).Enabled
}

func (a *App) validateEnabledPaymentMethods(methods []string) error {
	for _, method := range normalizePaymentMethods(methods) {
		desc := a.describePaymentMethod(method)
		if !desc.Enabled {
			return fmt.Errorf("payment method %q is not enabled: %s", method, desc.Notes)
		}
	}
	return nil
}

func (a *App) describePaymentMethod(method string) paymentMethodDescriptor {
	method = strings.ToLower(strings.TrimSpace(method))
	integrations := a.resolvedIntegrations()
	activeProvider := a.resolveWalletProviderName(integrations.Wallet)
	switch method {
	case "wallet_balance":
		enabled := integrations.Wallet.LegacyPaymentModeEnabled
		readiness := "production_ready"
		notes := "Usable today after the buyer wallet is funded."
		if !enabled {
			readiness = "disabled_by_admin"
			notes = "Legacy payment mode is disabled in admin controls."
		}
		return paymentMethodDescriptor{
			ID:              method,
			DisplayName:     "Prepaid Wallet Balance",
			Enabled:         enabled,
			Configured:      enabled,
			ProductionReady: enabled,
			Readiness:       readiness,
			Integration:     "deduct x402 spend from in-app USDC balance",
			Notes:           notes,
			Docs:            "Fund wallet via Stripe onramp or another supported balance credit flow, then spend with wallet_balance.",
			Network:         "base",
			Asset:           "USDC",
		}
	case "x402_wallet":
		enabled := integrations.Wallet.ManagedAutoPayEnabled && activeProvider != ""
		configured := enabled
		switch activeProvider {
		case "firefly":
			configured = configured && integrations.Wallet.FireflyEnabled && strings.TrimSpace(integrations.Wallet.FireflySignerURL) != ""
		case "cdp":
			configured = configured && integrations.Wallet.CDPEnabled && strings.TrimSpace(integrations.Wallet.CDPAPIKeyID) != "" &&
				strings.TrimSpace(integrations.Wallet.CDPAPIKeySecret) != "" && strings.TrimSpace(integrations.Wallet.CDPWalletSecret) != ""
		default:
			configured = false
		}
		productionReady := configured && strings.EqualFold(strings.TrimSpace(integrations.X402.Mode), "facilitator") && strings.TrimSpace(integrations.X402.FacilitatorURL) != ""
		readiness := "development_only"
		notes := "Managed marketplace wallet signs x402 payments automatically. Switch to facilitator mode for production settlement verification."
		if !integrations.Wallet.ManagedAutoPayEnabled {
			readiness = "disabled_by_admin"
			notes = "Managed wallet auto-pay is disabled in admin controls."
		} else if activeProvider == "firefly" {
			notes = "Uses FireFly Signer with Keystore V3/fswallet-compatible storage."
		} else if activeProvider == "cdp" {
			notes = "Uses the existing CDP managed wallet adapter."
		}
		if productionReady {
			readiness = "production_ready"
			notes = "Uses managed wallet signing with verified x402 facilitator settlement via " + activeProvider + "."
		}
		return paymentMethodDescriptor{
			ID:              method,
			DisplayName:     "Marketplace Wallet",
			Enabled:         enabled,
			Configured:      configured,
			ProductionReady: productionReady,
			Readiness:       readiness,
			Integration:     "provider-managed wallet signing with x402 facilitator verify/settle",
			Notes:           notes,
			Docs:            "Set the managed wallet provider plus X402 facilitator settings for verified settlement.",
			Network:         "base",
			Asset:           "USDC",
		}
	case "coinbase_commerce":
		return paymentMethodDescriptor{
			ID:              method,
			DisplayName:     "Coinbase Commerce",
			Enabled:         false,
			Configured:      false,
			ProductionReady: false,
			Readiness:       "not_configured",
			Integration:     "invoice + webhook settlement bridge",
			Notes:           "Hidden from defaults until session creation, webhook verification, and settlement mapping are implemented.",
			Docs:            "Provider adapter is not implemented yet.",
		}
	case "stripe":
		return paymentMethodDescriptor{
			ID:              method,
			DisplayName:     "Stripe",
			Enabled:         false,
			Configured:      false,
			ProductionReady: false,
			Readiness:       "not_configured",
			Integration:     "direct Stripe payment adapter",
			Notes:           "Stripe top-up and Stripe Connect exist, but direct x402 settlement via Stripe is not implemented.",
			Docs:            "Do not offer this as an x402 settlement method until adapter and webhook flows are complete.",
		}
	default:
		displayName := strings.ToUpper(strings.ReplaceAll(method, "_", " "))
		if method == "" {
			displayName = "UNKNOWN"
		}
		return paymentMethodDescriptor{
			ID:              method,
			DisplayName:     displayName,
			Enabled:         false,
			Configured:      false,
			ProductionReady: false,
			Readiness:       "custom_adapter_required",
			Integration:     "custom adapter",
			Notes:           "Implement create, verify, settle, and webhook handling before enabling this method.",
			Docs:            "Implement adapter to map provider settlement into verified x402 records.",
		}
	}
}

func ensurePaymentPolicyDefaults(policy models.PaymentPolicy, defaultMethods []string) models.PaymentPolicy {
	if len(policy.AllowedMethods) == 0 {
		policy.AllowedMethods = defaultMethods
	}
	policy.WalletBalanceUSDC = clampNonNegative(policy.WalletBalanceUSDC)
	policy.MinimumBalanceUSDC = clampNonNegative(policy.MinimumBalanceUSDC)
	policy.AutoTopUpAmountUSD = clampNonNegative(policy.AutoTopUpAmountUSD)
	policy.AutoTopUpTriggerUSD = clampNonNegative(policy.AutoTopUpTriggerUSD)
	if strings.TrimSpace(policy.FundingMethod) == "" {
		policy.FundingMethod = "stripe_onramp"
	}
	if !policy.HardStopOnLowFunds {
		// Keep guardrail enabled by default unless a policy already exists and explicitly disables it.
		if policy.UpdatedAt.IsZero() {
			policy.HardStopOnLowFunds = true
		}
	}
	return policy
}

func clampNonNegative(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

func walletDebitAllowed(policy models.PaymentPolicy, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("invalid wallet debit amount")
	}
	if policy.WalletBalanceUSDC < amount {
		return fmt.Errorf("insufficient prepaid wallet balance")
	}
	next := policy.WalletBalanceUSDC - amount
	if policy.HardStopOnLowFunds && next < policy.MinimumBalanceUSDC {
		return fmt.Errorf("minimum wallet balance would be breached")
	}
	return nil
}

func settledSpendForWindow(intents []models.X402Intent, now time.Time) (daily float64, monthly float64) {
	for _, intent := range intents {
		if intent.Status != "settled" || intent.SettledAt.IsZero() {
			continue
		}
		if intent.SettledAt.Year() == now.Year() && intent.SettledAt.Month() == now.Month() {
			monthly += intent.AmountUSDC
			if intent.SettledAt.Day() == now.Day() {
				daily += intent.AmountUSDC
			}
		}
	}
	return daily, monthly
}

func (a *App) validateCaps(policy models.PaymentPolicy, server models.Server, intents []models.X402Intent, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("invalid x402 amount")
	}
	if policy.PerCallCapUSDC > 0 && amount > policy.PerCallCapUSDC {
		return fmt.Errorf("user per-call cap exceeded")
	}
	if server.PerCallCapUSDC > 0 && amount > server.PerCallCapUSDC {
		return fmt.Errorf("server per-call cap exceeded")
	}

	now := time.Now().UTC()
	daily, monthly := settledSpendForWindow(intents, now)

	if policy.DailySpendCapUSDC > 0 && (daily+amount) > policy.DailySpendCapUSDC {
		return fmt.Errorf("user daily cap exceeded")
	}
	if policy.MonthlySpendCapUSDC > 0 && (monthly+amount) > policy.MonthlySpendCapUSDC {
		return fmt.Errorf("user monthly cap exceeded")
	}
	if server.DailyCapUSDC > 0 && (daily+amount) > server.DailyCapUSDC {
		return fmt.Errorf("server daily cap exceeded")
	}
	if server.MonthlyCapUSDC > 0 && (monthly+amount) > server.MonthlyCapUSDC {
		return fmt.Errorf("server monthly cap exceeded")
	}
	return nil
}
