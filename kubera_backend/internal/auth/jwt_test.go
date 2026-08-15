package auth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVerifierAcceptsValidNeonStyleToken(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	const keyID = "test-key"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]string{{
				"kid": keyID,
				"kty": "OKP",
				"alg": "EdDSA",
				"crv": "Ed25519",
				"x":   base64.RawURLEncoding.EncodeToString(publicKey),
			}},
		})
	}))
	defer server.Close()

	verifier, err := newVerifier("", server.URL, "https://auth.example", "kubera")
	if err != nil {
		t.Fatal(err)
	}
	token := signedToken(t, privateKey, keyID, map[string]any{
		"sub":  "auth-user-1",
		"name": "Asha",
		"iss":  "https://auth.example",
		"aud":  "kubera",
		"exp":  time.Now().Add(time.Minute).Unix(),
	})

	claims, err := verifier.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.Subject != "auth-user-1" || claims.Name != "Asha" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerifierRejectsExpiredToken(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	const keyID = "test-key"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]string{{
				"kid": keyID,
				"kty": "OKP",
				"alg": "EdDSA",
				"crv": "Ed25519",
				"x":   base64.RawURLEncoding.EncodeToString(publicKey),
			}},
		})
	}))
	defer server.Close()

	verifier, err := newVerifier("", server.URL, "https://auth.example", "")
	if err != nil {
		t.Fatal(err)
	}
	token := signedToken(t, privateKey, keyID, map[string]any{
		"sub": "auth-user-1",
		"iss": "https://auth.example",
		"exp": time.Now().Add(-time.Minute).Unix(),
	})
	if _, err := verifier.Verify(context.Background(), token); err == nil {
		t.Fatal("Verify() accepted an expired token")
	}
}

func TestVerifierDoesNotInferIssuerFromBaseURL(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	const keyID = "test-key"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kid": keyID, "kty": "OKP", "alg": "EdDSA", "crv": "Ed25519", "x": base64.RawURLEncoding.EncodeToString(publicKey),
		}}})
	}))
	defer server.Close()

	verifier, err := newVerifier("https://auth-base.example", server.URL, "", "")
	if err != nil {
		t.Fatal(err)
	}
	token := signedToken(t, privateKey, keyID, map[string]any{
		"sub": "auth-user-1", "iss": "https://issuer.example", "exp": time.Now().Add(time.Minute).Unix(),
	})
	if _, err := verifier.Verify(context.Background(), token); err != nil {
		t.Fatalf("Verify() rejected a token signed by the configured JWKS: %v", err)
	}
}

func TestDefaultShopName(t *testing.T) {
	if got := defaultShopName(" Asha "); got != "Asha's Shop" {
		t.Fatalf("defaultShopName() = %q", got)
	}
	if got := defaultShopName(""); got != "My Shop" {
		t.Fatalf("defaultShopName() = %q", got)
	}
}

func signedToken(t *testing.T, privateKey ed25519.PrivateKey, keyID string, claims map[string]any) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "EdDSA", "kid": keyID, "typ": "JWT"})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	signature := ed25519.Sign(privateKey, []byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}
