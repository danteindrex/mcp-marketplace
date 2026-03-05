package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type createX402IntentRequest struct {
	ServerID string  `json:"serverId"`
	ToolName string  `json:"toolName"`
	Amount   float64 `json:"amount"`
}

func (a *App) createX402Intent(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	var req createX402IntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ServerID == "" || req.ToolName == "" || req.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if _, ok := a.store.GetServerByID(req.ServerID); !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}
	intent := a.store.CreateX402Intent(models.X402Intent{
		TenantID:   claims.TenantID,
		UserID:     claims.UserID,
		ServerID:   req.ServerID,
		ToolName:   req.ToolName,
		AmountUSDC: req.Amount,
		Network:    "base",
		Asset:      "USDC",
	})
	w.Header().Set("PAYMENT-REQUIRED", "x402v2")
	w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope"`)
	writeJSON(w, http.StatusPaymentRequired, intent)
}

func (a *App) settleX402Intent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intent, ok := a.store.SettleX402Intent(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "intent not found"})
		return
	}
	writeJSON(w, http.StatusOK, intent)
}

func (a *App) listX402Intents(w http.ResponseWriter, r *http.Request) {
	claims, _ := getClaims(r.Context())
	intents := a.store.ListX402Intents(claims.TenantID, claims.UserID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": intents, "count": len(intents)})
}
