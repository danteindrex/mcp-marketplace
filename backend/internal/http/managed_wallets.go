package http

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

const (
	cdpAPIHost                 = "api.cdp.coinbase.com"
	cdpAPIBaseURL              = "https://" + cdpAPIHost
	cdpAccountsPath            = "/platform/v2/evm/accounts"
	cdpAccountSignTypedDataFmt = "/platform/v2/evm/accounts/%s/sign/typed-data"
	baseMainnetNetwork         = "eip155:8453"
	baseUSDCAddress            = "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913"
	baseUSDCName               = "USD Coin"
	baseUSDCVersion            = "2"
)

type walletProvider interface {
	provisionWallet(ctx context.Context, role, tenantID, userID string) (models.ManagedWallet, error)
	signPayment(ctx context.Context, wallet models.ManagedWallet, requirement map[string]interface{}) (map[string]interface{}, error)
}

type cdpWalletProvider struct {
	provider      string
	network       string
	asset         string
	custodyMode   string
	apiKeyID      string
	apiKeySecret  string
	walletSecret  string
	allowInsecure bool
	client        *http.Client
}

type cdpBearerClaims struct {
	URIs []string `json:"uris,omitempty"`
	jwt.RegisteredClaims
}

type cdpCreateAccountRequest struct {
	Name string `json:"name,omitempty"`
}

type cdpTypedDataDomain struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	ChainID           string `json:"chainId"`
	VerifyingContract string `json:"verifyingContract"`
}

type cdpTypedDataField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type cdpSignTypedDataRequest struct {
	Domain      cdpTypedDataDomain             `json:"domain"`
	Types       map[string][]cdpTypedDataField `json:"types"`
	PrimaryType string                         `json:"primaryType"`
	Message     map[string]interface{}         `json:"message"`
}

type fireflyWalletProvider struct {
	provider      string
	network       string
	asset         string
	custodyMode   string
	signerURL     string
	authToken     string
	keystoreDir   string
	passphrase    string
	allowInsecure bool
	client        *http.Client
}

type jsonRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type jsonRPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func newWalletProvider(cfg config.Config) walletProvider {
	provider := nonEmpty(strings.TrimSpace(cfg.WalletProvider), "cdp")
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "firefly":
		return &fireflyWalletProvider{
			provider:      "firefly",
			network:       nonEmpty(strings.TrimSpace(cfg.WalletDefaultNetwork), "base"),
			asset:         nonEmpty(strings.TrimSpace(cfg.WalletDefaultAsset), "USDC"),
			custodyMode:   nonEmpty(strings.TrimSpace(cfg.WalletCustodyMode), "provider_managed"),
			signerURL:     strings.TrimRight(strings.TrimSpace(cfg.FireflySignerURL), "/"),
			authToken:     strings.TrimSpace(cfg.FireflyAuthToken),
			keystoreDir:   strings.TrimSpace(cfg.FireflyKeystoreDir),
			passphrase:    strings.TrimSpace(cfg.FireflyKeystorePassphrase),
			allowInsecure: cfg.AllowInsecureDefaults,
			client:        &http.Client{Timeout: 15 * time.Second},
		}
	default:
		return &cdpWalletProvider{
			provider:      "cdp",
			network:       nonEmpty(strings.TrimSpace(cfg.WalletDefaultNetwork), "base"),
			asset:         nonEmpty(strings.TrimSpace(cfg.WalletDefaultAsset), "USDC"),
			custodyMode:   nonEmpty(strings.TrimSpace(cfg.WalletCustodyMode), "provider_managed"),
			apiKeyID:      strings.TrimSpace(cfg.CDPAPIKeyID),
			apiKeySecret:  strings.TrimSpace(cfg.CDPAPIKeySecret),
			walletSecret:  strings.TrimSpace(cfg.CDPWalletSecret),
			allowInsecure: cfg.AllowInsecureDefaults,
			client:        &http.Client{Timeout: 15 * time.Second},
		}
	}
}

