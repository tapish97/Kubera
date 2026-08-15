package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"kubera_backend/internal/requestlog"
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
			slog.Warn("authentication rejected", "method", r.Method, "path", r.URL.Path, "reason", "missing bearer token")
			writeAuthError(w, "authentication required", http.StatusUnauthorized)
			return
		}
		claims, err := m.verifier.Verify(r.Context(), rawToken)
		if err != nil {
			slog.Warn("authentication rejected", "method", r.Method, "path", r.URL.Path, "reason", "invalid or expired token")
			writeAuthError(w, "invalid or expired authentication token", http.StatusUnauthorized)
			return
		}
		principal, err := resolvePrincipal(r.Context(), m.db, claims)
		if err != nil {
			slog.Error("could not resolve authenticated account", "auth_user_id", claims.Subject, "error", err)
			writeAuthError(w, "could not resolve authenticated account", http.StatusInternalServerError)
			return
		}
		authenticated := r.WithContext(withPrincipal(r.Context(), principal))
		requestlog.Middleware(slog.Default(), func(request *http.Request) []any {
			resolved, _ := PrincipalFromContext(request.Context())
			return []any{"auth_user_id", resolved.AuthUserID, "profile_id", resolved.ProfileID, "shop_id", resolved.ShopID}
		}, next).ServeHTTP(w, authenticated)
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
