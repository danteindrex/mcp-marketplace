package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

const (
	TokenPurposeAppAuth     = "app_auth"
	TokenPurposeOAuthAccess = "oauth_access"
)

type Claims struct {
	UserID   string      `json:"userId"`
	TenantID string      `json:"tenantId"`
	Role     models.Role `json:"role"`
	Purpose  string      `json:"purpose"`
	Scopes   []string    `json:"scopes,omitempty"`
	Resource string      `json:"resource,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
	issuer     string
}

type jwk struct {
	KeyType   string   `json:"kty"`
	Use       string   `json:"use"`
	KeyID     string   `json:"kid"`
	Algorithm string   `json:"alg"`
	Modulus   string   `json:"n"`
	Exponent  string   `json:"e"`
	Ops       []string `json:"key_ops,omitempty"`
}

type JWKS struct {
	Keys []jwk `json:"keys"`
}

func NewJWTManager(cfg config.Config) (*JWTManager, error) {
	privateKey, publicKey, err := loadOrGenerateKeys(cfg)
	if err != nil {
		return nil, err
	}
	keyID := cfg.JWTKeyID
	if keyID == "" {
		keyID = "default-rs256-key"
	}
	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		keyID:      keyID,
		issuer:     strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
	}, nil
}

func (j *JWTManager) Generate(user models.User) (string, error) {
	return j.GenerateAppToken(user)
}

func (j *JWTManager) GenerateAppToken(user models.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Role:     user.Role,
		Purpose:  TokenPurposeAppAuth,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    j.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}
	return j.signClaims(claims)
}

func (j *JWTManager) GenerateOAuthAccessToken(user models.User, resource string, scopes []string) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Role:     user.Role,
		Purpose:  TokenPurposeOAuthAccess,
		Scopes:   scopes,
		Resource: resource,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    j.issuer,
			Audience:  jwt.ClaimStrings{resource},
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}
	return j.signClaims(claims)
}

func (j *JWTManager) signClaims(claims Claims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = j.keyID
	return t.SignedString(j.privateKey)
}

func (j *JWTManager) Parse(token string) (*Claims, error) {
	return j.ParseForPurpose(token, "")
}

func (j *JWTManager) ParseForPurpose(token string, expectedPurpose string) (*Claims, error) {
	parserOptions := []jwt.ParserOption{jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()})}
	if j.issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(j.issuer))
	}
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("invalid signing method")
		}
		if kid, _ := t.Header["kid"].(string); kid != "" && kid != j.keyID {
			return nil, errors.New("unknown key id")
		}
		return j.publicKey, nil
	}, parserOptions...)
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	if expectedPurpose != "" && claims.Purpose != expectedPurpose {
		return nil, errors.New("invalid token purpose")
	}
	if claims.Purpose == "" {
		return nil, errors.New("missing token purpose")
	}
	return claims, nil
}

func (j *JWTManager) JWKS() JWKS {
	return JWKS{
		Keys: []jwk{buildJWK(j.publicKey, j.keyID)},
	}
}

func loadOrGenerateKeys(cfg config.Config) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	hasPrivate := cfg.JWTPrivateKeyPEM != ""
	hasPublic := cfg.JWTPublicKeyPEM != ""
	if hasPrivate || hasPublic {
		if !hasPrivate || !hasPublic {
			return nil, nil, fmt.Errorf("JWT_PRIVATE_KEY_PEM and JWT_PUBLIC_KEY_PEM must both be set")
		}
		privateKey, err := parseRSAPrivateKey(cfg.JWTPrivateKeyPEM)
		if err != nil {
			return nil, nil, err
		}
		publicKey, err := parseRSAPublicKey(cfg.JWTPublicKeyPEM)
		if err != nil {
			return nil, nil, err
		}
		return privateKey, publicKey, nil
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate rsa key: %w", err)
	}
	return privateKey, &privateKey.PublicKey, nil
}

func parseRSAPrivateKey(raw string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, errors.New("invalid JWT_PRIVATE_KEY_PEM")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("JWT_PRIVATE_KEY_PEM is not an RSA private key")
	}
	return key, nil
}

func parseRSAPublicKey(raw string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, errors.New("invalid JWT_PUBLIC_KEY_PEM")
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	key, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("JWT_PUBLIC_KEY_PEM is not an RSA public key")
	}
	return key, nil
}

func buildJWK(publicKey *rsa.PublicKey, keyID string) jwk {
	return jwk{
		KeyType:   "RSA",
		Use:       "sig",
		KeyID:     keyID,
		Algorithm: "RS256",
		Modulus:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
		Exponent:  encodeExponent(publicKey.E),
		Ops:       []string{"verify"},
	}
}

func encodeExponent(exponent int) string {
	return base64.RawURLEncoding.EncodeToString(big.NewInt(int64(exponent)).Bytes())
}