func (a *App) currentWalletProvider() walletProvider {
	resolved := a.resolvedIntegrations()
	cfg := a.cfg
	cfg.WalletProvider = a.resolveWalletProviderName(resolved.Wallet)
	cfg.WalletManagedAutoPayEnabled = resolved.Wallet.ManagedAutoPayEnabled
	cfg.WalletLegacyPaymentModeEnabled = resolved.Wallet.LegacyPaymentModeEnabled
	cfg.WalletExternalWalletsEnabled = resolved.Wallet.ExternalWalletsEnabled
	cfg.WalletCDPEnabled = resolved.Wallet.CDPEnabled
	cfg.WalletFireflyEnabled = resolved.Wallet.FireflyEnabled
	cfg.CDPAPIKeyID = resolved.Wallet.CDPAPIKeyID
	cfg.CDPAPIKeySecret = resolved.Wallet.CDPAPIKeySecret
	cfg.CDPWalletSecret = resolved.Wallet.CDPWalletSecret
	cfg.FireflySignerURL = resolved.Wallet.FireflySignerURL
	cfg.FireflyAuthToken = resolved.Wallet.FireflyAuthToken
	cfg.FireflyKeystoreDir = resolved.Wallet.FireflyKeystoreDir
	cfg.FireflyKeystorePassphrase = resolved.Wallet.FireflyKeystorePassphrase
	cfg.WalletDefaultNetwork = resolved.Wallet.DefaultNetwork
	cfg.WalletDefaultAsset = resolved.Wallet.DefaultAsset
	cfg.WalletCustodyMode = resolved.Wallet.CustodyMode
	return newWalletProvider(cfg)
}

func (p *cdpWalletProvider) provisionWallet(ctx context.Context, role, tenantID, userID string) (models.ManagedWallet, error) {
	if !strings.EqualFold(p.provider, "cdp") {
		return models.ManagedWallet{}, fmt.Errorf("unsupported wallet provider %q", p.provider)
	}
	if !p.isConfigured() {
		if p.allowInsecure {
			return p.provisionFallbackWallet(role, tenantID, userID), nil
		}
		return models.ManagedWallet{}, fmt.Errorf("cdp wallet provider is not configured")
	}

	address, providerWalletID, err := p.liveProvisionWallet(ctx, role, tenantID, userID)
	if err != nil {
		if p.allowInsecure {
			return p.provisionFallbackWallet(role, tenantID, userID), nil
		}
		return models.ManagedWallet{}, err
	}
	if providerWalletID == "" {
		providerWalletID = address
	}

	return models.ManagedWallet{
		Provider:           "cdp",
		ProviderWalletID:   providerWalletID,
		ProviderAccountID:  nonEmpty(userID, tenantID),
		TenantID:           tenantID,
		UserID:             userID,
		Role:               strings.ToLower(strings.TrimSpace(role)),
		Network:            p.network,
		Asset:              p.asset,
		Address:            strings.TrimSpace(address),
		Status:             "active",
		CustodyMode:        p.custodyMode,
		IsDefaultFunding:   strings.EqualFold(role, "buyer"),
		IsDefaultReceiving: strings.EqualFold(role, "seller"),
		LastSyncedAt:       time.Now().UTC(),
	}, nil
}

func (p *cdpWalletProvider) signPayment(ctx context.Context, wallet models.ManagedWallet, requirement map[string]interface{}) (map[string]interface{}, error) {
	if strings.TrimSpace(wallet.Address) == "" {
		return nil, fmt.Errorf("managed wallet address is missing")
	}
	if !p.isConfigured() {
		if p.allowInsecure {
			return p.fallbackPaymentPayload(wallet, requirement), nil
		}
		return nil, fmt.Errorf("cdp wallet provider is not configured")
	}

	payload, err := p.liveSignPayment(ctx, wallet, requirement)
	if err != nil {
		if p.allowInsecure {
			return p.fallbackPaymentPayload(wallet, requirement), nil
		}
		return nil, err
	}
	return payload, nil
}

func (p *cdpWalletProvider) isConfigured() bool {
	return strings.TrimSpace(p.apiKeyID) != "" &&
		strings.TrimSpace(p.apiKeySecret) != "" &&
		strings.TrimSpace(p.walletSecret) != ""
}

