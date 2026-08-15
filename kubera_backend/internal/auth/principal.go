package auth

import (
	"context"
	"errors"
	"strings"

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
	err = tx.QueryRow(ctx, `
		INSERT INTO user_profiles (auth_user_id, name)
		VALUES ($1, NULLIF($2, ''))
		ON CONFLICT (auth_user_id) DO UPDATE
		SET name = COALESCE(user_profiles.name, EXCLUDED.name)
		RETURNING id
	`, claims.Subject, strings.TrimSpace(claims.Name)).Scan(&profileID)
	if err != nil {
		return Principal{}, err
	}

	var shopID, shopName string
	err = tx.QueryRow(ctx, `
		SELECT id, name
		FROM shops
		WHERE owner_profile_id = $1 AND is_active = TRUE
		ORDER BY created_at, id
		LIMIT 1
	`, profileID).Scan(&shopID, &shopName)
	if errors.Is(err, pgx.ErrNoRows) {
		shopName = defaultShopName(claims.Name)
		err = tx.QueryRow(ctx, `
			INSERT INTO shops (owner_profile_id, name)
			VALUES ($1, $2)
			RETURNING id, name
		`, profileID, shopName).Scan(&shopID, &shopName)
	}
	if err != nil {
		return Principal{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Principal{}, err
	}
	return Principal{
		AuthUserID: claims.Subject,
		ProfileID:  profileID,
		ShopID:     shopID,
		ShopName:   shopName,
	}, nil
}

func defaultShopName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "My Shop"
	}
	return name + "'s Shop"
}
