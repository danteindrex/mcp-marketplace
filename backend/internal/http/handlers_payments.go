package http

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type updatePaymentControlsRequest struct {
	MonthlySpendCapUSDC float64  `json:"monthlySpendCapUsdc"`
	DailySpendCapUSDC   float64  `json:"dailySpendCapUsdc"`
	PerCallCapUSDC      float64  `json:"perCallCapUsdc"`
	AllowedMethods      []string `json:"allowedMethods"`
	SIWXWallet          string   `json:"siwxWallet"`
	MinimumBalanceUSDC  float64  `json:"minimumBalanceUsdc"`
	HardStopOnLowFunds  bool     `json:"hardStopOnLowFunds"`
	AutoTopUpEnabled    bool     `json:"autoTopUpEnabled"`
	AutoTopUpAmountUSD  float64  `json:"autoTopUpAmountUsd"`
	AutoTopUpTriggerUSD float64  `json:"autoTopUpTriggerUsd"`
	FundingMethod       string   `json:"fundingMethod"`
	WalletAddress       string   `json:"walletAddress"`
}

func (a *App) buyerPaymentControls(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	buyerWallet, walletErr := a.ensureBuyerManagedWallet(r.Context(), claims.TenantID, claims.UserID)
	if r.Method == http.MethodPut {
		var req updatePaymentControlsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		policy := models.PaymentPolicy{
			TenantID:            claims.TenantID,
			UserID:              claims.UserID,
			MonthlySpendCapUSDC: req.MonthlySpendCapUSDC,
			DailySpendCapUSDC:   req.DailySpendCapUSDC,
			PerCallCapUSDC:      req.PerCallCapUSDC,
			AllowedMethods:      normalizePaymentMethods(req.AllowedMethods),
			SIWXWallet:          strings.TrimSpace(req.SIWXWallet),
			MinimumBalanceUSDC:  clampNonNegative(req.MinimumBalanceUSDC),
			HardStopOnLowFunds:  req.HardStopOnLowFunds,
			AutoTopUpEnabled:    req.AutoTopUpEnabled,
			AutoTopUpAmountUSD:  clampNonNegative(req.AutoTopUpAmountUSD),
			AutoTopUpTriggerUSD: clampNonNegative(req.AutoTopUpTriggerUSD),
			FundingMethod:       strings.ToLower(strings.TrimSpace(req.FundingMethod)),
			WalletAddress:       strings.TrimSpace(req.WalletAddress),
		}
		existing, ok := a.store.GetPaymentPolicy(claims.TenantID, claims.UserID)
		if ok {
			policy.WalletBalanceUSDC = existing.WalletBalanceUSDC
			policy.LastTopUpAt = existing.LastTopUpAt
		}
		if len(policy.AllowedMethods) == 0 {
			policy.AllowedMethods = a.defaultAllowedPaymentMethods()
		} else if err := a.validateEnabledPaymentMethods(policy.AllowedMethods); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if policy.FundingMethod == "" {
			policy.FundingMethod = "stripe_onramp"
		}
		if strings.TrimSpace(policy.WalletAddress) == "" && walletErr == nil {
			policy.WalletAddress = buyerWallet.Address
		}
		if strings.TrimSpace(policy.SIWXWallet) == "" && walletErr == nil {
			policy.SIWXWallet = buyerWallet.Address
		}
		policy = a.store.UpsertPaymentPolicy(policy)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"policy":        policy,
			"methods":       a.paymentMethodsCatalog(),
			"managedWallet": buyerWallet,
			"walletConfig":  a.runtimePublicConfig()["wallet"],
		})
		return
	}

	policy := a.effectivePaymentPolicy(claims.TenantID, claims.UserID)
	if walletErr == nil {
		if strings.TrimSpace(policy.WalletAddress) == "" {
			policy.WalletAddress = buyerWallet.Address
		}
		if strings.TrimSpace(policy.SIWXWallet) == "" {
			policy.SIWXWallet = buyerWallet.Address
		}
	}
	intents := a.store.ListX402Intents(claims.TenantID, claims.UserID)
	topups := a.store.ListWalletTopUps(claims.TenantID, claims.UserID, 20)
	daily, monthly := settledSpendForWindow(intents, time.Now().UTC())
	integrations := a.resolvedIntegrations()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"policy":            policy,
		"methods":           a.paymentMethodsCatalog(),
		"dailySpendUsdc":    daily,
		"monthlySpendUsdc":  monthly,
		"dailyRemaining":    capRemaining(policy.DailySpendCapUSDC, daily),
		"monthlyRemaining":  capRemaining(policy.MonthlySpendCapUSDC, monthly),
		"facilitatorMode":   integrations.X402.Mode,
		"facilitatorTarget": integrations.X402.FacilitatorURL,
		"wallet": map[string]interface{}{
			"balanceUsdc":         policy.WalletBalanceUSDC,
			"minimumBalanceUsdc":  policy.MinimumBalanceUSDC,
			"hardStopOnLowFunds":  policy.HardStopOnLowFunds,
			"autoTopUpEnabled":    policy.AutoTopUpEnabled,
			"autoTopUpAmountUsd":  policy.AutoTopUpAmountUSD,
			"autoTopUpTriggerUsd": policy.AutoTopUpTriggerUSD,
			"fundingMethod":       policy.FundingMethod,
			"walletAddress":       policy.WalletAddress,
			"lastTopUpAt":         policy.LastTopUpAt,
		},
		"topups":       topups,
		"walletConfig": a.runtimePublicConfig()["wallet"],
		"managedWallet": map[string]interface{}{
			"id":          buyerWallet.ID,
			"provider":    buyerWallet.Provider,
			"network":     buyerWallet.Network,
			"asset":       buyerWallet.Asset,
			"address":     buyerWallet.Address,
			"status":      buyerWallet.Status,
			"custodyMode": buyerWallet.CustodyMode,
		},
	})
}