func (p *cdpWalletProvider) provisionFallbackWallet(role, tenantID, userID string) models.ManagedWallet {
	seed := hashAny(map[string]interface{}{
		"provider": p.provider,
		"tenantId": tenantID,
		"userId":   userID,
		"role":     role,
		"network":  p.network,
		"asset":    p.asset,
	})
	address := "0x" + seed[:40]
	return models.ManagedWallet{
		Provider:           "cdp",
		ProviderWalletID:   "cdp_wallet_" + seed[:24],
		ProviderAccountID:  nonEmpty(userID, tenantID),
		TenantID:           tenantID,
		UserID:             userID,
		Role:               strings.ToLower(strings.TrimSpace(role)),
		Network:            p.network,
		Asset:              p.asset,
		Address:            address,
		Status:             "active",
		CustodyMode:        p.custodyMode,
		IsDefaultFunding:   strings.EqualFold(role, "buyer"),
		IsDefaultReceiving: strings.EqualFold(role, "seller"),
		LastSyncedAt:       time.Now().UTC(),
	}
}

func (p *cdpWalletProvider) fallbackPaymentPayload(wallet models.ManagedWallet, requirement map[string]interface{}) map[string]interface{} {
	requirementHash := hashAny(requirement)
	proofSeed := sha256.Sum256([]byte(wallet.Address + ":" + requirementHash + ":" + wallet.ProviderWalletID))
	accepted := p.requirementToAccepted(requirement)
	return map[string]interface{}{
		"x402Version":       2,
		"paymentIdentifier": "pay_" + requirementHash[:24],
		"method":            "x402_wallet",
		"provider":          wallet.Provider,
		"providerManaged":   true,
		"walletAddress":     wallet.Address,
		"accepted":          accepted,
		"payload": map[string]interface{}{
			"proof": "cdp:" + hex.EncodeToString(proofSeed[:]),
		},
		"resource": map[string]interface{}{
			"url":         stringFromAny(requirement["resource"]),
			"description": stringFromAny(requirement["description"]),
		},
	}
}

func (p *fireflyWalletProvider) isConfigured() bool {
	return strings.TrimSpace(p.signerURL) != "" &&
		strings.TrimSpace(p.keystoreDir) != "" &&
		strings.TrimSpace(p.passphrase) != ""
}

func (p *fireflyWalletProvider) provisionWallet(ctx context.Context, role, tenantID, userID string) (models.ManagedWallet, error) {
	if !p.isConfigured() {
		if p.allowInsecure {
			return p.provisionFallbackWallet(role, tenantID, userID), nil
		}
		return models.ManagedWallet{}, fmt.Errorf("firefly wallet provider is not configured")
	}
	if err := os.MkdirAll(p.keystoreDir, 0o700); err != nil {
		return models.ManagedWallet{}, err
	}
	ks := keystore.NewKeyStore(p.keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.NewAccount(p.passphrase)
	if err != nil {
		if p.allowInsecure {
			return p.provisionFallbackWallet(role, tenantID, userID), nil
		}
		return models.ManagedWallet{}, err
	}
	if err := p.normalizeKeystoreFiles(account.Address.Hex(), account.URL.Path); err != nil {
		return models.ManagedWallet{}, err
	}
	return models.ManagedWallet{
		Provider:           "firefly",
		ProviderWalletID:   strings.ToLower(account.Address.Hex()),
		ProviderAccountID:  nonEmpty(userID, tenantID),
		TenantID:           tenantID,
		UserID:             userID,
		Role:               strings.ToLower(strings.TrimSpace(role)),
		Network:            p.network,
		Asset:              p.asset,
		Address:            account.Address.Hex(),
		Status:             "active",
		CustodyMode:        p.custodyMode,
		IsDefaultFunding:   strings.EqualFold(role, "buyer"),
		IsDefaultReceiving: strings.EqualFold(role, "seller"),
		LastSyncedAt:       time.Now().UTC(),
	}, nil
}

func (p *fireflyWalletProvider) normalizeKeystoreFiles(address string, originalPath string) error {
	trimmed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(address)), "0x")
	if trimmed == "" || strings.TrimSpace(originalPath) == "" {
		return nil
	}
	targetKeyPath := filepath.Join(p.keystoreDir, trimmed+".key.json")
	if !strings.EqualFold(filepath.Clean(originalPath), filepath.Clean(targetKeyPath)) {
		if err := os.Rename(originalPath, targetKeyPath); err != nil {
			return err
		}
	}
	targetPasswordPath := filepath.Join(p.keystoreDir, trimmed+".password")
	return os.WriteFile(targetPasswordPath, []byte(p.passphrase), 0o600)
}

