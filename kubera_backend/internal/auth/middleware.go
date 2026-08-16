package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

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
		authStarted := time.Now()
		rawToken, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			requestlog.AddAttributes(r.Context(), "auth_total_ms", time.Since(authStarted).Milliseconds())
			slog.Warn("authentication rejected", "method", r.Method, "path", r.URL.Path, "reason", "missing bearer token")
			writeAuthError(w, "authentication required", http.StatusUnauthorized)
			return
		}
		verifyStarted := time.Now()
		claims, err := m.verifier.Verify(r.Context(), rawToken)
		verifyDuration := time.Since(verifyStarted)
		if err != nil {
			requestlog.AddAttributes(r.Context(), "auth_verify_ms", verifyDuration.Milliseconds(), "auth_total_ms", time.Since(authStarted).Milliseconds())
			slog.Warn("authentication rejected", "method", r.Method, "path", r.URL.Path, "reason", "invalid or expired token")
			writeAuthError(w, "invalid or expired authentication token", http.StatusUnauthorized)
			return
		}
		principalStarted := time.Now()
		principal, err := resolvePrincipal(r.Context(), m.db, claims)
		principalDuration := time.Since(principalStarted)
		authDuration := time.Since(authStarted)
		requestlog.AddAttributes(r.Context(),
			"auth_verify_ms", verifyDuration.Milliseconds(),
			"principal_query_ms", principalDuration.Milliseconds(),
			"auth_total_ms", authDuration.Milliseconds(),
		)
		if err != nil {
			slog.Error("could not resolve authenticated account", "auth_user_id", claims.Subject, "error", err)
			writeAuthError(w, "could not resolve authenticated account", http.StatusInternalServerError)
			return
		}
		if authDuration >= requestlog.SlowThreshold() {
			slog.Warn("slow authentication", "method", r.Method, "path", r.URL.Path, "duration_ms", authDuration.Milliseconds(), "verify_ms", verifyDuration.Milliseconds(), "principal_query_ms", principalDuration.Milliseconds())
		}
		authenticated := r.WithContext(withPrincipal(r.Context(), principal))
		requestlog.AddAttributes(authenticated.Context(), "auth_user_id", principal.AuthUserID, "profile_id", principal.ProfileID, "shop_id", principal.ShopID)
		next.ServeHTTP(w, authenticated)
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
