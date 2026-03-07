package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type updateMerchantPayoutProfileRequest struct {
	PreferredMethod   string  `json:"preferredMethod"`
	StablecoinAddress string  `json:"stablecoinAddress"`
	StablecoinNetwork string  `json:"stablecoinNetwork"`
	MinPayoutUSDC     float64 `json:"minPayoutUsdc"`
	HoldDays          int     `json:"holdDays"`
	TaxFormStatus     string  `json:"taxFormStatus"`
}

func (a *App) merchantPayoutProfile(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	profile := a.effectiveSellerPayoutProfile(claims.TenantID)
	if r.Method == http.MethodPut {
		var req updateMerchantPayoutProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		method := strings.ToLower(strings.TrimSpace(req.PreferredMethod))
		if method != "" && method != "stablecoin" && method != "stripe_connect" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "preferredMethod must be stablecoin or stripe_connect"})
			return
		}
		if method != "" {
			profile.PreferredMethod = method
		}
		if strings.TrimSpace(req.StablecoinAddress) != "" {
			profile.StablecoinAddress = strings.TrimSpace(req.StablecoinAddress)
		}
		if strings.TrimSpace(req.StablecoinNetwork) != "" {
			profile.StablecoinNetwork = strings.ToLower(strings.TrimSpace(req.StablecoinNetwork))
		}
		if req.MinPayoutUSDC >= 0 {
			profile.MinPayoutUSDC = roundUSDC(req.MinPayoutUSDC)
		}
		if req.HoldDays >= 0 {
			profile.HoldDays = req.HoldDays
		}
		if strings.TrimSpace(req.TaxFormStatus) != "" {
			profile.TaxFormStatus = strings.ToLower(strings.TrimSpace(req.TaxFormStatus))
		}
		profile = a.store.UpsertSellerPayoutProfile(profile)
	}
	entries := a.store.ListLedgerEntries(claims.TenantID, 2000)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"profile":        profile,
		"payableUsdc":    sellerPayableBalance(entries, claims.TenantID),
		"recentPayouts":  a.store.ListPayoutRecords(claims.TenantID, 20),
		"payoutMethods":  a.payoutMethodsCatalog(),
		"stripeConnect":  map[string]interface{}{"configured": a.stripeConnect.configured()},
		"reconciliation": map[string]interface{}{"imbalances": transactionImbalances(entries)},
	})
}

func (a *App) createMerchantStripeOnboardingLink(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	profile := a.effectiveSellerPayoutProfile(claims.TenantID)
	if strings.TrimSpace(profile.StripeAccountID) == "" {
		accountID, raw, err := a.stripeConnect.createExpressAccount(r.Context(), claims.TenantID)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		profile.StripeAccountID = accountID
		if len(raw) > 0 {
			applyConnectSnapshot(&profile, connectSnapshotFromObject(raw))
		}
	}
	link, err := a.stripeConnect.createOnboardingLink(r.Context(), profile.StripeAccountID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	profile.PreferredMethod = "stripe_connect"
	profile.StripeOnboardingURL = link
	profile = a.store.UpsertSellerPayoutProfile(profile)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"onboardingUrl": link,
		"profile":       profile,
		"configured":    a.stripeConnect.configured(),
	})
}

func (a *App) refreshMerchantStripeKYC(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	profile := a.effectiveSellerPayoutProfile(claims.TenantID)
	if strings.TrimSpace(profile.StripeAccountID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "stripe account is not connected"})
		return
	}
	snapshot, err := a.stripeConnect.fetchAccount(r.Context(), profile.StripeAccountID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	applyConnectSnapshot(&profile, snapshot)
	profile = a.store.UpsertSellerPayoutProfile(profile)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"profile": profile,
		"kyc": map[string]interface{}{
			"status":           profile.KYCStatus,
			"currentlyDue":     profile.KYCCurrentlyDue,
			"eventuallyDue":    profile.KYCEventuallyDue,
			"disabledReason":   profile.KYCDisabledReason,
			"chargesEnabled":   profile.KYCChargesEnabled,
			"payoutsEnabled":   profile.KYCPayoutsEnabled,
			"detailsSubmitted": profile.KYCDetailsSubmitted,
		},
	})
}