func (p *fireflyWalletProvider) signPayment(ctx context.Context, wallet models.ManagedWallet, requirement map[string]interface{}) (map[string]interface{}, error) {
	if strings.TrimSpace(wallet.Address) == "" {
		return nil, fmt.Errorf("managed wallet address is missing")
	}
	if !p.isConfigured() {
		if p.allowInsecure {
			return p.fallbackPaymentPayload(wallet, requirement), nil
		}
		return nil, fmt.Errorf("firefly wallet provider is not configured")
	}
	payload, err := p.liveSignPayment(ctx, wallet, requirement)
	if err != nil {
		if p.allowInsecure {
			return p.fallbackPaymentPayload(wallet, requirement), nil
		}
		return nil, err
	}
	return payload, nil
}

func (p *fireflyWalletProvider) provisionFallbackWallet(role, tenantID, userID string) models.ManagedWallet {
	seed := hashAny(map[string]interface{}{
		"provider": p.provider,
		"tenantId": tenantID,
		"userId":   userID,
		"role":     role,
		"network":  p.network,
		"asset":    p.asset,
	})
	address := "0x" + seed[:40]
	return models.ManagedWallet{
		Provider:           "firefly",
		ProviderWalletID:   "firefly_wallet_" + seed[:24],
		ProviderAccountID:  nonEmpty(userID, tenantID),
		TenantID:           tenantID,
		UserID:             userID,
		Role:               strings.ToLower(strings.TrimSpace(role)),
		Network:            p.network,
		Asset:              p.asset,
		Address:            address,
		Status:             "active",
		CustodyMode:        p.custodyMode,
		IsDefaultFunding:   strings.EqualFold(role, "buyer"),
		IsDefaultReceiving: strings.EqualFold(role, "seller"),
		LastSyncedAt:       time.Now().UTC(),
	}
}

func (p *fireflyWalletProvider) fallbackPaymentPayload(wallet models.ManagedWallet, requirement map[string]interface{}) map[string]interface{} {
	requirementHash := hashAny(requirement)
	proofSeed := sha256.Sum256([]byte(wallet.Address + ":" + requirementHash + ":" + wallet.ProviderWalletID))
	accepted := (&cdpWalletProvider{network: p.network, asset: p.asset}).requirementToAccepted(requirement)
	return map[string]interface{}{
		"x402Version":       2,
		"paymentIdentifier": "pay_" + requirementHash[:24],
		"method":            "x402_wallet",
		"provider":          wallet.Provider,
		"providerManaged":   true,
		"walletAddress":     wallet.Address,
		"accepted":          accepted,
		"payload": map[string]interface{}{
			"proof": "firefly:" + hex.EncodeToString(proofSeed[:]),
		},
		"resource": map[string]interface{}{
			"url":         stringFromAny(requirement["resource"]),
			"description": stringFromAny(requirement["description"]),
		},
	}
}

