package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type createX402IntentRequest struct {
	ServerID       string  `json:"serverId"`
	ToolName       string  `json:"toolName"`
	Amount         float64 `json:"amount"`
	PaymentMethod  string  `json:"paymentMethod"`
	IdempotencyKey string  `json:"idempotencyKey"`
}

type settleX402IntentRequest struct {
	PaymentResponse map[string]interface{} `json:"paymentResponse"`
}

func (a *App) createX402Intent(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	var req createX402IntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ServerID == "" || req.ToolName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	server, ok := a.store.GetServerByID(req.ServerID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	if server.PricingType != "x402" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "server is not configured for x402"})
		return
	}
	amount := server.PricingAmount
	if amount <= 0 {
		amount = req.Amount
	}
	if amount <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid x402 amount"})
		return
	}
	method := strings.ToLower(strings.TrimSpace(req.PaymentMethod))
	policy := a.effectivePaymentPolicy(claims.TenantID, claims.UserID)
	if method == "" {
		method = firstMethod(policy.AllowedMethods)
	}
	if !a.isPaymentMethodEnabled(method) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": a.describePaymentMethod(method).Notes})
		return
	}
	if !hasScope(policy.AllowedMethods, method) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "payment method is not allowed by buyer policy"})
		return
	}
	if len(server.PaymentMethods) > 0 && !hasScope(server.PaymentMethods, method) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "payment method is not enabled by server"})
		return
	}
	intents := a.store.ListX402Intents(claims.TenantID, claims.UserID)
	if err := a.validateCaps(policy, server, intents, amount); err != nil {
		writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
			"error": err.Error(),
			"caps":  policy,
		})
		return
	}
	if method == "wallet_balance" {
		if err := walletDebitAllowed(policy, amount); err != nil {
			writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
				"error":  err.Error(),
				"wallet": policy,
			})
			return
		}
	}
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	}
	if idempotencyKey == "" {
		idempotencyKey = "rpc_" + strings.TrimSpace(req.ToolName) + "_" + time.Now().UTC().Format(time.RFC3339Nano)
	}
	resource := server.CanonicalResourceURI
	if strings.TrimSpace(resource) == "" {
		resource = a.cfg.BaseURL + "/mcp/hub/" + claims.TenantID + "/" + claims.UserID
	}
	requirement := buildX402Requirement(server, req.ToolName, amount, resource, method, idempotencyKey)
	challengeBytes, _ := json.Marshal([]map[string]interface{}{requirement})
	intent := a.store.CreateX402Intent(models.X402Intent{
		TenantID:           claims.TenantID,
		UserID:             claims.UserID,
		ServerID:           req.ServerID,
		ToolName:           req.ToolName,
		AmountUSDC:         amount,
		Network:            "base",
		Asset:              "USDC",
		Challenge:          string(challengeBytes),
		PaymentMethod:      method,
		IdempotencyKey:     idempotencyKey,
		X402Version:        "2",
		Resource:           resource,
		VerificationStatus: "pending",
		Quantity:           1,
		RemainingQuantity:  0,
		RequestFingerprint: hashAny(map[string]interface{}{"serverId": req.ServerID, "toolName": req.ToolName, "idempotencyKey": idempotencyKey}),
	})
	w.Header().Set("PAYMENT-REQUIRED", string(challengeBytes))
	w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope"`)
	writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
		"intent":      intent,
		"requirement": requirement,
	})
}