func (a *App) merchantLedger(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	entries := a.store.ListLedgerEntries(claims.TenantID, parseLimit(r.URL.Query().Get("limit"), 200, 2000))
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":           entries,
		"count":           len(entries),
		"payableUsdc":     sellerPayableBalance(entries, claims.TenantID),
		"imbalances":      transactionImbalances(entries),
		"supportedMethod": a.payoutMethodsCatalog(),
	})
}

func (a *App) merchantPayouts(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	records := a.store.ListPayoutRecords(claims.TenantID, parseLimit(r.URL.Query().Get("limit"), 100, 1000))
	byStatus := map[string]int{}
	for _, record := range records {
		byStatus[record.Status]++
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":    records,
		"count":    len(records),
		"byStatus": byStatus,
	})
}

func (a *App) adminFeePolicies(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"default": a.defaultFeePolicy(),
			"items":   a.store.ListPaymentFeePolicies(),
		})
		return
	}
	claims, _ := getClaims(r.Context())
	var req struct {
		Scope          string   `json:"scope"`
		TenantID       string   `json:"tenantId"`
		ServerID       string   `json:"serverId"`
		PlatformFeeBps *int     `json:"platformFeeBps"`
		MinFeeUSDC     *float64 `json:"minFeeUsdc"`
		MaxFeeUSDC     *float64 `json:"maxFeeUsdc"`
		HoldDays       *int     `json:"holdDays"`
		AutoPayouts    *bool    `json:"autoPayouts"`
		PayoutCadence  *string  `json:"payoutCadence"`
		Enabled        *bool    `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	scope := strings.ToLower(strings.TrimSpace(req.Scope))
	if scope != "global" && scope != "tenant" && scope != "server" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "scope must be global, tenant, or server"})
		return
	}
	tenantID := strings.TrimSpace(req.TenantID)
	serverID := strings.TrimSpace(req.ServerID)
	switch scope {
	case "global":
		tenantID = ""
		serverID = ""
	case "tenant":
		if tenantID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenantId is required for tenant scope"})
			return
		}
		if _, ok := a.store.GetTenantByID(tenantID); !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown tenant"})
			return
		}
		serverID = ""
	case "server":
		if tenantID == "" || serverID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenantId and serverId are required for server scope"})
			return
		}
		server, ok := a.store.GetServerByID(serverID)
		if !ok || server.TenantID != tenantID {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown server for tenant"})
			return
		}
	}

	policy := a.defaultFeePolicy()
	if existing, ok := a.store.GetPaymentFeePolicy(scope, tenantID, serverID); ok {
		policy = existing
	}
	policy.Scope = scope
	policy.TenantID = tenantID
	policy.ServerID = serverID
	policy.CreatedBy = claims.UserID
	if req.PlatformFeeBps != nil {
		if *req.PlatformFeeBps < 0 || *req.PlatformFeeBps > 10000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "platformFeeBps must be between 0 and 10000"})
			return
		}
		policy.PlatformFeeBps = *req.PlatformFeeBps
	}
	if req.MinFeeUSDC != nil {
		policy.MinFeeUSDC = clampNonNegative(*req.MinFeeUSDC)
	}
	if req.MaxFeeUSDC != nil {
		policy.MaxFeeUSDC = clampNonNegative(*req.MaxFeeUSDC)
	}
	if req.HoldDays != nil {
		if *req.HoldDays < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "holdDays cannot be negative"})
			return
		}
		policy.HoldDays = *req.HoldDays
	}
	if req.AutoPayouts != nil {
		policy.AutoPayouts = *req.AutoPayouts
	}
	if req.PayoutCadence != nil {
		cadence := strings.ToLower(strings.TrimSpace(*req.PayoutCadence))
		if cadence != "" {
			policy.PayoutCadence = cadence
		}
	}
	if req.Enabled != nil {
		policy.Enabled = *req.Enabled
	}
	policy = a.store.UpsertPaymentFeePolicy(policy)
	writeJSON(w, http.StatusOK, map[string]interface{}{"policy": policy})
}

func (a *App) adminLedger(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.URL.Query().Get("tenantId"))
	entries := a.store.ListLedgerEntries(tenantID, parseLimit(r.URL.Query().Get("limit"), 500, 5000))
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":      entries,
		"count":      len(entries),
		"imbalances": transactionImbalances(entries),
	})
}

func (a *App) adminReconciliation(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.URL.Query().Get("tenantId"))
	entries := a.store.ListLedgerEntries(tenantID, parseLimit(r.URL.Query().Get("limit"), 2000, 10000))
	accountSums := map[string]map[string]float64{}
	for _, entry := range entries {
		row, ok := accountSums[entry.Account]
		if !ok {
			row = map[string]float64{"debits": 0, "credits": 0}
			accountSums[entry.Account] = row
		}
		if strings.EqualFold(entry.EntryType, "debit") {
			row["debits"] += entry.AmountUSDC
		} else {
			row["credits"] += entry.AmountUSDC
		}
	}
	byAccount := make([]map[string]interface{}, 0, len(accountSums))
	for account, row := range accountSums {
		net := roundUSDC(row["credits"] - row["debits"])
		byAccount = append(byAccount, map[string]interface{}{
			"account": account,
			"debits":  roundUSDC(row["debits"]),
			"credits": roundUSDC(row["credits"]),
			"net":     net,
		})
	}
	sort.Slice(byAccount, func(i, j int) bool {
		return byAccount[i]["account"].(string) < byAccount[j]["account"].(string)
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tenantId":    tenantID,
		"count":       len(entries),
		"imbalances":  transactionImbalances(entries),
		"byAccount":   byAccount,
		"generatedAt": time.Now().UTC(),
	})
}

func (a *App) adminPayoutProfiles(w http.ResponseWriter, r *http.Request) {
	profiles := a.store.ListSellerPayoutProfiles()
	items := make([]map[string]interface{}, 0, len(profiles))
	for _, profile := range profiles {
		entries := a.store.ListLedgerEntries(profile.TenantID, 5000)
		items = append(items, map[string]interface{}{
			"profile":      profile,
			"payableUsdc":  sellerPayableBalance(entries, profile.TenantID),
			"payoutCount":  len(a.store.ListPayoutRecords(profile.TenantID, 500)),
			"kycCompleted": profile.KYCStatus == "verified",
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items, "count": len(items)})
}

func (a *App) adminSetPayoutBlock(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(chi.URLParam(r, "tenantID"))
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenant id is required"})
		return
	}
	if _, ok := a.store.GetTenantByID(tenantID); !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "tenant not found"})
		return
	}
	var req struct {
		Blocked bool   `json:"blocked"`
		Reason  string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	profile := a.effectiveSellerPayoutProfile(tenantID)
	profile.PayoutBlocked = req.Blocked
	profile.PayoutBlockReason = strings.TrimSpace(req.Reason)
	if profile.PayoutBlocked && profile.PayoutBlockReason == "" {
		profile.PayoutBlockReason = "blocked_by_admin"
	}
	profile = a.store.UpsertSellerPayoutProfile(profile)
	writeJSON(w, http.StatusOK, map[string]interface{}{"profile": profile})
}

func (a *App) adminPayouts(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.URL.Query().Get("tenantId"))
	records := a.store.ListPayoutRecords(tenantID, parseLimit(r.URL.Query().Get("limit"), 200, 2000))
	byStatus := map[string]int{}
	totalNet := 0.0
	totalFees := 0.0
	for _, record := range records {
		byStatus[record.Status]++
		totalNet += record.NetUSDC
		totalFees += record.FeeUSDC
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":         records,
		"count":         len(records),
		"byStatus":      byStatus,
		"totalNetUsdc":  roundUSDC(totalNet),
		"totalFeesUsdc": roundUSDC(totalFees),
	})
}

func (a *App) adminRunPayout(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	var req struct {
		TenantID   string  `json:"tenantId"`
		AmountUSDC float64 `json:"amountUsdc"`
		Method     string  `json:"method"`
		DryRun     bool    `json:"dryRun"`
		Force      bool    `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenantId is required"})
		return
	}
	if _, ok := a.store.GetTenantByID(tenantID); !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "tenant not found"})
		return
	}
	profile := a.effectiveSellerPayoutProfile(tenantID)
	method := strings.ToLower(strings.TrimSpace(req.Method))
	if method == "" {
		method = profile.PreferredMethod
	}
	if method == "" {
		method = "stablecoin"
	}
	entries := a.store.ListLedgerEntries(tenantID, 10000)
	payable := sellerPayableBalance(entries, tenantID)
	amount := payable
	if req.AmountUSDC > 0 && req.AmountUSDC < amount {
		amount = roundUSDC(req.AmountUSDC)
	}
	if amount <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no payable balance available"})
		return
	}
	minimum := profile.MinPayoutUSDC
	if minimum <= 0 {
		minimum = 10
	}
	if !req.Force && amount < minimum {
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"error":             fmt.Sprintf("amount %.2f is below minimum payout %.2f", amount, minimum),
			"amountUsdc":        amount,
			"minimumPayoutUsdc": minimum,
		})
		return
	}
	if profile.PayoutBlocked && !req.Force {
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"error":  "payouts are blocked for tenant",
			"reason": profile.PayoutBlockReason,
		})
		return
	}
	if method == "stripe_connect" && strings.TrimSpace(profile.StripeAccountID) == "" {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "seller stripe account is not connected"})
		return
	}
	if method == "stripe_connect" && !req.Force && profile.KYCStatus != "verified" {
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"error":            "seller stripe kyc is not verified",
			"kycStatus":        profile.KYCStatus,
			"currentlyDue":     profile.KYCCurrentlyDue,
			"disabledReason":   profile.KYCDisabledReason,
			"detailsSubmitted": profile.KYCDetailsSubmitted,
		})
		return
	}
	if req.DryRun {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"dryRun":            true,
			"tenantId":          tenantID,
			"method":            method,
			"amountUsdc":        amount,
			"payableUsdc":       payable,
			"minimumPayoutUsdc": minimum,
			"profile":           profile,
		})
		return
	}

	record := a.store.CreatePayoutRecord(models.PayoutRecord{
		TenantID:          tenantID,
		Method:            method,
		Status:            "processing",
		AmountUSDC:        amount,
		NetUSDC:           amount,
		StripeAccountID:   profile.StripeAccountID,
		StablecoinAddress: profile.StablecoinAddress,
		Metadata: map[string]interface{}{
			"triggeredBy": claims.UserID,
		},
	})

	switch method {
	case "stripe_connect":
		ref, fee, err := a.stripeConnect.createPayout(r.Context(), profile.StripeAccountID, amount, map[string]string{
			"tenant_id": tenantID,
			"payout_id": record.ID,
		})
		if err != nil {
			record.Status = "failed"
			record.FailureReason = err.Error()
			_ = a.store.UpdatePayoutRecord(record)
			writeJSON(w, http.StatusBadGateway, map[string]interface{}{"record": record})
			return
		}
		record.ExternalRef = ref
		record.FeeUSDC = roundUSDC(fee)
		record.NetUSDC = roundUSDC(amount - fee)
	default:
		ref := "usdc_tx_" + hashAny(map[string]interface{}{
			"payoutId": record.ID,
			"tenantId": tenantID,
			"amount":   amount,
			"at":       time.Now().UTC().Format(time.RFC3339Nano),
		})[:20]
		record.ExternalRef = ref
		record.FeeUSDC = 0
		record.NetUSDC = amount
	}

	record.Status = "completed"
	record.CompletedAt = time.Now().UTC()
	if err := a.postPayoutAccounting(record); err != nil {
		record.Status = "failed"
		record.FailureReason = err.Error()
		_ = a.store.UpdatePayoutRecord(record)
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"record": record})
		return
	}
	_ = a.store.UpdatePayoutRecord(record)
	a.store.AddAuditLog(models.AuditLog{
		TenantID:   tenantID,
		ActorID:    claims.UserID,
		Action:     "payments.payout.run",
		TargetType: "payout",
		TargetID:   record.ID,
		Outcome:    "success",
		Metadata:   map[string]interface{}{"method": method, "amountUsdc": amount},
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"record":            record,
		"payableBeforeUsdc": payable,
		"payableAfterUsdc":  sellerPayableBalance(a.store.ListLedgerEntries(tenantID, 10000), tenantID),
	})
}

