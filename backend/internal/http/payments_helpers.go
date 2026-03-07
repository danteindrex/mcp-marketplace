package http

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

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
	defaultMethods := a.supportedMethodsOrDefault()
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
	supported := a.supportedMethodsOrDefault()
	out := make([]map[string]interface{}, 0, len(supported))
	for _, method := range supported {
		item := map[string]interface{}{
			"id":          method,
			"enabled":     true,
			"configured":  true,
			"integration": "native",
			"notes":       "",
		}
		switch method {
		case "wallet_balance":
			item["displayName"] = "Prepaid Wallet Balance"
			item["integration"] = "deduct x402 spend from in-app USDC balance"
			item["network"] = "base"
			item["asset"] = "USDC"
			item["docs"] = "Fund wallet via Stripe onramp session, then use wallet_balance for no-subscription spend."
			item["configured"] = true
		case "x402_wallet":
			item["displayName"] = "x402 Wallet Signature"
			item["integration"] = "x402 payment-response with facilitator verify/settle"
			item["network"] = "base"
			item["asset"] = "USDC"
			item["docs"] = "Set X402_MODE=facilitator and X402_FACILITATOR_URL for production verification."
			item["configured"] = strings.TrimSpace(a.cfg.X402Mode) != ""
		case "coinbase_commerce":
			item["displayName"] = "Coinbase Commerce"
			item["integration"] = "invoice + webhook settlement bridge"
			item["docs"] = "Configure COINBASE_COMMERCE_API_KEY and webhook secret in your deployment."
			item["configured"] = strings.TrimSpace(os.Getenv("COINBASE_COMMERCE_API_KEY")) != ""
		case "stripe":
			item["displayName"] = "Stripe"
			item["integration"] = "embedded onramp + webhook wallet credit"
			item["docs"] = "Configure STRIPE_SECRET_KEY and STRIPE_WEBHOOK_SECRET for onramp sessions."
			item["configured"] = strings.TrimSpace(a.cfg.StripeSecretKey) != ""
		default:
			item["displayName"] = strings.ToUpper(strings.ReplaceAll(method, "_", " "))
			item["integration"] = "custom adapter"
			item["docs"] = "Implement adapter to map provider settlement into x402 payment-response."
		}
		out = append(out, item)
	}
	return out
}

func (a *App) supportedMethodsOrDefault() []string {
	methods := normalizePaymentMethods(a.cfg.SupportedPayMethods)
	if len(methods) == 0 {
		return []string{"x402_wallet", "wallet_balance"}
	}
	return methods
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