func (p *fireflyWalletProvider) liveSignPayment(ctx context.Context, wallet models.ManagedWallet, requirement map[string]interface{}) (map[string]interface{}, error) {
	builder := &cdpWalletProvider{network: p.network, asset: p.asset}
	accepted := builder.requirementToAccepted(requirement)
	if strings.TrimSpace(stringFromAny(accepted["payTo"])) == "" {
		return nil, fmt.Errorf("x402 payTo/payment address is missing")
	}
	amountUnits := stringFromAny(accepted["amount"])
	chainID := builder.chainIDForNetwork(stringFromAny(accepted["network"]))
	tokenAddress := builder.assetAddressForRequirement(stringFromAny(accepted["network"]), stringFromAny(accepted["asset"]))

	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, err
	}
	nonceHex := "0x" + hex.EncodeToString(nonceBytes)
	validAfter := time.Now().UTC().Unix()
	validBefore := time.Now().UTC().Add(time.Duration(intFromAny(accepted["maxTimeoutSeconds"])) * time.Second).Unix()
	if validBefore <= validAfter {
		validBefore = time.Now().UTC().Add(time.Hour).Unix()
	}

	typedData := map[string]interface{}{
		"types": map[string]interface{}{
			"EIP712Domain": []map[string]string{
				{"name": "name", "type": "string"},
				{"name": "version", "type": "string"},
				{"name": "chainId", "type": "uint256"},
				{"name": "verifyingContract", "type": "address"},
			},
			"TransferWithAuthorization": []map[string]string{
				{"name": "from", "type": "address"},
				{"name": "to", "type": "address"},
				{"name": "value", "type": "uint256"},
				{"name": "validAfter", "type": "uint256"},
				{"name": "validBefore", "type": "uint256"},
				{"name": "nonce", "type": "bytes32"},
			},
		},
		"domain": map[string]interface{}{
			"name":              baseUSDCName,
			"version":           baseUSDCVersion,
			"chainId":           chainID.String(),
			"verifyingContract": tokenAddress,
		},
		"primaryType": "TransferWithAuthorization",
		"message": map[string]interface{}{
			"from":        wallet.Address,
			"to":          accepted["payTo"],
			"value":       amountUnits,
			"validAfter":  fmt.Sprintf("%d", validAfter),
			"validBefore": fmt.Sprintf("%d", validBefore),
			"nonce":       nonceHex,
		},
	}
	typedDataJSON, _ := json.Marshal(typedData)
	signature, err := p.signTypedDataRPC(ctx, wallet.Address, string(typedDataJSON))
	if err != nil {
		return nil, err
	}
	requirementHash := hashAny(requirement)
	return map[string]interface{}{
		"x402Version":       2,
		"paymentIdentifier": "pay_" + requirementHash[:24],
		"method":            "x402_wallet",
		"provider":          wallet.Provider,
		"providerManaged":   true,
		"walletAddress":     wallet.Address,
		"accepted":          accepted,
		"payload": map[string]interface{}{
			"signature": signature,
			"authorization": map[string]interface{}{
				"from":        wallet.Address,
				"to":          accepted["payTo"],
				"value":       amountUnits,
				"validAfter":  fmt.Sprintf("%d", validAfter),
				"validBefore": fmt.Sprintf("%d", validBefore),
				"nonce":       nonceHex,
			},
		},
		"resource": map[string]interface{}{
			"url":         stringFromAny(requirement["resource"]),
			"description": stringFromAny(requirement["description"]),
		},
	}, nil
}

