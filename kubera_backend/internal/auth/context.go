package auth

import (
	"context"
	"time"
)

type contextKey string

const principalKey contextKey = "authenticated-principal"

type Principal struct {
	AuthUserID            string     `json:"auth_user_id"`
	ProfileID             string     `json:"profile_id"`
	ProfileName           string     `json:"profile_name"`
	PreferredLocale       string     `json:"preferred_locale"`
	ShopID                string     `json:"shop_id"`
	ShopName              string     `json:"shop_name"`
	Currency              string     `json:"currency"`
	Timezone              string     `json:"timezone"`
	LocationLabel         string     `json:"location_label"`
	Latitude              *float64   `json:"latitude"`
	Longitude             *float64   `json:"longitude"`
	OnboardingCompletedAt *time.Time `json:"onboarding_completed_at"`
}

func withPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalKey).(Principal)
	return principal, ok
}

func ShopIDFromContext(ctx context.Context) (string, bool) {
	principal, ok := PrincipalFromContext(ctx)
	return principal.ShopID, ok && principal.ShopID != ""
}
