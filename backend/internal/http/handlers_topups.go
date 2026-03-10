package http

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type createStripeTopUpSessionRequest struct {
	AmountUSD     float64 `json:"amountUsd"`
	WalletAddress string  `json:"walletAddress"`
	PaymentMethod string  `json:"paymentMethod"`
}

func (a *App) listBuyerWalletTopUps(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	items := a.store.ListWalletTopUps(claims.TenantID, claims.UserID, 50)
	stripe := a.currentStripeOnrampService()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":            items,
		"count":            len(items),
		"minimumTopUpUsd":  stripe.minTopupUSD,
		"defaultTopUpUsd":  stripe.defaultUSD,
		"stripeConfigured": stripe.configured(),
	})
}

func (a *App) createBuyerStripeTopUpSession(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	var req createStripeTopUpSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	amount := req.AmountUSD
	stripe := a.currentStripeOnrampService()
	if amount <= 0 {
		amount = stripe.defaultUSD
	}
	if amount < stripe.minTopupUSD {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("minimum top-up is %.2f USD", stripe.minTopupUSD),
		})
		return
	}
	if amount > 100000 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "top-up amount exceeds allowed maximum"})
		return
	}

	policy := a.effectivePaymentPolicy(claims.TenantID, claims.UserID)
	walletAddress := strings.TrimSpace(req.WalletAddress)
	if walletAddress == "" {
		walletAddress = strings.TrimSpace(policy.WalletAddress)
	}
	if walletAddress == "" {
		walletAddress = strings.TrimSpace(policy.SIWXWallet)
	}

	session, err := stripe.createSession(r.Context(), stripeCreateSessionInput{
		AmountUSD:     amount,
		WalletAddress: walletAddress,
		CustomerIP:    firstForwardedIP(r.Header.Get("X-Forwarded-For")),
		TenantID:      claims.TenantID,
		UserID:        claims.UserID,
	})
	if err != nil {
		status := http.StatusBadGateway
		if stripeErr, ok := err.(*stripeAPIError); ok {
			if stripeErr.Status == http.StatusBadRequest || stripeErr.Status == http.StatusNotFound {
				status = http.StatusConflict
			}
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	method := strings.ToLower(strings.TrimSpace(req.PaymentMethod))
	if method == "" {
		method = "stripe_onramp"
	}
	topup := a.store.CreateWalletTopUp(models.WalletTopUp{
		TenantID:           claims.TenantID,
		UserID:             claims.UserID,
		Provider:           "stripe",
		ProviderSessionID:  session.ID,
		Status:             "pending",
		SourceCurrency:     "usd",
		SourceAmount:       amount,
		DestinationAsset:   "USDC",
		DestinationNetwork: "base",
		PaymentMethod:      method,
		WalletAddress:      walletAddress,
		HostedURL:          session.HostedURL,
		Metadata: map[string]interface{}{
			"stripeClientSecret": session.ClientSecret,
		},
	})

	if walletAddress != "" || policy.FundingMethod == "" {
		policy.WalletAddress = walletAddress
		if policy.FundingMethod == "" {
			policy.FundingMethod = "stripe_onramp"
		}
		a.store.UpsertPaymentPolicy(policy)
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"topup": topup,
		"stripe": map[string]interface{}{
			"sessionId":       session.ID,
			"clientSecret":    session.ClientSecret,
			"hostedUrl":       session.HostedURL,
			"configured":      stripe.configured(),
			"minimumTopUpUsd": stripe.minTopupUSD,
		},
	})
}

