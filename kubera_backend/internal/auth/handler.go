package auth

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var currencyCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

type onboardingRequest struct {
	ProfileName   string   `json:"profile_name"`
	ShopName      string   `json:"shop_name"`
	Currency      string   `json:"currency"`
	Timezone      string   `json:"timezone"`
	LocationLabel string   `json:"location_label"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

type profileRequest struct {
	Name string `json:"name"`
}

type shopRequest struct {
	Name          string   `json:"name"`
	Currency      string   `json:"currency"`
	Timezone      string   `json:"timezone"`
	LocationLabel string   `json:"location_label"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeAuthError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	writeAuthJSON(w, http.StatusOK, principal)
}

func (h *Handler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeAuthError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var req onboardingRequest
	if !decodeAuthJSON(w, r, &req) {
		return
	}
	req.ProfileName = strings.TrimSpace(req.ProfileName)
	req.ShopName = strings.TrimSpace(req.ShopName)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	req.Timezone = strings.TrimSpace(req.Timezone)
	req.LocationLabel = strings.Join(strings.Fields(req.LocationLabel), " ")
	if message := validateSettings(req.ProfileName, req.ShopName, req.Currency, req.Timezone); message != "" {
		writeAuthError(w, message, http.StatusBadRequest)
		return
	}
	if message := validateLocation(req.LocationLabel, req.Latitude, req.Longitude); message != "" {
		writeAuthError(w, message, http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin(r.Context())
	if err != nil {
		writeAuthError(w, "could not start onboarding", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	_, err = tx.Exec(r.Context(), `
		UPDATE user_profiles
		SET name = $1,
			onboarding_completed_at = COALESCE(onboarding_completed_at, NOW()),
			updated_at = NOW()
		WHERE id = $2
	`, req.ProfileName, principal.ProfileID)
	if err == nil {
		_, err = tx.Exec(r.Context(), `
			UPDATE shops
			SET name = $1, currency = $2, timezone = $3, location_label = NULLIF($4, ''), latitude = $5, longitude = $6, updated_at = NOW()
			WHERE id = $7 AND owner_profile_id = $8 AND is_active = TRUE
		`, req.ShopName, req.Currency, req.Timezone, req.LocationLabel, req.Latitude, req.Longitude, principal.ShopID, principal.ProfileID)
	}
	if err != nil {
		writeAuthError(w, "could not save onboarding", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeAuthError(w, "could not save onboarding", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	principal.ProfileName = req.ProfileName
	principal.ShopName = req.ShopName
	principal.Currency = req.Currency
	principal.Timezone = req.Timezone
	principal.LocationLabel = req.LocationLabel
	principal.Latitude = req.Latitude
	principal.Longitude = req.Longitude
	principal.OnboardingCompletedAt = &now
	writeAuthJSON(w, http.StatusOK, principal)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeAuthError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	var req profileRequest
	if !decodeAuthJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 2 || len(req.Name) > 100 {
		writeAuthError(w, "name must be between 2 and 100 characters", http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(r.Context(), `
		UPDATE user_profiles SET name = $1, updated_at = NOW() WHERE id = $2
	`, req.Name, principal.ProfileID)
	if err != nil {
		writeAuthError(w, "could not update profile", http.StatusInternalServerError)
		return
	}
	principal.ProfileName = req.Name
	writeAuthJSON(w, http.StatusOK, principal)
}

func (h *Handler) UpdateShop(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeAuthError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	var req shopRequest
	if !decodeAuthJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	req.Timezone = strings.TrimSpace(req.Timezone)
	req.LocationLabel = strings.Join(strings.Fields(req.LocationLabel), " ")
	if message := validateShop(req.Name, req.Currency, req.Timezone); message != "" {
		writeAuthError(w, message, http.StatusBadRequest)
		return
	}
	if message := validateLocation(req.LocationLabel, req.Latitude, req.Longitude); message != "" {
		writeAuthError(w, message, http.StatusBadRequest)
		return
	}
	command, err := h.db.Exec(r.Context(), `
		UPDATE shops
		SET name = $1, currency = $2, timezone = $3, location_label = NULLIF($4, ''), latitude = $5, longitude = $6, updated_at = NOW()
		WHERE id = $7 AND owner_profile_id = $8 AND is_active = TRUE
	`, req.Name, req.Currency, req.Timezone, req.LocationLabel, req.Latitude, req.Longitude, principal.ShopID, principal.ProfileID)
	if err != nil || command.RowsAffected() != 1 {
		writeAuthError(w, "could not update shop", http.StatusInternalServerError)
		return
	}
	principal.ShopName = req.Name
	principal.Currency = req.Currency
	principal.Timezone = req.Timezone
	principal.LocationLabel = req.LocationLabel
	principal.Latitude = req.Latitude
	principal.Longitude = req.Longitude
	writeAuthJSON(w, http.StatusOK, principal)
}

func validateSettings(profileName, shopName, currency, timezone string) string {
	if len(profileName) < 2 || len(profileName) > 100 {
		return "name must be between 2 and 100 characters"
	}
	return validateShop(shopName, currency, timezone)
}

func validateShop(shopName, currency, timezone string) string {
	if len(shopName) < 2 || len(shopName) > 120 {
		return "shop name must be between 2 and 120 characters"
	}
	if !currencyCodePattern.MatchString(currency) {
		return "currency must be a three-letter code"
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return "invalid timezone"
	}
	return ""
}

func validateLocation(label string, latitude, longitude *float64) string {
	if len([]rune(label)) > 180 {
		return "location must be 180 characters or fewer"
	}
	if (latitude == nil) != (longitude == nil) {
		return "latitude and longitude must be provided together"
	}
	if latitude != nil && (*latitude < -90 || *latitude > 90 || *longitude < -180 || *longitude > 180) {
		return "invalid location coordinates"
	}
	return ""
}

func decodeAuthJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeAuthError(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

func writeAuthJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