type updateServerPaymentConfigRequest struct {
	PaymentMethods []string `json:"paymentMethods"`
	PaymentAddress string   `json:"paymentAddress"`
	PerCallCapUSDC float64  `json:"perCallCapUsdc"`
	DailyCapUSDC   float64  `json:"dailyCapUsdc"`
	MonthlyCapUSDC float64  `json:"monthlyCapUsdc"`
}

func (a *App) merchantServerPaymentConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	server, ok := a.store.GetServerByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if !a.ensureServerTenantAccess(w, r, server) {
		return
	}

	if r.Method == http.MethodPut {
		var req updateServerPaymentConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		server.PaymentMethods = normalizePaymentMethods(req.PaymentMethods)
		if len(server.PaymentMethods) == 0 {
			server.PaymentMethods = a.defaultAllowedPaymentMethods()
		} else if err := a.validateEnabledPaymentMethods(server.PaymentMethods); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		server.PaymentAddress = strings.TrimSpace(req.PaymentAddress)
		if server.PaymentAddress == "" {
			if wallet, err := a.ensureSellerManagedWallet(r.Context(), server.TenantID); err == nil {
				server.PaymentAddress = wallet.Address
			}
		}
		server.PerCallCapUSDC = req.PerCallCapUSDC
		server.DailyCapUSDC = req.DailyCapUSDC
		server.MonthlyCapUSDC = req.MonthlyCapUSDC
		server.UpdatedAt = time.Now().UTC()
		a.store.UpdateServer(server)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"serverId": server.ID,
		"config": map[string]interface{}{
			"paymentMethods": server.PaymentMethods,
			"paymentAddress": server.PaymentAddress,
			"perCallCapUsdc": server.PerCallCapUSDC,
			"dailyCapUsdc":   server.DailyCapUSDC,
			"monthlyCapUsdc": server.MonthlyCapUSDC,
		},
		"methods": a.paymentMethodsCatalog(),
	})
}

