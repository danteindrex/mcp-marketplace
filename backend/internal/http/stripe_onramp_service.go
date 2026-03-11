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
	secretKey     string
	webhookKey    string
	returnURL     string
	refreshURL    string
	minTopupUSD   float64
	defaultUSD    float64
	baseURL       string
	client        *http.Client
	allowInsecure bool
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

type stripeAPIError struct {
	Provider string
	Status   int
	Message  string
}

func (e *stripeAPIError) Error() string {
	msg := strings.TrimSpace(e.Message)
	if msg == "" {
		msg = http.StatusText(e.Status)
	}
	if strings.TrimSpace(e.Provider) == "" {
		return msg
	}
	return e.Provider + ": " + msg
}

func newStripeOnrampService(cfg config.Config) *stripeOnrampService {
	return &stripeOnrampService{
		secretKey:   strings.TrimSpace(cfg.StripeSecretKey),
		webhookKey:  strings.TrimSpace(cfg.StripeWebhookSecret),
		returnURL:   strings.TrimSpace(cfg.StripeOnrampReturnURL),
		refreshURL:  strings.TrimSpace(cfg.StripeOnrampRefreshURL),
		minTopupUSD: cfg.StripeOnrampMinUSD,
		defaultUSD:  cfg.StripeOnrampDefaultUSD,
		baseURL:     "https://api.stripe.com",
		client: &http.Client{
			Timeout: 12 * time.Second,
		},
		allowInsecure: cfg.AllowInsecureDefaults,
	}
}

func (s *stripeOnrampService) configured() bool {
	return strings.TrimSpace(s.secretKey) != ""
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
		if s.allowInsecure {
			mockID := "onramp_test_" + randomToken("test_")
			return stripeOnrampSession{
				ID:           mockID,
				ClientSecret: "test_secret_" + mockID,
				HostedURL:    "https://buy.stripe.com/test/" + mockID,
				Raw: map[string]interface{}{
					"id":            mockID,
					"client_secret": "test_secret_" + mockID,
					"url":           "https://buy.stripe.com/test/" + mockID,
					"mode":          "test",
				},
			}, nil
		}
		return stripeOnrampSession{}, fmt.Errorf("stripe onramp is not configured")
	}

	form := url.Values{}
	form.Set("source_currency", "usd")
	form.Set("source_amount", fmt.Sprintf("%.2f", amount))
	form.Add("destination_currencies[]", "usdc")
	form.Add("destination_networks[]", "base")
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
		HostedURL:    stringFromAny(resp["redirect_url"]),
		Raw:          resp,
	}
	if out.HostedURL == "" {
		out.HostedURL = stringFromAny(resp["hosted_url"])
	}
	if out.ID == "" {
		return stripeOnrampSession{}, fmt.Errorf("stripe onramp session missing id")
	}
	return out, nil
}

func (s *stripeOnrampService) verifyWebhookSignature(payload []byte, signatureHeader string) error {
	if s.allowInsecure {
		return nil // Skip verification in test mode
	}
	if strings.TrimSpace(s.webhookKey) == "" {
		return fmt.Errorf("stripe webhook secret is required")
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
	return decodeStripeResponse(resp, "stripe onramp")
}

func decodeStripeResponse(resp *http.Response, provider string) (map[string]interface{}, error) {
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
			return nil, &stripeAPIError{Provider: provider, Status: resp.StatusCode, Message: msg}
		}
		if errObj, ok := out["error"].(map[string]interface{}); ok {
			if msg := stringFromAny(errObj["message"]); msg != "" {
				return nil, &stripeAPIError{Provider: provider, Status: resp.StatusCode, Message: msg}
			}
		}
		return nil, &stripeAPIError{Provider: provider, Status: resp.StatusCode, Message: strings.TrimSpace(string(raw))}
	}
	return out, nil
}
