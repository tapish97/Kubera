package auth

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var errInvalidToken = errors.New("invalid authentication token")

type Claims struct {
	Subject  string
	Name     string
	Issuer   string
	Audience []string
	Expires  time.Time
}

type verifier struct {
	jwksURL  string
	issuer   string
	audience string
	client   *http.Client
	mu       sync.RWMutex
	keys     map[string]ed25519.PublicKey
	fetched  time.Time
}

type jwksResponse struct {
	Keys []struct {
		KeyID     string `json:"kid"`
		KeyType   string `json:"kty"`
		Algorithm string `json:"alg"`
		Curve     string `json:"crv"`
		X         string `json:"x"`
	} `json:"keys"`
}

func newVerifier(baseURL, jwksURL, issuer, audience string) (*verifier, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	if jwksURL == "" && baseURL != "" {
		jwksURL = baseURL + "/.well-known/jwks.json"
	}
	if jwksURL == "" {
		return nil, errors.New("NEON_AUTH_BASE_URL or NEON_AUTH_JWKS_URL must be set")
	}
	return &verifier{
		jwksURL:  jwksURL,
		issuer:   strings.TrimRight(issuer, "/"),
		audience: audience,
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (v *verifier) Verify(ctx context.Context, rawToken string) (Claims, error) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return Claims{}, errInvalidToken
	}

	var header struct {
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
	}
	if err := decodeJWTPart(parts[0], &header); err != nil || header.Algorithm != "EdDSA" || header.KeyID == "" {
		return Claims{}, errInvalidToken
	}

	key, err := v.key(ctx, header.KeyID, false)
	if err != nil {
		return Claims{}, errInvalidToken
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !ed25519.Verify(key, []byte(parts[0]+"."+parts[1]), signature) {
		key, err = v.key(ctx, header.KeyID, true)
		if err != nil || !ed25519.Verify(key, []byte(parts[0]+"."+parts[1]), signature) {
			return Claims{}, errInvalidToken
		}
	}

	var payload struct {
		Subject   string          `json:"sub"`
		Name      string          `json:"name"`
		Issuer    string          `json:"iss"`
		Audience  json.RawMessage `json:"aud"`
		Expires   int64           `json:"exp"`
		NotBefore int64           `json:"nbf"`
	}
	if err := decodeJWTPart(parts[1], &payload); err != nil {
		return Claims{}, errInvalidToken
	}

	now := time.Now()
	if payload.Subject == "" || payload.Expires == 0 || !now.Before(time.Unix(payload.Expires, 0)) {
		return Claims{}, errInvalidToken
	}
	if payload.NotBefore != 0 && now.Add(30*time.Second).Before(time.Unix(payload.NotBefore, 0)) {
		return Claims{}, errInvalidToken
	}
	if v.issuer != "" && strings.TrimRight(payload.Issuer, "/") != v.issuer {
		return Claims{}, errInvalidToken
	}

	audience, err := parseAudience(payload.Audience)
	if err != nil || (v.audience != "" && !contains(audience, v.audience)) {
		return Claims{}, errInvalidToken
	}

	return Claims{
		Subject:  payload.Subject,
		Name:     payload.Name,
		Issuer:   payload.Issuer,
		Audience: audience,
		Expires:  time.Unix(payload.Expires, 0),
	}, nil
}

func (v *verifier) key(ctx context.Context, keyID string, forceRefresh bool) (ed25519.PublicKey, error) {
	v.mu.RLock()
	key := v.keys[keyID]
	fetched := v.fetched
	v.mu.RUnlock()
	if !forceRefresh && key != nil && time.Since(fetched) < time.Hour {
		return key, nil
	}
	if err := v.refresh(ctx); err != nil {
		return nil, err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	key = v.keys[keyID]
	if key == nil {
		return nil, fmt.Errorf("signing key %q not found", keyID)
	}
	return key, nil
}

func (v *verifier) refresh(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}
	response, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned %s", response.Status)
	}
	var document jwksResponse
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return err
	}
	keys := make(map[string]ed25519.PublicKey)
	for _, item := range document.Keys {
		if item.KeyID == "" || item.KeyType != "OKP" || item.Algorithm != "EdDSA" || item.Curve != "Ed25519" {
			continue
		}
		raw, err := base64.RawURLEncoding.DecodeString(item.X)
		if err != nil || len(raw) != ed25519.PublicKeySize {
			continue
		}
		keys[item.KeyID] = ed25519.PublicKey(raw)
	}
	if len(keys) == 0 {
		return errors.New("JWKS contains no supported Ed25519 signing keys")
	}
	v.keys = keys
	v.fetched = time.Now()
	return nil
}

func decodeJWTPart(encoded string, target any) error {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func parseAudience(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err == nil {
		return many, nil
	}
	var one string
	if err := json.Unmarshal(raw, &one); err != nil {
		return nil, err
	}
	return []string{one}, nil
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
