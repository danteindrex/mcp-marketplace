package http

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

const (
	accountBuyerWalletPrefix = "buyer_wallet"
	accountExternalClearing  = "external_clearing"
	accountSellerPayable     = "seller_payable"
	accountPlatformRevenue   = "platform_revenue"
	accountPayoutsSent       = "seller_payouts_sent"
	accountOnrampClearing    = "onramp_clearing"
)

func (a *App) defaultFeePolicy() models.PaymentFeePolicy {
	now := time.Now().UTC()
	return models.PaymentFeePolicy{
		ID:             "feepol_default",
		Scope:          "global",
		PlatformFeeBps: a.cfg.PlatformFeeBps,
		MinFeeUSDC:     a.cfg.PlatformMinFeeUSDC,
		MaxFeeUSDC:     a.cfg.PlatformMaxFeeUSDC,
		HoldDays:       a.cfg.PlatformHoldDays,
		AutoPayouts:    false,
		PayoutCadence:  "manual",
		Enabled:        true,
		CreatedBy:      "system",
		UpdatedAt:      now,
		CreatedAt:      now,
	}
}

func (a *App) effectiveFeePolicy(server models.Server) models.PaymentFeePolicy {
	policy := a.defaultFeePolicy()
	if p, ok := a.store.GetPaymentFeePolicy("global", "", ""); ok && p.Enabled {
		policy = p
	}
	if p, ok := a.store.GetPaymentFeePolicy("tenant", server.TenantID, ""); ok && p.Enabled {
		policy = mergeFeePolicy(policy, p)
	}
	if p, ok := a.store.GetPaymentFeePolicy("server", server.TenantID, server.ID); ok && p.Enabled {
		policy = mergeFeePolicy(policy, p)
	}
	if policy.PlatformFeeBps < 0 {
		policy.PlatformFeeBps = 0
	}
	if policy.PlatformFeeBps > 10000 {
		policy.PlatformFeeBps = 10000
	}
	return policy
}

func mergeFeePolicy(base models.PaymentFeePolicy, override models.PaymentFeePolicy) models.PaymentFeePolicy {
	merged := base
	merged.ID = override.ID
	merged.Scope = override.Scope
	merged.TenantID = override.TenantID
	merged.ServerID = override.ServerID
	merged.PlatformFeeBps = override.PlatformFeeBps
	merged.MinFeeUSDC = override.MinFeeUSDC
	merged.MaxFeeUSDC = override.MaxFeeUSDC
	merged.HoldDays = override.HoldDays
	merged.AutoPayouts = override.AutoPayouts
	if strings.TrimSpace(override.PayoutCadence) != "" {
		merged.PayoutCadence = override.PayoutCadence
	}
	merged.Enabled = override.Enabled
	merged.CreatedBy = override.CreatedBy
	merged.UpdatedAt = override.UpdatedAt
	merged.CreatedAt = override.CreatedAt
	return merged
}

func feeSplit(amount float64, policy models.PaymentFeePolicy) (platformFee float64, sellerNet float64) {
	if amount <= 0 {
		return 0, 0
	}
	fee := roundUSDC(amount * float64(policy.PlatformFeeBps) / 10000.0)
	if policy.MinFeeUSDC > 0 && fee < policy.MinFeeUSDC {
		fee = policy.MinFeeUSDC
	}
	if policy.MaxFeeUSDC > 0 && fee > policy.MaxFeeUSDC {
		fee = policy.MaxFeeUSDC
	}
	if fee > amount {
		fee = amount
	}
	net := roundUSDC(amount - fee)
	return roundUSDC(fee), net
}

func roundUSDC(v float64) float64 {
	return math.Round(v*1_000_000) / 1_000_000
}

func buyerWalletAccount(tenantID, userID string) string {
	return accountBuyerWalletPrefix + ":" + strings.TrimSpace(tenantID) + ":" + strings.TrimSpace(userID)
}

func sellerPayableAccount(tenantID string) string {
	return accountSellerPayable + ":" + strings.TrimSpace(tenantID)
}

func payoutsSentAccount(tenantID string) string {
	return accountPayoutsSent + ":" + strings.TrimSpace(tenantID)
}

func (a *App) postIntentAccounting(intent models.X402Intent) (models.X402Intent, error) {
	if intent.Status != "settled" {
		return intent, nil
	}
	if intent.AccountingPosted {
		return intent, nil
	}
	server, ok := a.store.GetServerByID(intent.ServerID)
	if !ok {
		return intent, fmt.Errorf("server not found for accounting")
	}
	policy := a.effectiveFeePolicy(server)
	platformFee, sellerNet := feeSplit(intent.AmountUSDC, policy)
	txnID := "txn_intent_" + intent.ID

	debitAccount := accountExternalClearing
	if strings.EqualFold(intent.PaymentMethod, "wallet_balance") {
		debitAccount = buyerWalletAccount(intent.TenantID, intent.UserID)
	}

	entries := []models.LedgerEntry{
		{
			TransactionID: txnID,
			TenantID:      intent.TenantID,
			UserID:        intent.UserID,
			ServerID:      intent.ServerID,
			IntentID:      intent.ID,
			Account:       debitAccount,
			EntryType:     "debit",
			AmountUSDC:    roundUSDC(intent.AmountUSDC),
			Category:      "intent_settlement",
			Description:   "Buyer settlement for MCP usage",
			Reference:     intent.PaymentIdentifier,
		},
		{
			TransactionID: txnID,
			TenantID:      server.TenantID,
			ServerID:      intent.ServerID,
			IntentID:      intent.ID,
			Account:       sellerPayableAccount(server.TenantID),
			EntryType:     "credit",
			AmountUSDC:    sellerNet,
			Category:      "seller_payable",
			Description:   "Seller payable accrual",
			Reference:     intent.PaymentIdentifier,
		},
	}
	if platformFee > 0 {
		entries = append(entries, models.LedgerEntry{
			TransactionID: txnID,
			TenantID:      "tenant_platform",
			ServerID:      intent.ServerID,
			IntentID:      intent.ID,
			Account:       accountPlatformRevenue,
			EntryType:     "credit",
			AmountUSDC:    platformFee,
			Category:      "platform_fee",
			Description:   "Platform take rate",
			Reference:     intent.PaymentIdentifier,
		})
	}
	a.store.CreateLedgerEntries(entries)
	intent.PlatformFeeBps = policy.PlatformFeeBps
	intent.PlatformFeeUSDC = platformFee
	intent.SellerNetUSDC = sellerNet
	intent.AccountingPosted = true
	intent.AccountingRef = txnID
	if ok := a.store.UpdateX402Intent(intent); !ok {
		return intent, fmt.Errorf("failed to persist intent accounting")
	}
	return intent, nil
}

