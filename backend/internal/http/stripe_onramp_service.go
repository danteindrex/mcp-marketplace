package http

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
)

type stripeOnrampService struct {
	secretKey   string
	webhookKey  string
	allowMock   bool
	returnURL   string
	refreshURL  string
	minTopupUSD float64
	defaultUSD  float64
	baseURL     string
	client      *http.Client
}

type stripeCreateSessionInput struct {
	AmountUSD     float64
	WalletAddress string
	CustomerIP    string
	TenantID      string
	UserID        string
}

type stripeOnrampSession struct {
	ID           string
	ClientSecret string
	HostedURL    string
	Raw          map[string]interface{}
}

func newStripeOnrampService(cfg config.Config) *stripeOnrampService {
	return &stripeOnrampService{
		secretKey:   strings.TrimSpace(cfg.StripeSecretKey),
		webhookKey:  strings.TrimSpace(cfg.StripeWebhookSecret),
		allowMock:   cfg.AllowInsecureDefaults,
		returnURL:   strings.TrimSpace(cfg.StripeOnrampReturnURL),
		refreshURL:  strings.TrimSpace(cfg.StripeOnrampRefreshURL),
		minTopupUSD: cfg.StripeOnrampMinUSD,
		defaultUSD:  cfg.StripeOnrampDefaultUSD,
		baseURL:     "https://api.stripe.com",
		client: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

func (s *stripeOnrampService) configured() bool {
	if strings.TrimSpace(s.secretKey) == "" {
		return false
	}
	if s.allowMock {
		return true
	}
	return strings.TrimSpace(s.webhookKey) != ""
}

func (s *stripeOnrampService) createSession(ctx context.Context, in stripeCreateSessionInput) (stripeOnrampSession, error) {
	amount := in.AmountUSD
	if amount <= 0 {
		amount = s.defaultUSD
	}
	if amount < s.minTopupUSD {
		return stripeOnrampSession{}, fmt.Errorf("minimum top-up amount is %.2f USD", s.minTopupUSD)
	}

	if !s.configured() {
		if !s.allowMock {
			if strings.TrimSpace(s.secretKey) == "" {
				return stripeOnrampSession{}, fmt.Errorf("stripe onramp is not configured")
			}
			return stripeOnrampSession{}, fmt.Errorf("stripe onramp requires STRIPE_WEBHOOK_SECRET in non-dev mode")
		}
		mockID := "onramp_mock_" + hashAny(map[string]interface{}{
			"tenantId": in.TenantID,
			"userId":   in.UserID,
			"amount":   amount,
			"at":       time.Now().UTC().Format(time.RFC3339Nano),
		})[:16]
		return stripeOnrampSession{
			ID:           mockID,
			ClientSecret: "mock_client_secret_" + mockID,
			HostedURL:    "",
			Raw: map[string]interface{}{
				"id":            mockID,
				"object":        "crypto.onramp_session",
				"status":        "requires_payment",
				"source_amount": fmt.Sprintf("%.2f", amount),
				"mode":          "mock",
			},
		}, nil
	}

	form := url.Values{}
	form.Set("source_currency", "usd")
	form.Set("source_amount", fmt.Sprintf("%.2f", amount))
	form.Set("destination_currencies[0]", "usdc")
	form.Set("destination_networks[0]", "base")
	if strings.TrimSpace(in.WalletAddress) != "" {
		// Base is EVM-compatible; Stripe's wallet_addresses key uses ethereum for EVM addresses.
		form.Set("wallet_addresses[ethereum]", strings.TrimSpace(in.WalletAddress))
		form.Set("lock_wallet_address", "true")
	}
	if strings.TrimSpace(in.CustomerIP) != "" {
		form.Set("customer_ip_address", strings.TrimSpace(in.CustomerIP))
	}
	form.Set("metadata[tenant_id]", in.TenantID)
	form.Set("metadata[user_id]", in.UserID)

	resp, err := s.postForm(ctx, s.baseURL+"/v1/crypto/onramp_sessions", form)
	if err != nil {
		return stripeOnrampSession{}, err
	}
	out := stripeOnrampSession{
		ID:           stringFromAny(resp["id"]),
		ClientSecret: stringFromAny(resp["client_secret"]),
		HostedURL:    stringFromAny(resp["hosted_url"]),
		Raw:          resp,
	}
	if out.ID == "" {
		return stripeOnrampSession{}, fmt.Errorf("stripe onramp session missing id")
	}
	return out, nil
}

func (s *stripeOnrampService) verifyWebhookSignature(payload []byte, signatureHeader string) error {
	if strings.TrimSpace(s.webhookKey) == "" {
		if !s.allowMock {
			return fmt.Errorf("stripe webhook secret is required in non-dev mode")
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
	mac := hmac.New(sha256.New, []byte(s.webhookKey))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	for _, sig := range signatures {
		if subtle.ConstantTimeCompare([]byte(expected), []byte(sig)) == 1 {
			return nil
		}
	}
	return fmt.Errorf("stripe signature mismatch")
}

func parseStripeSignature(header string) (string, []string) {
	parts := strings.Split(header, ",")
	ts := ""
	sigs := make([]string, 0)
	for _, raw := range parts {
		part := strings.TrimSpace(raw)
		if strings.HasPrefix(part, "t=") {
			ts = strings.TrimSpace(strings.TrimPrefix(part, "t="))
			continue
		}
		if strings.HasPrefix(part, "v1=") {
			v := strings.TrimSpace(strings.TrimPrefix(part, "v1="))
			if v != "" {
				sigs = append(sigs, v)
			}
		}
	}
	return ts, sigs
}

func (s *stripeOnrampService) postForm(ctx context.Context, endpoint string, form url.Values) (map[string]interface{}, error) {
	body := bytes.NewBufferString(form.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+s.secretKey)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeStripeResponse(resp)
}

func decodeStripeResponse(resp *http.Response) (map[string]interface{}, error) {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	out := map[string]interface{}{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, err
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if msg := stringFromAny(out["error"]); msg != "" {
			return nil, fmt.Errorf("stripe onramp error: %s", msg)
		}
		if errObj, ok := out["error"].(map[string]interface{}); ok {
			if msg := stringFromAny(errObj["message"]); msg != "" {
				return nil, fmt.Errorf("stripe onramp error: %s", msg)
			}
		}
		return nil, fmt.Errorf("stripe onramp status %d", resp.StatusCode)
	}
	return out, nil
}