func (a *App) merchantPaymentsOverview(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	servers := a.store.ListMerchantServers(claims.TenantID)
	serverName := map[string]string{}
	serverByID := map[string]models.Server{}
	serverSet := map[string]struct{}{}
	for _, srv := range servers {
		serverName[srv.ID] = srv.Name
		serverByID[srv.ID] = srv
		serverSet[srv.ID] = struct{}{}
	}

	all := a.store.ListAllX402Intents()
	type row struct {
		ServerID        string  `json:"serverId"`
		ServerName      string  `json:"serverName"`
		GrossUsdc       float64 `json:"grossUsdc"`
		PlatformFeeUsdc float64 `json:"platformFeeUsdc"`
		NetUsdc         float64 `json:"netUsdc"`
		SettledCnt      int     `json:"settledCount"`
		PendingCnt      int     `json:"pendingCount"`
	}
	rowsByServer := map[string]*row{}
	methodBreakdown := map[string]int{}
	totalGross := 0.0
	totalPlatform := 0.0
	totalNet := 0.0
	settledCount := 0
	pendingCount := 0
	for _, intent := range all {
		if _, ok := serverSet[intent.ServerID]; !ok {
			continue
		}
		method := intent.PaymentMethod
		if method == "" {
			method = "x402_wallet"
		}
		methodBreakdown[method]++

		rw, ok := rowsByServer[intent.ServerID]
		if !ok {
			rw = &row{
				ServerID:   intent.ServerID,
				ServerName: serverName[intent.ServerID],
			}
			rowsByServer[intent.ServerID] = rw
		}
		if intent.Status == "settled" {
			fee := roundUSDC(intent.PlatformFeeUSDC)
			net := roundUSDC(intent.SellerNetUSDC)
			if !intent.AccountingPosted || net <= 0 {
				if server, ok := serverByID[intent.ServerID]; ok {
					computedFee, computedNet := feeSplit(intent.AmountUSDC, a.effectiveFeePolicy(server))
					fee = computedFee
					net = computedNet
				} else {
					fee = roundUSDC(intent.AmountUSDC * float64(a.cfg.PlatformFeeBps) / 10000.0)
					if fee > intent.AmountUSDC {
						fee = intent.AmountUSDC
					}
					net = roundUSDC(intent.AmountUSDC - fee)
				}
			}
			totalGross += intent.AmountUSDC
			totalPlatform += fee
			totalNet += net
			settledCount++
			rw.GrossUsdc += intent.AmountUSDC
			rw.PlatformFeeUsdc += fee
			rw.NetUsdc += net
			rw.SettledCnt++
		} else {
			pendingCount++
			rw.PendingCnt++
		}
	}
	rows := make([]row, 0, len(rowsByServer))
	for _, r := range rowsByServer {
		r.GrossUsdc = roundUSDC(r.GrossUsdc)
		r.PlatformFeeUsdc = roundUSDC(r.PlatformFeeUsdc)
		r.NetUsdc = roundUSDC(r.NetUsdc)
		rows = append(rows, *r)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].NetUsdc > rows[j].NetUsdc
	})
	entries := a.store.ListLedgerEntries(claims.TenantID, 5000)
	profile := a.effectiveSellerPayoutProfile(claims.TenantID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"totalSettledUsdc":     roundUSDC(totalGross),
		"totalPlatformFeeUsdc": roundUSDC(totalPlatform),
		"totalNetUsdc":         roundUSDC(totalNet),
		"sellerPayableUsdc":    sellerPayableBalance(entries, claims.TenantID),
		"settledCount":         settledCount,
		"pendingCount":         pendingCount,
		"byServer":             rows,
		"methodBreakdown":      methodBreakdown,
		"methods":              a.paymentMethodsCatalog(),
		"payoutMethods":        a.payoutMethodsCatalog(),
		"payoutProfile":        profile,
		"recentPayouts":        a.store.ListPayoutRecords(claims.TenantID, 10),
	})
}