func (a *App) postTopUpAccounting(topup models.WalletTopUp) (models.WalletTopUp, error) {
	if topup.AccountingPosted {
		return topup, nil
	}
	amount := roundUSDC(topup.DestinationAmount)
	if amount <= 0 {
		return topup, fmt.Errorf("invalid top-up accounting amount")
	}
	txnID := "txn_topup_" + topup.ID
	entries := []models.LedgerEntry{
		{
			TransactionID: txnID,
			TenantID:      topup.TenantID,
			UserID:        topup.UserID,
			TopUpID:       topup.ID,
			Account:       accountOnrampClearing,
			EntryType:     "debit",
			AmountUSDC:    amount,
			Category:      "topup",
			Description:   "Onramp transfer received",
			Reference:     topup.ProviderSessionID,
		},
		{
			TransactionID: txnID,
			TenantID:      topup.TenantID,
			UserID:        topup.UserID,
			TopUpID:       topup.ID,
			Account:       buyerWalletAccount(topup.TenantID, topup.UserID),
			EntryType:     "credit",
			AmountUSDC:    amount,
			Category:      "topup",
			Description:   "Buyer wallet funded via onramp",
			Reference:     topup.ProviderSessionID,
		},
	}
	a.store.CreateLedgerEntries(entries)
	topup.AccountingPosted = true
	topup.AccountingRef = txnID
	if !a.store.UpdateWalletTopUp(topup) {
		return topup, fmt.Errorf("failed to persist top-up accounting")
	}
	return topup, nil
}

func (a *App) postPayoutAccounting(record models.PayoutRecord) error {
	if strings.TrimSpace(record.ID) == "" {
		return fmt.Errorf("payout id is required")
	}
	amount := roundUSDC(record.NetUSDC)
	if amount <= 0 {
		amount = roundUSDC(record.AmountUSDC)
	}
	if amount <= 0 {
		return fmt.Errorf("invalid payout amount")
	}
	txnID := "txn_payout_" + record.ID
	entries := []models.LedgerEntry{
		{
			TransactionID: txnID,
			TenantID:      record.TenantID,
			PayoutID:      record.ID,
			Account:       sellerPayableAccount(record.TenantID),
			EntryType:     "debit",
			AmountUSDC:    amount,
			Category:      "seller_payout",
			Description:   "Seller payable payout released",
			Reference:     record.ExternalRef,
		},
		{
			TransactionID: txnID,
			TenantID:      record.TenantID,
			PayoutID:      record.ID,
			Account:       payoutsSentAccount(record.TenantID),
			EntryType:     "credit",
			AmountUSDC:    amount,
			Category:      "seller_payout",
			Description:   "Seller payout sent",
			Reference:     record.ExternalRef,
		},
	}
	a.store.CreateLedgerEntries(entries)
	return nil
}

func sellerPayableBalance(entries []models.LedgerEntry, tenantID string) float64 {
	account := sellerPayableAccount(tenantID)
	balance := 0.0
	for _, entry := range entries {
		if entry.Account != account {
			continue
		}
		if strings.EqualFold(entry.EntryType, "credit") {
			balance += entry.AmountUSDC
			continue
		}
		balance -= entry.AmountUSDC
	}
	if balance < 0 {
		return 0
	}
	return roundUSDC(balance)
}

func transactionImbalances(entries []models.LedgerEntry) []map[string]interface{} {
	type sums struct {
		debits  float64
		credits float64
	}
	acc := map[string]*sums{}
	for _, entry := range entries {
		item, ok := acc[entry.TransactionID]
		if !ok {
			item = &sums{}
			acc[entry.TransactionID] = item
		}
		if strings.EqualFold(entry.EntryType, "debit") {
			item.debits += entry.AmountUSDC
		} else {
			item.credits += entry.AmountUSDC
		}
	}
	out := make([]map[string]interface{}, 0)
	for tx, sum := range acc {
		if roundUSDC(sum.debits) == roundUSDC(sum.credits) {
			continue
		}
		out = append(out, map[string]interface{}{
			"transactionId": tx,
			"debits":        roundUSDC(sum.debits),
			"credits":       roundUSDC(sum.credits),
			"delta":         roundUSDC(sum.debits - sum.credits),
		})
	}
	return out
}