func (a *App) handleStripeConnectWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook payload"})
		return
	}
	if err := a.stripeConnect.verifyWebhookSignature(payload, r.Header.Get("Stripe-Signature")); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	event := map[string]interface{}{}
	if err := json.Unmarshal(payload, &event); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook json"})
		return
	}
	eventType := strings.ToLower(strings.TrimSpace(stringFromAny(event["type"])))
	if !strings.Contains(eventType, "account.updated") {
		writeJSON(w, http.StatusAccepted, map[string]interface{}{"received": true, "ignored": eventType})
		return
	}
	data, _ := event["data"].(map[string]interface{})
	object, _ := data["object"].(map[string]interface{})
	if len(object) == 0 {
		writeJSON(w, http.StatusAccepted, map[string]interface{}{"received": true, "ignored": "missing object"})
		return
	}
	snapshot := connectSnapshotFromObject(object)
	if strings.TrimSpace(snapshot.AccountID) == "" {
		writeJSON(w, http.StatusAccepted, map[string]interface{}{"received": true, "ignored": "missing account id"})
		return
	}
	updated := 0
	for _, profile := range a.store.ListSellerPayoutProfiles() {
		if strings.TrimSpace(profile.StripeAccountID) != snapshot.AccountID {
			continue
		}
		applyConnectSnapshot(&profile, snapshot)
		_ = a.store.UpsertSellerPayoutProfile(profile)
		updated++
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"received": true, "updatedProfiles": updated})
}