func (p *fireflyWalletProvider) signTypedDataRPC(ctx context.Context, address string, typedDataJSON string) (string, error) {
	body, _ := json.Marshal(jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_signTypedData_v4",
		Params:  []interface{}{address, typedDataJSON},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.signerURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(p.authToken) != "" {
		req.Header.Set("Authorization", "Bearer "+p.authToken)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out jsonRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", fmt.Errorf("firefly signer error %d: %s", out.Error.Code, out.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("firefly signer request failed with status %d", resp.StatusCode)
	}
	var sig string
	if err := json.Unmarshal(out.Result, &sig); err != nil {
		return "", err
	}
	if strings.TrimSpace(sig) == "" {
		return "", fmt.Errorf("firefly signer returned empty signature")
	}
	return sig, nil
}

func (p *cdpWalletProvider) liveProvisionWallet(ctx context.Context, role, tenantID, userID string) (string, string, error) {
	name := strings.TrimSpace(strings.ToLower(role + "-" + nonEmpty(userID, tenantID)))
	reqBody := cdpCreateAccountRequest{Name: name}
	resp, err := p.cdpRequest(ctx, http.MethodPost, cdpAccountsPath, reqBody)
	if err != nil {
		return "", "", err
	}
	address := firstString(resp,
		"address",
		"account.address",
		"evmAccount.address",
		"data.address",
		"result.address",
	)
	if address == "" {
		return "", "", fmt.Errorf("cdp account response did not include an address")
	}
	return address, firstString(resp, "id", "account.id", "evmAccount.id", "data.id", "result.id"), nil
}

func (p *cdpWalletProvider) liveSignPayment(ctx context.Context, wallet models.ManagedWallet, requirement map[string]interface{}) (map[string]interface{}, error) {
	accepted := p.requirementToAccepted(requirement)
	if strings.TrimSpace(accepted["payTo"].(string)) == "" {
		return nil, fmt.Errorf("x402 payTo/payment address is missing")
	}

	amountUnits := accepted["amount"].(string)
	chainID := p.chainIDForNetwork(stringFromAny(accepted["network"]))
	tokenAddress := p.assetAddressForRequirement(stringFromAny(accepted["network"]), stringFromAny(accepted["asset"]))
	tokenName := stringFromAny(mapStringAny(accepted["extra"])["name"])
	if tokenName == "" {
		tokenName = baseUSDCName
	}
	tokenVersion := stringFromAny(mapStringAny(accepted["extra"])["version"])
	if tokenVersion == "" {
		tokenVersion = baseUSDCVersion
	}

	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, err
	}
	nonceHex := "0x" + hex.EncodeToString(nonceBytes)
	validAfter := time.Now().UTC().Unix()
	validBefore := time.Now().UTC().Add(time.Duration(intFromAny(accepted["maxTimeoutSeconds"])) * time.Second).Unix()
	if validBefore <= validAfter {
		validBefore = time.Now().UTC().Add(time.Hour).Unix()
	}

	domain := cdpTypedDataDomain{
		Name:              tokenName,
		Version:           tokenVersion,
		ChainID:           chainID.String(),
		VerifyingContract: tokenAddress,
	}
	types := map[string][]cdpTypedDataField{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"TransferWithAuthorization": {
			{Name: "from", Type: "address"},
			{Name: "to", Type: "address"},
			{Name: "value", Type: "uint256"},
			{Name: "validAfter", Type: "uint256"},
			{Name: "validBefore", Type: "uint256"},
			{Name: "nonce", Type: "bytes32"},
		},
	}
	message := map[string]interface{}{
		"from":        wallet.Address,
		"to":          accepted["payTo"],
		"value":       amountUnits,
		"validAfter":  fmt.Sprintf("%d", validAfter),
		"validBefore": fmt.Sprintf("%d", validBefore),
		"nonce":       nonceHex,
	}
	signReq := cdpSignTypedDataRequest{
		Domain:      domain,
		Types:       types,
		PrimaryType: "TransferWithAuthorization",
		Message:     message,
	}
	resp, err := p.cdpRequest(ctx, http.MethodPost, fmt.Sprintf(cdpAccountSignTypedDataFmt, wallet.Address), signReq)
	if err != nil {
		return nil, err
	}
	signature := firstString(resp, "signature", "result.signature", "data.signature")
	if signature == "" {
		return nil, fmt.Errorf("cdp sign typed-data response did not include a signature")
	}

	requirementHash := hashAny(requirement)
	return map[string]interface{}{
		"x402Version":       2,
		"paymentIdentifier": "pay_" + requirementHash[:24],
		"method":            "x402_wallet",
		"provider":          wallet.Provider,
		"providerManaged":   true,
		"walletAddress":     wallet.Address,
		"accepted":          accepted,
		"payload": map[string]interface{}{
			"signature": signature,
			"authorization": map[string]interface{}{
				"from":        wallet.Address,
				"to":          accepted["payTo"],
				"value":       amountUnits,
				"validAfter":  fmt.Sprintf("%d", validAfter),
				"validBefore": fmt.Sprintf("%d", validBefore),
				"nonce":       nonceHex,
			},
		},
		"resource": map[string]interface{}{
			"url":         stringFromAny(requirement["resource"]),
			"description": stringFromAny(requirement["description"]),
		},
	}, nil
}

func (p *cdpWalletProvider) requirementToAccepted(requirement map[string]interface{}) map[string]interface{} {
	network := stringFromAny(requirement["network"])
	if network == "" {
		network = p.network
	}
	asset := stringFromAny(requirement["asset"])
	if asset == "" {
		asset = p.asset
	}
	amount := decimalAmountToUnitsString(requirementAmountFromAny(requirement["amount"]), 6)
	if units := stringFromAny(requirement["amountUnits"]); units != "" {
		amount = units
	}
	payTo := strings.TrimSpace(stringFromAny(requirement["paymentAddress"]))
	if payTo == "" {
		payTo = strings.TrimSpace(stringFromAny(requirement["payTo"]))
	}
	timeout := intFromAny(requirement["timeToLiveSecs"])
	if timeout <= 0 {
		timeout = 300
	}

	return map[string]interface{}{
		"scheme":            "exact",
		"network":           p.normalizeNetwork(network),
		"asset":             p.assetAddressForRequirement(network, asset),
		"amount":            amount,
		"payTo":             payTo,
		"maxTimeoutSeconds": timeout,
		"extra": map[string]interface{}{
			"name":                baseUSDCName,
			"version":             baseUSDCVersion,
			"assetTransferMethod": "eip3009",
			"serverId":            stringFromAny(requirement["serverId"]),
			"serverSlug":          stringFromAny(requirement["serverSlug"]),
			"toolName":            stringFromAny(requirement["toolName"]),
			"idempotencyKey":      stringFromAny(requirement["idempotencyKey"]),
		},
	}
}

