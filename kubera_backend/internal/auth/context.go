package auth

import "context"

type contextKey string

const principalKey contextKey = "authenticated-principal"

type Principal struct {
	AuthUserID string `json:"auth_user_id"`
	ProfileID  string `json:"profile_id"`
	ShopID     string `json:"shop_id"`
	ShopName   string `json:"shop_name"`
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
