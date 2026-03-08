package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
)

type stripeConnectService struct {
	secretKey  string
	webhookKey string
	allowMock  bool
	returnURL  string
	refreshURL string
	baseURL    string
	client     *http.Client
}

type stripeConnectAccountSnapshot struct {
	AccountID           string
	DetailsSubmitted    bool
	ChargesEnabled      bool
	PayoutsEnabled      bool
	CurrentlyDue        []string
	EventuallyDue       []string
	PendingVerification []string
	DisabledReason      string
	Raw                 map[string]interface{}
}

func newStripeConnectService(cfg config.Config) *stripeConnectService {
	return &stripeConnectService{
		secretKey:  strings.TrimSpace(cfg.StripeSecretKey),
		webhookKey: strings.TrimSpace(cfg.StripeConnectWebhookSecret),
		allowMock:  cfg.AllowInsecureDefaults,
		returnURL:  strings.TrimSpace(cfg.StripeConnectReturnURL),
		refreshURL: strings.TrimSpace(cfg.StripeConnectRefreshURL),
		baseURL:    "https://api.stripe.com",
		client: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

func (s *stripeConnectService) configured() bool {
	if strings.TrimSpace(s.secretKey) == "" {
		return false
	}
	if s.allowMock {
		return true
	}
	return strings.TrimSpace(s.webhookKey) != ""
}

func (s *stripeConnectService) createExpressAccount(ctx context.Context, tenantID string) (string, map[string]interface{}, error) {
	if !s.configured() {
		if !s.allowMock {
			if strings.TrimSpace(s.secretKey) == "" {
				return "", nil, fmt.Errorf("stripe connect is not configured")
			}
			return "", nil, fmt.Errorf("stripe connect requires STRIPE_CONNECT_WEBHOOK_SECRET in non-dev mode")
		}
		mockID := "acct_mock_" + hashAny(map[string]interface{}{
			"tenantId": tenantID,
			"at":       time.Now().UTC().Format(time.RFC3339Nano),
		})[:16]
		return mockID, map[string]interface{}{
			"id":                mockID,
			"type":              "express",
			"charges_enabled":   false,
			"payouts_enabled":   false,
			"details_submitted": false,
			"requirements": map[string]interface{}{
				"currently_due":  []interface{}{"business_profile.url", "individual.verification.document"},
				"eventually_due": []interface{}{},
			},
		}, nil
	}

	form := url.Values{}
	form.Set("type", "express")
	form.Set("metadata[tenant_id]", strings.TrimSpace(tenantID))
	form.Set("capabilities[card_payments][requested]", "true")
	form.Set("capabilities[transfers][requested]", "true")
	resp, err := s.postForm(ctx, s.baseURL+"/v1/accounts", "", form)
	if err != nil {
		return "", nil, err
	}
	accountID := strings.TrimSpace(stringFromAny(resp["id"]))
	if accountID == "" {
		return "", nil, fmt.Errorf("stripe connect account missing id")
	}
	return accountID, resp, nil
}

func (s *stripeConnectService) createOnboardingLink(ctx context.Context, accountID string) (string, error) {
	if strings.TrimSpace(accountID) == "" {
		return "", fmt.Errorf("stripe account id is required")
	}
	if !s.configured() {
		if !s.allowMock {
			if strings.TrimSpace(s.secretKey) == "" {
				return "", fmt.Errorf("stripe connect is not configured")
			}
			return "", fmt.Errorf("stripe connect requires STRIPE_CONNECT_WEBHOOK_SECRET in non-dev mode")
		}
		return "https://dashboard.stripe.com/connect/onboarding/" + accountID, nil
	}
	form := url.Values{}
	form.Set("account", accountID)
	form.Set("type", "account_onboarding")
	form.Set("refresh_url", s.effectiveRefreshURL())
	form.Set("return_url", s.effectiveReturnURL())
	resp, err := s.postForm(ctx, s.baseURL+"/v1/account_links", "", form)
	if err != nil {
		return "", err
	}
	link := strings.TrimSpace(stringFromAny(resp["url"]))
	if link == "" {
		return "", fmt.Errorf("stripe connect onboarding link missing url")
	}
	return link, nil
}

func (s *stripeConnectService) fetchAccount(ctx context.Context, accountID string) (stripeConnectAccountSnapshot, error) {
	if strings.TrimSpace(accountID) == "" {
		return stripeConnectAccountSnapshot{}, fmt.Errorf("stripe account id is required")
	}
	if !s.configured() {
		if !s.allowMock {
			if strings.TrimSpace(s.secretKey) == "" {
				return stripeConnectAccountSnapshot{}, fmt.Errorf("stripe connect is not configured")
			}
			return stripeConnectAccountSnapshot{}, fmt.Errorf("stripe connect requires STRIPE_CONNECT_WEBHOOK_SECRET in non-dev mode")
		}
		return stripeConnectAccountSnapshot{
			AccountID:           accountID,
			DetailsSubmitted:    false,
			ChargesEnabled:      false,
			PayoutsEnabled:      false,
			CurrentlyDue:        []string{"individual.verification.document"},
			EventuallyDue:       []string{},
			PendingVerification: []string{},
			DisabledReason:      "",
			Raw: map[string]interface{}{
				"id": accountID,
			},
		}, nil
	}
	resp, err := s.getJSON(ctx, s.baseURL+"/v1/accounts/"+url.PathEscape(strings.TrimSpace(accountID)), "")
	if err != nil {
		return stripeConnectAccountSnapshot{}, err
	}
	return connectSnapshotFromObject(resp), nil
}

func (s *stripeConnectService) createPayout(ctx context.Context, accountID string, amountUSDC float64, metadata map[string]string) (string, float64, error) {
	if strings.TrimSpace(accountID) == "" {
		return "", 0, fmt.Errorf("stripe account id is required")
	}
	if amountUSDC <= 0 {
		return "", 0, fmt.Errorf("payout amount must be positive")
	}
	if !s.configured() {
		if !s.allowMock {
			if strings.TrimSpace(s.secretKey) == "" {
				return "", 0, fmt.Errorf("stripe connect is not configured")
			}
			return "", 0, fmt.Errorf("stripe connect requires STRIPE_CONNECT_WEBHOOK_SECRET in non-dev mode")
		}
		ref := "po_mock_" + hashAny(map[string]interface{}{
			"accountId": accountID,
			"amount":    amountUSDC,
			"at":        time.Now().UTC().Format(time.RFC3339Nano),
		})[:16]
		return ref, 0, nil
	}
	amountCents := int(math.Round(amountUSDC * 100))
	if amountCents <= 0 {
		return "", 0, fmt.Errorf("payout amount is too small")
	}
	form := url.Values{}
	form.Set("amount", strconv.Itoa(amountCents))
	form.Set("currency", "usd")
	for k, v := range metadata {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		form.Set("metadata["+k+"]", v)
	}
	resp, err := s.postForm(ctx, s.baseURL+"/v1/payouts", accountID, form)
	if err != nil {
		return "", 0, err
	}
	ref := strings.TrimSpace(stringFromAny(resp["id"]))
	if ref == "" {
		return "", 0, fmt.Errorf("stripe payout missing id")
	}
	return ref, 0, nil
}

func (s *stripeConnectService) verifyWebhookSignature(payload []byte, signatureHeader string) error {
	if strings.TrimSpace(s.webhookKey) == "" {
		if !s.allowMock {
			return fmt.Errorf("stripe connect webhook secret is required in non-dev mode")
		}
		return nil
	}
	timestamp, signatures := parseStripeSignature(signatureHeader)
	if timestamp == "" || len(signatures) == 0 {
		return fmt.Errorf("invalid stripe signature header")
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid stripe signature timestamp")
	}
	eventTime := time.Unix(ts, 0)
	now := time.Now().UTC()
	if eventTime.Before(now.Add(-5*time.Minute)) || eventTime.After(now.Add(5*time.Minute)) {
		return fmt.Errorf("stale stripe signature timestamp")
	}
	// Reuse the same HMAC verification routine as Onramp signatures.
	return (&stripeOnrampService{webhookKey: s.webhookKey}).verifyWebhookSignature(payload, signatureHeader)
}

func (s *stripeConnectService) effectiveReturnURL() string {
	if strings.TrimSpace(s.returnURL) != "" {
		return s.returnURL
	}
	return "http://localhost:3000/merchant/revenue?connect=return"
}

func (s *stripeConnectService) effectiveRefreshURL() string {
	if strings.TrimSpace(s.refreshURL) != "" {
		return s.refreshURL
	}
	return "http://localhost:3000/merchant/revenue?connect=refresh"
}

func connectSnapshotFromObject(obj map[string]interface{}) stripeConnectAccountSnapshot {
	req, _ := obj["requirements"].(map[string]interface{})
	return stripeConnectAccountSnapshot{
		AccountID:           strings.TrimSpace(stringFromAny(obj["id"])),
		DetailsSubmitted:    boolFromAny(obj["details_submitted"]),
		ChargesEnabled:      boolFromAny(obj["charges_enabled"]),
		PayoutsEnabled:      boolFromAny(obj["payouts_enabled"]),
		CurrentlyDue:        stringSliceFromAny(req["currently_due"]),
		EventuallyDue:       stringSliceFromAny(req["eventually_due"]),
		PendingVerification: stringSliceFromAny(req["pending_verification"]),
		DisabledReason:      strings.TrimSpace(stringFromAny(req["disabled_reason"])),
		Raw:                 obj,
	}
}

func (s *stripeConnectService) postForm(ctx context.Context, endpoint, connectedAccount string, form url.Values) (map[string]interface{}, error) {
	body := strings.NewReader(form.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+s.secretKey)
	if strings.TrimSpace(connectedAccount) != "" {
		req.Header.Set("Stripe-Account", strings.TrimSpace(connectedAccount))
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeStripeResponse(resp)
}

func (s *stripeConnectService) getJSON(ctx context.Context, endpoint, connectedAccount string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.secretKey)
	if strings.TrimSpace(connectedAccount) != "" {
		req.Header.Set("Stripe-Account", strings.TrimSpace(connectedAccount))
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	parsed := map[string]interface{}{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return nil, err
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if errObj, ok := parsed["error"].(map[string]interface{}); ok {
			if msg := strings.TrimSpace(stringFromAny(errObj["message"])); msg != "" {
				return nil, fmt.Errorf("stripe connect error: %s", msg)
			}
		}
		return nil, fmt.Errorf("stripe connect status %d", resp.StatusCode)
	}
	return parsed, nil
}

func boolFromAny(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "1", "true", "yes", "on":
			return true
		}
	case float64:
		return val != 0
	}
	return false
}

func stringSliceFromAny(v interface{}) []string {
	arr, ok := v.([]interface{})
	if !ok || len(arr) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s := strings.TrimSpace(stringFromAny(item))
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}
