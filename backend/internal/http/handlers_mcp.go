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

type mcpJSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type mcpJSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type mcpJSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      interface{}      `json:"id,omitempty"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *mcpJSONRPCError `json:"error,omitempty"`
}

func (a *App) mcpHub(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	tenantID := chi.URLParam(r, "tenantID")
	userID := chi.URLParam(r, "userID")
	if claims.Role != models.RoleAdmin && (claims.TenantID != tenantID || claims.UserID != userID) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	hub := a.ensureHubProfile(tenantID, userID)
	routes := a.store.ListHubRoutes(hub.ID)
	tools := mcpToolsFromRoutes(routes)

	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"protocol":  "mcp",
			"transport": "streamable-http",
			"resource":  hub.HubURL,
			"hub":       hub,
			"tools":     tools,
		})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req mcpJSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json-rpc request"})
		return
	}

	switch req.Method {
	case "initialize":
		writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2025-11-05",
				"serverInfo": map[string]string{
					"name":    "mcp-marketplace-hub",
					"version": "1.0.0",
				},
				"capabilities": map[string]interface{}{
					"tools": map[string]bool{
						"listChanged": false,
					},
				},
			},
		})
	case "tools/list":
		writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": tools,
			},
		})
	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
			Meta      map[string]interface{} `json:"_meta"`
		}
		if len(req.Params) > 0 {
			_ = json.Unmarshal(req.Params, &params)
		}
		var targetRoute models.HubRoute
		toolExists := false
		for _, route := range routes {
			if !route.Enabled {
				continue
			}
			if route.ToolName == params.Name {
				targetRoute = route
				toolExists = true
				break
			}
		}
		if !toolExists {
			writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &mcpJSONRPCError{
					Code:    -32601,
					Message: "tool not found",
				},
			})
			return
		}
		server, ok := a.store.GetServerByID(targetRoute.ServerID)
		if !ok {
			writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &mcpJSONRPCError{
					Code:    -32603,
					Message: "upstream server config missing",
				},
			})
			return
		}
		resultMeta := map[string]interface{}{}

		if server.PricingType == "x402" {
			policy := a.effectivePaymentPolicy(claims.TenantID, claims.UserID)
			intents := a.store.ListX402Intents(claims.TenantID, claims.UserID)
			credit, creditOK, err := a.consumeSettledCredit(intents, server.ID, params.Name)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}

			if !creditOK {
				if err := a.validateCaps(policy, server, intents, server.PricingAmount); err != nil {
					writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
						JSONRPC: "2.0",
						ID:      req.ID,
						Error: &mcpJSONRPCError{
							Code:    -32002,
							Message: "payment cap exceeded",
							Data: map[string]interface{}{
								"caps":  policy,
								"error": err.Error(),
							},
						},
					})
					return
				}

				paymentResponse := extractPaymentResponse(params.Meta)
				idempotencyKey := extractIdempotencyKey(req.ID, params.Meta, params.Name)
				selectedMethod := nonEmpty(stringFromAny(params.Meta["x402/payment-method"]), firstMethod(policy.AllowedMethods))
				requirement := buildX402Requirement(
					server,
					params.Name,
					server.PricingAmount,
					nonEmpty(server.CanonicalResourceURI, hub.HubURL),
					selectedMethod,
					idempotencyKey,
				)

				if paymentResponse == nil && selectedMethod == "wallet_balance" {
					walletPolicy, walletPaymentID, derr := a.applyWalletDebit(claims.TenantID, claims.UserID, server.PricingAmount, idempotencyKey)
					if derr != nil {
						writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
							JSONRPC: "2.0",
							ID:      req.ID,
							Error: &mcpJSONRPCError{
								Code:    -32002,
								Message: "wallet funding required",
								Data: map[string]interface{}{
									"error": derr.Error(),
									"wallet": map[string]interface{}{
										"balanceUsdc":        walletPolicy.WalletBalanceUSDC,
										"minimumBalanceUsdc": walletPolicy.MinimumBalanceUSDC,
										"hardStopOnLowFunds": walletPolicy.HardStopOnLowFunds,
									},
									"_meta": map[string]interface{}{
										"x402/payment-requirements": []map[string]interface{}{requirement},
									},
								},
							},
						})
						return
					}
					challengeBytes, _ := json.Marshal([]map[string]interface{}{requirement})
					pending := a.store.CreateX402Intent(models.X402Intent{
						TenantID:           claims.TenantID,
						UserID:             claims.UserID,
						ServerID:           server.ID,
						ToolName:           params.Name,
						AmountUSDC:         server.PricingAmount,
						Network:            "base",
						Asset:              "USDC",
						Challenge:          string(challengeBytes),
						PaymentMethod:      "wallet_balance",
						PaymentIdentifier:  walletPaymentID,
						IdempotencyKey:     idempotencyKey,
						X402Version:        "2",
						Resource:           nonEmpty(server.CanonicalResourceURI, hub.HubURL),
						VerificationStatus: "verified",
						VerificationNote:   "debited from prepaid wallet balance",
						Quantity:           1,
						RemainingQuantity:  1,
						RequestFingerprint: hashAny(map[string]interface{}{"tool": params.Name, "idempotencyKey": idempotencyKey, "requestID": req.ID}),
					})
					settled, _ := a.store.SettleX402Intent(pending.ID)
					settled.PaymentIdentifier = walletPaymentID
					settled.PaymentMethod = "wallet_balance"
					settled.VerificationStatus = "verified"
					settled.VerificationNote = "debited from prepaid wallet balance"
					if settled.RemainingQuantity <= 0 {
						settled.RemainingQuantity = 1
					}
					settled.RemainingQuantity--
					_ = a.store.UpdateX402Intent(settled)
					a.ensurePaidEntitlement(settled)
					credit = settled
					creditOK = true
					resultMeta["wallet/balance-usdc"] = walletPolicy.WalletBalanceUSDC
				}

				if !creditOK && paymentResponse == nil {
					challengeBytes, _ := json.Marshal([]map[string]interface{}{requirement})
					intent := a.store.CreateX402Intent(models.X402Intent{
						TenantID:           claims.TenantID,
						UserID:             claims.UserID,
						ServerID:           server.ID,
						ToolName:           params.Name,
						AmountUSDC:         server.PricingAmount,
						Network:            "base",
						Asset:              "USDC",
						Challenge:          string(challengeBytes),
						PaymentMethod:      selectedMethod,
						IdempotencyKey:     idempotencyKey,
						X402Version:        "2",
						Resource:           nonEmpty(server.CanonicalResourceURI, hub.HubURL),
						VerificationStatus: "pending",
						Quantity:           1,
						RemainingQuantity:  0,
						RequestFingerprint: hashAny(map[string]interface{}{"tool": params.Name, "idempotencyKey": idempotencyKey, "requestID": req.ID}),
					})
					w.Header().Set("PAYMENT-REQUIRED", string(challengeBytes))
					writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
						JSONRPC: "2.0",
						ID:      req.ID,
						Error: &mcpJSONRPCError{
							Code:    -32002,
							Message: "payment required",
							Data: map[string]interface{}{
								"intentId": intent.ID,
								"_meta": map[string]interface{}{
									"x402/payment-requirements": []map[string]interface{}{requirement},
								},
							},
						},
					})
					return
				}

				paymentIdentifier := strings.TrimSpace(stringFromAny(paymentResponse["paymentIdentifier"]))
				if paymentIdentifier == "" {
					paymentIdentifier = strings.TrimSpace(stringFromAny(paymentResponse["id"]))
				}
				if paymentIdentifier != "" {
					existingSettled := a.findSettledByPaymentIdentifier(intents, paymentIdentifier)
					if existingSettled != nil && existingSettled.RemainingQuantity > 0 {
						existingSettled.RemainingQuantity--
						_ = a.store.UpdateX402Intent(*existingSettled)
						credit = *existingSettled
						creditOK = true
					}
				}

				if !creditOK {
					verifyRes, verr := a.x402.verifyAndSettle(r.Context(), requirement, paymentResponse)
					if verr != nil {
						writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
							JSONRPC: "2.0",
							ID:      req.ID,
							Error: &mcpJSONRPCError{
								Code:    -32002,
								Message: "payment verification failed",
								Data: map[string]interface{}{
									"error": verr.Error(),
									"_meta": map[string]interface{}{
										"x402/payment-requirements": []map[string]interface{}{requirement},
									},
								},
							},
						})
						return
					}
					if !verifyRes.Valid {
						writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
							JSONRPC: "2.0",
							ID:      req.ID,
							Error: &mcpJSONRPCError{
								Code:    -32002,
								Message: "payment verification rejected",
								Data:    map[string]interface{}{"error": verifyRes.Note},
							},
						})
						return
					}
					if verifyRes.PaymentIdentifier != "" {
						if existing := a.findSettledByPaymentIdentifier(intents, verifyRes.PaymentIdentifier); existing != nil {
							if existing.RemainingQuantity > 0 {
								existing.RemainingQuantity--
								_ = a.store.UpdateX402Intent(*existing)
								credit = *existing
								creditOK = true
							} else {
								writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
									JSONRPC: "2.0",
									ID:      req.ID,
									Error: &mcpJSONRPCError{
										Code:    -32002,
										Message: "payment replay detected with exhausted credits",
									},
								})
								return
							}
						}
					}
					if !creditOK {
						challengeBytes, _ := json.Marshal([]map[string]interface{}{requirement})
						pending := a.store.CreateX402Intent(models.X402Intent{
							TenantID:           claims.TenantID,
							UserID:             claims.UserID,
							ServerID:           server.ID,
							ToolName:           params.Name,
							AmountUSDC:         server.PricingAmount,
							Network:            nonEmpty(verifyRes.Network, "base"),
							Asset:              nonEmpty(verifyRes.Asset, "USDC"),
							Challenge:          string(challengeBytes),
							PaymentMethod:      nonEmpty(verifyRes.Method, "x402_wallet"),
							PaymentIdentifier:  verifyRes.PaymentIdentifier,
							IdempotencyKey:     idempotencyKey,
							X402Version:        "2",
							Resource:           nonEmpty(server.CanonicalResourceURI, hub.HubURL),
							VerificationStatus: "verified",
							VerificationNote:   verifyRes.Note,
							FacilitatorTx:      verifyRes.TxHash,
							Quantity:           1,
							RemainingQuantity:  1,
							RequestFingerprint: hashAny(map[string]interface{}{"tool": params.Name, "idempotencyKey": idempotencyKey, "requestID": req.ID}),
						})
						settled, _ := a.store.SettleX402Intent(pending.ID)
						settled.PaymentIdentifier = verifyRes.PaymentIdentifier
						settled.PaymentMethod = nonEmpty(verifyRes.Method, "x402_wallet")
						settled.VerificationStatus = "verified"
						settled.VerificationNote = verifyRes.Note
						settled.FacilitatorTx = verifyRes.TxHash
						if settled.RemainingQuantity <= 0 {
							settled.RemainingQuantity = 1
						}
						settled.RemainingQuantity--
						_ = a.store.UpdateX402Intent(settled)
						a.ensurePaidEntitlement(settled)
						credit = settled
						creditOK = true
					}
				}
			}

			if creditOK {
				w.Header().Set("PAYMENT-RESPONSE", credit.PaymentIdentifier)
				resultMeta["x402/payment-response"] = map[string]interface{}{
					"paymentIdentifier": credit.PaymentIdentifier,
					"method":            credit.PaymentMethod,
					"status":            "accepted",
					"remainingQuantity": credit.RemainingQuantity,
				}
			}
		}

		result := map[string]interface{}{
			"content": []map[string]string{
				{
					"type": "text",
					"text": "Tool executed by marketplace hub in demo mode.",
				},
			},
			"isError": false,
		}
		if len(resultMeta) > 0 {
			result["_meta"] = resultMeta
		}
		writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		})
	default:
		writeJSON(w, http.StatusOK, mcpJSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &mcpJSONRPCError{
				Code:    -32601,
				Message: "method not found",
			},
		})
	}
}

func mcpToolsFromRoutes(routes []models.HubRoute) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(routes))
	for _, route := range routes {
		if !route.Enabled {
			continue
		}
		out = append(out, map[string]interface{}{
			"name": route.ToolName,
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			"description": "Tool proxied via personal marketplace hub.",
		})
	}
	return out
}

func extractPaymentResponse(meta map[string]interface{}) map[string]interface{} {
	if len(meta) == 0 {
		return nil
	}
	raw, ok := meta["x402/payment-response"]
	if !ok || raw == nil {
		return nil
	}
	if v, ok := raw.(map[string]interface{}); ok {
		return v
	}
	if text, ok := raw.(string); ok {
		out := map[string]interface{}{}
		if err := json.Unmarshal([]byte(text), &out); err == nil {
			return out
		}
	}
	return nil
}

func extractIdempotencyKey(requestID interface{}, meta map[string]interface{}, toolName string) string {
	if len(meta) > 0 {
		if key := strings.TrimSpace(stringFromAny(meta["x402/idempotency-key"])); key != "" {
			return key
		}
	}
	if key := strings.TrimSpace(stringFromAny(requestID)); key != "" {
		return "mcp_" + key
	}
	return "mcp_" + toolName + "_" + time.Now().UTC().Format(time.RFC3339Nano)
}

func firstMethod(methods []string) string {
	priority := []string{"x402_wallet", "wallet_balance", "stripe", "coinbase_commerce"}
	for _, candidate := range priority {
		if hasScope(methods, candidate) {
			return candidate
		}
	}
	if len(methods) > 0 {
		return strings.ToLower(strings.TrimSpace(methods[0]))
	}
	return "x402_wallet"
}

func nonEmpty(a string, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func (a *App) consumeSettledCredit(intents []models.X402Intent, serverID, toolName string) (models.X402Intent, bool, error) {
	sort.Slice(intents, func(i, j int) bool {
		return intents[i].SettledAt.Before(intents[j].SettledAt)
	})
	for _, intent := range intents {
		if intent.Status != "settled" {
			continue
		}
		if intent.ServerID != serverID {
			continue
		}
		if intent.ToolName != toolName {
			continue
		}
		if intent.RemainingQuantity <= 0 {
			continue
		}
		intent.RemainingQuantity--
		if ok := a.store.UpdateX402Intent(intent); !ok {
			return models.X402Intent{}, false, errCreditUpdate
		}
		return intent, true, nil
	}
	return models.X402Intent{}, false, nil
}

var errCreditUpdate = &mcpCreditError{message: "failed to update payment credits"}

type mcpCreditError struct {
	message string
}

func (e *mcpCreditError) Error() string {
	return e.message
}

func (a *App) findSettledByPaymentIdentifier(intents []models.X402Intent, paymentID string) *models.X402Intent {
	for i := range intents {
		if intents[i].Status != "settled" {
			continue
		}
		if intents[i].PaymentIdentifier == paymentID {
			cp := intents[i]
			return &cp
		}
	}
	return nil
}