func (p *cdpWalletProvider) normalizeNetwork(network string) string {
	switch strings.ToLower(strings.TrimSpace(network)) {
	case "", "base", "base-mainnet":
		return baseMainnetNetwork
	default:
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(network)), "eip155:") {
			return strings.TrimSpace(network)
		}
		return strings.TrimSpace(network)
	}
}

func (p *cdpWalletProvider) chainIDForNetwork(network string) *big.Int {
	normalized := p.normalizeNetwork(network)
	if strings.HasPrefix(normalized, "eip155:") {
		chainID := strings.TrimPrefix(normalized, "eip155:")
		if parsed, ok := new(big.Int).SetString(chainID, 10); ok {
			return parsed
		}
	}
	return big.NewInt(8453)
}

func (p *cdpWalletProvider) assetAddressForRequirement(network string, asset string) string {
	if strings.EqualFold(strings.TrimSpace(asset), "USDC") && p.normalizeNetwork(network) == baseMainnetNetwork {
		return baseUSDCAddress
	}
	if strings.HasPrefix(strings.TrimSpace(asset), "0x") {
		return strings.TrimSpace(asset)
	}
	return strings.TrimSpace(asset)
}

func (p *cdpWalletProvider) cdpRequest(ctx context.Context, method, path string, body interface{}) (map[string]interface{}, error) {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	} else {
		reqBody = []byte("{}")
	}
	req, err := http.NewRequestWithContext(ctx, method, cdpAPIBaseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	authHeader, err := p.buildBearerJWT(method, path)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+authHeader)
	req.Header.Set("X-Wallet-Auth", p.walletSecret)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("cdp response decode failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("cdp request failed with status %d: %s", resp.StatusCode, stringFromAny(out["error"]))
	}
	return out, nil
}

func (p *cdpWalletProvider) buildBearerJWT(method, path string) (string, error) {
	signingMethod, key, err := parseCDPPrivateKey(strings.TrimSpace(p.apiKeySecret))
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	claims := cdpBearerClaims{
		URIs: []string{fmt.Sprintf("%s %s%s", method, cdpAPIHost, path)},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strings.TrimSpace(p.apiKeyID),
			Issuer:    "coinbase-cloud",
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		},
	}
	if strings.TrimSpace(p.apiKeyID) == "" {
		return "", fmt.Errorf("cdp api key id is missing")
	}
	token := jwt.NewWithClaims(signingMethod, claims)
	token.Header["kid"] = strings.TrimSpace(p.apiKeyID)
	token.Header["typ"] = "JWT"
	return token.SignedString(key)
}

func parseCDPPrivateKey(secret string) (jwt.SigningMethod, interface{}, error) {
	if strings.HasPrefix(secret, "-----BEGIN") {
		block, _ := pem.Decode([]byte(secret))
		if block == nil {
			return nil, nil, fmt.Errorf("failed to decode CDP PEM private key")
		}
		if parsed, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
			return jwt.SigningMethodES256, parsed, nil
		}
		if keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			switch key := keyAny.(type) {
			case *ecdsa.PrivateKey:
				return jwt.SigningMethodES256, key, nil
			case ed25519.PrivateKey:
				return jwt.SigningMethodEdDSA, key, nil
			}
		}
		return nil, nil, fmt.Errorf("unsupported CDP PEM private key")
	}
	decoded, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to base64 decode CDP private key: %w", err)
	}
	if len(decoded) == ed25519.PrivateKeySize {
		return jwt.SigningMethodEdDSA, ed25519.PrivateKey(decoded), nil
	}
	return nil, nil, fmt.Errorf("unsupported CDP private key format")
}

func decimalAmountToUnitsString(amount float64, decimals int) string {
	if amount <= 0 {
		return "0"
	}
	value := new(big.Rat).SetFloat64(amount)
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	value.Mul(value, new(big.Rat).SetInt(multiplier))
	out := new(big.Int).Quo(value.Num(), value.Denom())
	return out.String()
}

