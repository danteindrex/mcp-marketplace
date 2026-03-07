package http

import (
	"fmt"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func (a *App) applyWalletCredit(tenantID, userID string, amount float64) (models.PaymentPolicy, error) {
	if amount <= 0 {
		return models.PaymentPolicy{}, fmt.Errorf("invalid wallet credit amount")
	}
	policy := a.effectivePaymentPolicy(tenantID, userID)
	policy.WalletBalanceUSDC += amount
	policy.LastTopUpAt = time.Now().UTC()
	policy = a.store.UpsertPaymentPolicy(policy)
	return policy, nil
}

func (a *App) applyWalletDebit(tenantID, userID string, amount float64, reference string) (models.PaymentPolicy, string, error) {
	policy := a.effectivePaymentPolicy(tenantID, userID)
	if err := walletDebitAllowed(policy, amount); err != nil {
		return policy, "", err
	}
	policy.WalletBalanceUSDC -= amount
	if policy.WalletBalanceUSDC < 0 {
		policy.WalletBalanceUSDC = 0
	}
	policy = a.store.UpsertPaymentPolicy(policy)
	ref := strings.TrimSpace(reference)
	if ref == "" {
		ref = "wallet_" + hashAny(map[string]interface{}{
			"tenantId": tenantID,
			"userId":   userID,
			"amount":   amount,
			"at":       time.Now().UTC().Format(time.RFC3339Nano),
		})[:16]
	}
	return policy, ref, nil
}
