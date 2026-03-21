package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
)

type x402VerificationResult struct {
	Valid             bool
	PaymentIdentifier string
	Method            string
	Network           string
	Asset             string
	TxHash            string
	Note              string
}

type x402Service struct {
	mode           string
	facilitatorURL string
	apiKey         string
	client         *http.Client
	allowInsecure  bool
}

func newX402Service(cfg config.Config) *x402Service {
	mode := strings.ToLower(strings.TrimSpace(cfg.X402Mode))
	// Default to facilitator mode for real payment validation
	// Set mode to "disabled" to bypass payment verification entirely
	// Insecure defaults enables test mode for development/testing
	if mode == "" {
		if cfg.AllowInsecureDefaults {
			mode = "test"
		} else {
			mode = "facilitator"
		}
	}
	return &x402Service{
		mode:           mode,
		facilitatorURL: strings.TrimRight(strings.TrimSpace(cfg.X402FacilitatorURL), "/"),
		apiKey:         strings.TrimSpace(cfg.X402FacilitatorAPIKey),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		allowInsecure: cfg.AllowInsecureDefaults,
	}
}

func (s *x402Service) verifyAndSettle(ctx context.Context, requirement map[string]interface{}, paymentResponse map[string]interface{}) (x402VerificationResult, error) {
	// Test mode - accept any payment for development/testing
	if s.mode == "test" {
		paymentID := stringFromAny(paymentResponse["paymentIdentifier"])
		if paymentID == "" {
			paymentID = stringFromAny(paymentResponse["id"])
		}
		if paymentID == "" {
			paymentID = "test_" + hashAny(paymentResponse)[:24]
		}
		method := stringFromAny(paymentResponse["method"])
		if method == "" {
			method = "x402_wallet"
		}
		return x402VerificationResult{
			Valid:             true,
			PaymentIdentifier: paymentID,
			Method:            method,
			Network:           stringFromAny(requirement["network"]),
			Asset:             stringFromAny(requirement["asset"]),
			TxHash:            "0xtest" + randomToken("tx_"),
			Note:              "test verification accepted",
		}, nil
	}
	// Real facilitator mode requires proper configuration
	if s.mode != "facilitator" || s.facilitatorURL == "" {
		if s.mode == "disabled" {
			paymentID := stringFromAny(paymentResponse["paymentIdentifier"])
			if paymentID == "" {
				paymentID = stringFromAny(paymentResponse["id"])
			}
			if paymentID == "" {
				paymentID = "disabled_" + hashAny(paymentResponse)[:24]
			}
			method := stringFromAny(paymentResponse["method"])
			if method == "" {
				method = "x402_wallet"
			}
			// Disabled mode - skip verification entirely (for development/testing)
			return x402VerificationResult{
				Valid:             true,
				PaymentIdentifier: paymentID,
				Method:            method,
				Network:           stringFromAny(requirement["network"]),
				Asset:             stringFromAny(requirement["asset"]),
				Note:              "verification disabled",
			}, nil
		}
		// Reject payments when facilitator mode is not properly configured
		return x402VerificationResult{}, fmt.Errorf("x402 facilitator mode not configured: X402Mode must be 'facilitator' and X402FacilitatorURL must be set")
	}

	verifyPayload := map[string]interface{}{
		"requirement": requirement,
		"response":    paymentResponse,
	}
	verifyResp, err := s.postJSON(ctx, s.facilitatorURL+"/verify", verifyPayload)
	if err != nil {
		return x402VerificationResult{}, err
	}
	if ok, _ := verifyResp["valid"].(bool); !ok {
		note, _ := verifyResp["message"].(string)
		if note == "" {
			note = "verification rejected by facilitator"
		}
		return x402VerificationResult{Valid: false, Note: note}, nil
	}

	settlePayload := map[string]interface{}{
		"requirement": requirement,
		"response":    paymentResponse,
	}
	settleResp, err := s.postJSON(ctx, s.facilitatorURL+"/settle", settlePayload)
	if err != nil {
		return x402VerificationResult{}, err
	}
	paymentID := stringFromAny(settleResp["paymentIdentifier"])
	if paymentID == "" {
		paymentID = stringFromAny(verifyResp["paymentIdentifier"])
	}
	if paymentID == "" {
		paymentID = hashAny(paymentResponse)
	}
	method := stringFromAny(paymentResponse["method"])
	if method == "" {
		method = stringFromAny(verifyResp["method"])
	}
	if method == "" {
		method = "x402_wallet"
	}
	txHash := stringFromAny(settleResp["txHash"])
	network := stringFromAny(verifyResp["network"])
	if network == "" {
		network = stringFromAny(requirement["network"])
	}
	asset := stringFromAny(verifyResp["asset"])
	if asset == "" {
		asset = stringFromAny(requirement["asset"])
	}
	note := stringFromAny(settleResp["message"])
	if note == "" {
		note = "settled via facilitator"
	}
	return x402VerificationResult{
		Valid:             true,
		PaymentIdentifier: paymentID,
		Method:            method,
		Network:           network,
		Asset:             asset,
		TxHash:            txHash,
		Note:              note,
	}, nil
}

func (s *x402Service) postJSON(ctx context.Context, url string, payload interface{}) (map[string]interface{}, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("facilitator returned status %d", resp.StatusCode)
	}
	return out, nil
}

func hashAny(v interface{}) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func stringFromAny(v interface{}) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}
