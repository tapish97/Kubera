package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Middleware struct {
	db       *pgxpool.Pool
	verifier *verifier
}

func NewMiddleware(db *pgxpool.Pool) (*Middleware, error) {
	verifier, err := newVerifier(
		os.Getenv("NEON_AUTH_BASE_URL"),
		os.Getenv("NEON_AUTH_JWKS_URL"),
		os.Getenv("NEON_AUTH_ISSUER"),
		os.Getenv("NEON_AUTH_AUDIENCE"),
	)
	if err != nil {
		return nil, err
	}
	return &Middleware{db: db, verifier: verifier}, nil
}

func (m *Middleware) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeAuthError(w, "authentication required", http.StatusUnauthorized)
			return
		}
		claims, err := m.verifier.Verify(r.Context(), rawToken)
		if err != nil {
			writeAuthError(w, "invalid or expired authentication token", http.StatusUnauthorized)
			return
		}
		principal, err := resolvePrincipal(r.Context(), m.db, claims)
		if err != nil {
			log.Printf("resolve authenticated user %q: %v", claims.Subject, err)
			writeAuthError(w, "could not resolve authenticated account", http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), principal)))
	})
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	returnToken := ""
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		returnToken = parts[1]
	}
	return returnToken, returnToken != ""
}

func writeAuthError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