func applyConnectSnapshot(profile *models.SellerPayoutProfile, snapshot stripeConnectAccountSnapshot) {
	profile.StripeAccountID = snapshot.AccountID
	profile.KYCDetailsSubmitted = snapshot.DetailsSubmitted
	profile.KYCChargesEnabled = snapshot.ChargesEnabled
	profile.KYCPayoutsEnabled = snapshot.PayoutsEnabled
	profile.KYCCurrentlyDue = snapshot.CurrentlyDue
	profile.KYCEventuallyDue = snapshot.EventuallyDue
	profile.KYCDisabledReason = snapshot.DisabledReason
	profile.KYCLastCheckedAt = time.Now().UTC()
	profile.KYCStatus = deriveKYCStatus(snapshot)
}

func deriveKYCStatus(snapshot stripeConnectAccountSnapshot) string {
	if strings.TrimSpace(snapshot.DisabledReason) != "" {
		return "restricted"
	}
	if len(snapshot.CurrentlyDue) > 0 || len(snapshot.PendingVerification) > 0 {
		return "pending"
	}
	if snapshot.DetailsSubmitted && snapshot.ChargesEnabled && snapshot.PayoutsEnabled {
		return "verified"
	}
	return "pending"
}

func parseLimit(raw string, fallback int, max int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	if n > max {
		return max
	}
	return n
}

