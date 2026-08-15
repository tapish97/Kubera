package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func resolvePrincipal(ctx context.Context, db *pgxpool.Pool, claims Claims) (Principal, error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Principal{}, err
	}
	defer tx.Rollback(ctx)

	// Serializes provisioning for one auth user without preventing future
	// support for multiple shops per profile.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, claims.Subject); err != nil {
		return Principal{}, err
	}

	var profileID string
	var profileName *string
	var onboardingCompletedAt *time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO user_profiles (auth_user_id, name)
		VALUES ($1, NULLIF($2, ''))
		ON CONFLICT (auth_user_id) DO UPDATE
		SET name = COALESCE(user_profiles.name, EXCLUDED.name)
		RETURNING id, name, onboarding_completed_at
	`, claims.Subject, strings.TrimSpace(claims.Name)).Scan(&profileID, &profileName, &onboardingCompletedAt)
	if err != nil {
		return Principal{}, err
	}

	var shopID, shopName, currency, timezone string
	err = tx.QueryRow(ctx, `
		SELECT id, name, currency, timezone
		FROM shops
		WHERE owner_profile_id = $1 AND is_active = TRUE
		ORDER BY created_at, id
		LIMIT 1
	`, profileID).Scan(&shopID, &shopName, &currency, &timezone)
	if errors.Is(err, pgx.ErrNoRows) {
		shopName = defaultShopName(claims.Name)
		err = tx.QueryRow(ctx, `
			INSERT INTO shops (owner_profile_id, name)
			VALUES ($1, $2)
			RETURNING id, name, currency, timezone
		`, profileID, shopName).Scan(&shopID, &shopName, &currency, &timezone)
	}
	if err != nil {
		return Principal{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Principal{}, err
	}
	return Principal{
		AuthUserID:            claims.Subject,
		ProfileID:             profileID,
		ProfileName:           stringValue(profileName),
		ShopID:                shopID,
		ShopName:              shopName,
		Currency:              currency,
		Timezone:              timezone,
		OnboardingCompletedAt: onboardingCompletedAt,
	}, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func defaultShopName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "My Shop"
	}
	return name + "'s Shop"
}