func (a *App) adminPaymentsOverview(w http.ResponseWriter, r *http.Request) {
	integrations := a.resolvedIntegrations()
	stripeConnect := a.currentStripeConnectService()
	all := a.store.ListAllX402Intents()
	total := 0.0
	settled := 0
	pending := 0
	failed := 0
	byMethod := map[string]int{}
	byTenant := map[string]float64{}
	for _, intent := range all {
		method := intent.PaymentMethod
		if method == "" {
			method = "x402_wallet"
		}
		byMethod[method]++
		switch intent.Status {
		case "settled":
			settled++
			total += intent.AmountUSDC
			server, ok := a.store.GetServerByID(intent.ServerID)
			if ok {
				byTenant[server.TenantID] += intent.AmountUSDC
			}
		case "failed":
			failed++
		default:
			pending++
		}
	}
	entries := a.store.ListLedgerEntries("", 10000)
	platformRevenue := 0.0
	sellerPayableByTenant := map[string]float64{}
	for _, entry := range entries {
		if entry.Account == accountPlatformRevenue {
			if strings.EqualFold(entry.EntryType, "credit") {
				platformRevenue += entry.AmountUSDC
			} else {
				platformRevenue -= entry.AmountUSDC
			}
		}
		if strings.HasPrefix(entry.Account, accountSellerPayable+":") {
			tenantID := strings.TrimPrefix(entry.Account, accountSellerPayable+":")
			if strings.EqualFold(entry.EntryType, "credit") {
				sellerPayableByTenant[tenantID] += entry.AmountUSDC
			} else {
				sellerPayableByTenant[tenantID] -= entry.AmountUSDC
			}
		}
	}
	for tenantID, amount := range sellerPayableByTenant {
		if amount < 0 {
			sellerPayableByTenant[tenantID] = 0
			continue
		}
		sellerPayableByTenant[tenantID] = roundUSDC(amount)
	}
	payouts := a.store.ListPayoutRecords("", 2000)
	byPayoutStatus := map[string]int{}
	payoutNet := 0.0
	for _, payout := range payouts {
		byPayoutStatus[payout.Status]++
		payoutNet += payout.NetUSDC
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"intentCount":           len(all),
		"settledCount":          settled,
		"pendingCount":          pending,
		"failedCount":           failed,
		"settledVolumeUsdc":     roundUSDC(total),
		"platformRevenueUsdc":   roundUSDC(platformRevenue),
		"sellerPayableByTenant": sellerPayableByTenant,
		"byMethod":              byMethod,
		"byTenant":              byTenant,
		"methods":               a.paymentMethodsCatalog(),
		"payouts": map[string]interface{}{
			"count":       len(payouts),
			"byStatus":    byPayoutStatus,
			"netSentUsdc": roundUSDC(payoutNet),
		},
		"feePolicies": a.store.ListPaymentFeePolicies(),
		"x402": map[string]interface{}{
			"mode":           integrations.X402.Mode,
			"facilitatorUrl": integrations.X402.FacilitatorURL,
			"apiKeySet":      strings.TrimSpace(integrations.X402.FacilitatorAPIKey) != "",
		},
		"wallet": map[string]interface{}{
			"provider":               a.resolveWalletProviderName(integrations.Wallet),
			"managedAutoPayEnabled":  integrations.Wallet.ManagedAutoPayEnabled,
			"legacyModeEnabled":      integrations.Wallet.LegacyPaymentModeEnabled,
			"externalWalletsEnabled": integrations.Wallet.ExternalWalletsEnabled,
		},
		"stripeConnect": map[string]interface{}{
			"configured": stripeConnect.configured(),
		},
	})
}

func capRemaining(cap float64, spend float64) float64 {
	if cap <= 0 {
		return -1
	}
	remaining := cap - spend
	if remaining < 0 {
		return 0
	}
	return remaining
}