func (a *App) handleStripeOnrampWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook payload"})
		return
	}
	if err := a.currentStripeOnrampService().verifyWebhookSignature(payload, r.Header.Get("Stripe-Signature")); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	event := map[string]interface{}{}
	if err := json.Unmarshal(payload, &event); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook json"})
		return
	}
	eventID := strings.TrimSpace(stringFromAny(event["id"]))
	eventType := strings.ToLower(strings.TrimSpace(stringFromAny(event["type"])))
	data, _ := event["data"].(map[string]interface{})
	object, _ := data["object"].(map[string]interface{})
	if len(object) == 0 {
		writeJSON(w, http.StatusAccepted, map[string]interface{}{"received": true, "ignored": "missing object"})
		return
	}
	sessionID := strings.TrimSpace(stringFromAny(object["id"]))
	if sessionID == "" {
		writeJSON(w, http.StatusAccepted, map[string]interface{}{"received": true, "ignored": "missing session id"})
		return
	}

	topup, ok := a.store.GetWalletTopUpByProviderSession("stripe", sessionID)
	if !ok {
		writeJSON(w, http.StatusAccepted, map[string]interface{}{"received": true, "ignored": "session not tracked"})
		return
	}
	if topup.Status == "fulfilled" {
		writeJSON(w, http.StatusOK, map[string]interface{}{"received": true, "idempotent": true})
		return
	}

	topup.ProviderEventID = eventID
	topup.ProviderSessionID = sessionID
	status := strings.ToLower(strings.TrimSpace(stringFromAny(object["status"])))
	if status != "" {
		topup.Status = status
	}

	txDetails, _ := object["transaction_details"].(map[string]interface{})
	if topup.SourceCurrency == "" {
		topup.SourceCurrency = nonEmpty(stringFromAny(txDetails["source_currency"]), topup.SourceCurrency)
	}
	if topup.SourceAmount <= 0 {
		topup.SourceAmount = firstPositive(floatFromAny(txDetails["source_amount"]), topup.SourceAmount)
	}
	if topup.DestinationAsset == "" {
		topup.DestinationAsset = nonEmpty(stringFromAny(txDetails["destination_currency"]), topup.DestinationAsset)
	}
	if topup.DestinationNetwork == "" {
		topup.DestinationNetwork = nonEmpty(stringFromAny(txDetails["destination_network"]), topup.DestinationNetwork)
	}

	shouldCredit := status == "fulfillment_complete" || strings.Contains(eventType, "fulfillment_complete")
	if shouldCredit {
		destAmount := topup.DestinationAmount
		if destAmount <= 0 {
			destAmount = firstPositive(
				floatFromAny(txDetails["destination_amount"]),
				floatFromAny(txDetails["destination_amount_decimal"]),
				floatFromAny(txDetails["destination_exchange_amount"]),
				floatFromAny(object["destination_amount"]),
				floatFromAny(object["destination_amount_decimal"]),
				floatFromAny(object["destination_exchange_amount"]),
				topup.SourceAmount,
			)
		}
		if destAmount <= 0 {
			topup.Status = "failed"
			topup.FailureReason = "fulfilled event missing destination amount"
			_ = a.store.UpdateWalletTopUp(topup)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "destination amount missing"})
			return
		}
		if _, err := a.applyWalletCredit(topup.TenantID, topup.UserID, destAmount); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		topup.Status = "fulfilled"
		topup.DestinationAmount = destAmount
		topup.FulfilledAt = time.Now().UTC()
		_ = a.store.UpdateWalletTopUp(topup)
		if _, err := a.postTopUpAccounting(topup); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"received": true, "creditedUsdc": destAmount})
		return
	}

	if status == "rejected" || strings.Contains(eventType, "rejected") || strings.Contains(eventType, "failed") {
		topup.Status = "failed"
		topup.FailureReason = nonEmpty(stringFromAny(object["failure_reason"]), stringFromAny(object["rejection_reason"]))
		_ = a.store.UpdateWalletTopUp(topup)
		writeJSON(w, http.StatusOK, map[string]interface{}{"received": true, "status": "failed"})
		return
	}

	_ = a.store.UpdateWalletTopUp(topup)
	writeJSON(w, http.StatusOK, map[string]interface{}{"received": true, "status": topup.Status})
}

func floatFromAny(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		if !math.IsNaN(val) && !math.IsInf(val, 0) {
			return val
		}
	case string:
		text := strings.TrimSpace(val)
		if text == "" {
			return 0
		}
		out, err := strconv.ParseFloat(text, 64)
		if err == nil && !math.IsNaN(out) && !math.IsInf(out, 0) {
			return out
		}
	}
	return 0
}

func firstPositive(values ...float64) float64 {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 0
}

func firstForwardedIP(raw string) string {
	parts := strings.Split(raw, ",")
	if len(parts) == 0 {
		return strings.TrimSpace(raw)
	}
	return strings.TrimSpace(parts[0])
}