func firstString(data map[string]interface{}, paths ...string) string {
	for _, path := range paths {
		current := interface{}(data)
		segments := strings.Split(path, ".")
		for _, segment := range segments {
			node, ok := current.(map[string]interface{})
			if !ok {
				current = nil
				break
			}
			current = node[segment]
		}
		if s, ok := current.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func mapStringAny(v interface{}) map[string]interface{} {
	out, _ := v.(map[string]interface{})
	if out == nil {
		return map[string]interface{}{}
	}
	return out
}

func requirementAmountFromAny(v interface{}) float64 {
	switch value := v.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case int32:
		return float64(value)
	case string:
		rat, ok := new(big.Rat).SetString(strings.TrimSpace(value))
		if ok {
			f, _ := rat.Float64()
			return f
		}
	}
	return 0
}

func intFromAny(v interface{}) int {
	switch value := v.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case int32:
		return int(value)
	case float64:
		return int(value)
	case float32:
		return int(value)
	case string:
		bi, ok := new(big.Int).SetString(strings.TrimSpace(value), 10)
		if ok {
			return int(bi.Int64())
		}
	}
	return 0
}

func (a *App) ensureManagedWallet(ctx context.Context, tenantID, userID, role string) (models.ManagedWallet, error) {
	activeProvider := a.resolveWalletProviderName(a.resolvedIntegrations().Wallet)
	if wallet, ok := a.store.GetManagedWalletByOwner(tenantID, userID, role); ok && strings.EqualFold(wallet.Provider, activeProvider) {
		return wallet, nil
	}
	provider := a.currentWalletProvider()
	wallet, err := provider.provisionWallet(ctx, role, tenantID, userID)
	if err != nil {
		return models.ManagedWallet{}, err
	}
	wallet = a.store.UpsertManagedWallet(wallet)
	if strings.EqualFold(role, "buyer") {
		policy := a.effectivePaymentPolicy(tenantID, userID)
		if strings.TrimSpace(policy.WalletAddress) == "" {
			policy.WalletAddress = wallet.Address
		}
		if strings.TrimSpace(policy.SIWXWallet) == "" {
			policy.SIWXWallet = wallet.Address
		}
		a.store.UpsertPaymentPolicy(policy)
	}
	return wallet, nil
}

func (a *App) ensureBuyerManagedWallet(ctx context.Context, tenantID, userID string) (models.ManagedWallet, error) {
	return a.ensureManagedWallet(ctx, tenantID, userID, "buyer")
}

func (a *App) ensureSellerManagedWallet(ctx context.Context, tenantID string) (models.ManagedWallet, error) {
	return a.ensureManagedWallet(ctx, tenantID, "", "seller")
}

func (a *App) signManagedWalletPayment(ctx context.Context, tenantID, userID string, requirement map[string]interface{}) (map[string]interface{}, models.ManagedWallet, error) {
	if !a.resolvedIntegrations().Wallet.ManagedAutoPayEnabled {
		return nil, models.ManagedWallet{}, fmt.Errorf("managed wallet auto-pay is disabled")
	}
	wallet, err := a.ensureBuyerManagedWallet(ctx, tenantID, userID)
	if err != nil {
		return nil, models.ManagedWallet{}, err
	}
	payload, err := a.currentWalletProvider().signPayment(ctx, wallet, requirement)
	if err != nil {
		return nil, models.ManagedWallet{}, err
	}
	return payload, wallet, nil
}

func (a *App) currentBuyerWalletAttribution(ctx context.Context, tenantID, userID string) (models.ManagedWallet, bool) {
	wallet, err := a.ensureBuyerManagedWallet(ctx, tenantID, userID)
	if err != nil {
		return models.ManagedWallet{}, false
	}
	return wallet, true
}

func applyIntentWallet(intent *models.X402Intent, wallet models.ManagedWallet) {
	intent.WalletID = strings.TrimSpace(nonEmpty(wallet.ID, wallet.ProviderWalletID))
	intent.WalletAddress = strings.TrimSpace(wallet.Address)
	intent.WalletProvider = strings.TrimSpace(wallet.Provider)
}

func (a *App) resolveServerPaymentAddress(server models.Server) string {
	if strings.TrimSpace(server.PaymentAddress) != "" {
		return strings.TrimSpace(server.PaymentAddress)
	}
	wallet, err := a.ensureSellerManagedWallet(context.Background(), server.TenantID)
	if err != nil {
		return ""
	}
	return wallet.Address
}