func (a *App) defaultSellerPayoutProfile(tenantID string) models.SellerPayoutProfile {
	now := time.Now().UTC()
	return models.SellerPayoutProfile{
		TenantID:          tenantID,
		PreferredMethod:   "stablecoin",
		StablecoinNetwork: "base",
		KYCStatus:         "pending",
		TaxFormStatus:     "missing",
		RiskLevel:         "medium",
		MinPayoutUSDC:     10,
		HoldDays:          a.cfg.PlatformHoldDays,
		UpdatedAt:         now,
		CreatedAt:         now,
	}
}

func (a *App) effectiveSellerPayoutProfile(tenantID string) models.SellerPayoutProfile {
	if item, ok := a.store.GetSellerPayoutProfile(tenantID); ok {
		if strings.TrimSpace(item.PreferredMethod) == "" {
			item.PreferredMethod = "stablecoin"
		}
		if strings.TrimSpace(item.StablecoinNetwork) == "" {
			item.StablecoinNetwork = "base"
		}
		if item.MinPayoutUSDC <= 0 {
			item.MinPayoutUSDC = 10
		}
		if item.HoldDays < 0 {
			item.HoldDays = 0
		}
		return item
	}
	return a.defaultSellerPayoutProfile(tenantID)
}

func (a *App) payoutMethodsCatalog() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":          "stablecoin",
			"displayName": "Stablecoin (USDC)",
			"enabled":     true,
			"configured":  true,
			"network":     "base",
			"asset":       "USDC",
			"notes":       "Direct USDC payout to seller wallet.",
		},
		{
			"id":          "stripe_connect",
			"displayName": "Stripe Connect",
			"enabled":     true,
			"configured":  a.stripeConnect.configured(),
			"notes":       "Hosted onboarding with Stripe KYC requirements and managed payouts.",
			"docs":        "https://docs.stripe.com/connect",
		},
	}
}