func (a *App) settleX402Intent(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id := chi.URLParam(r, "id")
	existing, ok := a.store.GetX402Intent(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "intent not found"})
		return
	}
	if claims.Role != models.RoleAdmin && (existing.TenantID != claims.TenantID || existing.UserID != claims.UserID) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	if existing.Status == "settled" {
		writeJSON(w, http.StatusOK, existing)
		return
	}

	if strings.EqualFold(existing.PaymentMethod, "wallet_balance") {
		policy, paymentID, err := a.applyWalletDebit(existing.TenantID, existing.UserID, existing.AmountUSDC, existing.IdempotencyKey)
		if err != nil {
			writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
				"error":  err.Error(),
				"wallet": policy,
			})
			return
		}
		intent, ok := a.store.SettleX402Intent(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "intent not found"})
			return
		}
		intent.PaymentIdentifier = paymentID
		intent.PaymentMethod = "wallet_balance"
		intent.Network = "base"
		intent.Asset = "USDC"
		intent.VerificationStatus = "verified"
		intent.VerificationNote = "debited from prepaid wallet balance"
		intent.FacilitatorTx = ""
		if intent.Quantity <= 0 {
			intent.Quantity = 1
		}
		if intent.RemainingQuantity <= 0 {
			intent.RemainingQuantity = intent.Quantity
		}
		_ = a.store.UpdateX402Intent(intent)
		intent, err = a.postIntentAccounting(intent)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		a.ensurePaidEntitlement(intent)
		writeJSON(w, http.StatusOK, intent)
		return
	}
	var req settleX402IntentRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
	}
	requirement := firstRequirement(existing.Challenge)
	verifyRes, err := a.currentX402Service().verifyAndSettle(r.Context(), requirement, req.PaymentResponse)
	if err != nil {
		writeJSON(w, http.StatusPaymentRequired, map[string]string{"error": err.Error()})
		return
	}
	if !verifyRes.Valid {
		writeJSON(w, http.StatusPaymentRequired, map[string]string{"error": verifyRes.Note})
		return
	}
	if verifyRes.PaymentIdentifier != "" {
		for _, item := range a.store.ListX402Intents(existing.TenantID, existing.UserID) {
			if item.ID != existing.ID && item.Status == "settled" && item.PaymentIdentifier == verifyRes.PaymentIdentifier {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "payment identifier has already been settled"})
				return
			}
		}
	}
	intent, ok := a.store.SettleX402Intent(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "intent not found"})
		return
	}
	intent.PaymentIdentifier = verifyRes.PaymentIdentifier
	intent.PaymentMethod = verifyRes.Method
	intent.Network = verifyRes.Network
	intent.Asset = verifyRes.Asset
	intent.FacilitatorTx = verifyRes.TxHash
	intent.VerificationStatus = "verified"
	intent.VerificationNote = verifyRes.Note
	if intent.Quantity <= 0 {
		intent.Quantity = 1
	}
	if intent.RemainingQuantity <= 0 {
		intent.RemainingQuantity = intent.Quantity
	}
	_ = a.store.UpdateX402Intent(intent)
	intent, err = a.postIntentAccounting(intent)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	a.ensurePaidEntitlement(intent)
	writeJSON(w, http.StatusOK, intent)
}

func (a *App) listX402Intents(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	intents := a.store.ListX402Intents(claims.TenantID, claims.UserID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": intents, "count": len(intents)})
}

func buildX402Requirement(server models.Server, toolName string, amount float64, resource string, method string, idempotencyKey string) map[string]interface{} {
	return map[string]interface{}{
		"x402Version":     "2",
		"scheme":          "exact",
		"network":         "base",
		"asset":           "USDC",
		"amount":          amount,
		"resource":        resource,
		"serverId":        server.ID,
		"serverSlug":      server.Slug,
		"toolName":        toolName,
		"method":          method,
		"idempotencyKey":  idempotencyKey,
		"paymentAddress":  server.PaymentAddress,
		"paymentMethods":  server.PaymentMethods,
		"description":     "Pay to execute MCP tool via marketplace x402 proxy.",
		"maxAmountUsdc":   amount,
		"timeToLiveSecs":  300,
		"settlementScope": "single_call",
	}
}

func firstRequirement(challenge string) map[string]interface{} {
	out := map[string]interface{}{}
	raw := strings.TrimSpace(challenge)
	if raw == "" {
		return out
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &arr); err == nil && len(arr) > 0 {
		return arr[0]
	}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func (a *App) ensurePaidEntitlement(intent models.X402Intent) {
	entitlements := a.store.ListEntitlements(intent.TenantID, intent.UserID)
	for _, ent := range entitlements {
		if ent.ServerID == intent.ServerID && ent.Status == "active" {
			return
		}
	}
	server, ok := a.store.GetServerByID(intent.ServerID)
	if !ok {
		return
	}
	a.store.GrantEntitlement(models.Entitlement{
		TenantID:      intent.TenantID,
		UserID:        intent.UserID,
		ServerID:      intent.ServerID,
		AllowedScopes: server.RequiredScopes,
		CloudAllowed:  true,
		LocalAllowed:  true,
	})
}
