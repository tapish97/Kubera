CREATE TABLE IF NOT EXISTS shop_daily_closings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID NOT NULL REFERENCES shops(id),
    business_date DATE NOT NULL,
    sales_transactions INTEGER NOT NULL DEFAULT 0,
    sales_revenue NUMERIC(14,2) NOT NULL DEFAULT 0,
    gross_profit NUMERIC(14,2) NOT NULL DEFAULT 0,
    purchase_value NUMERIC(14,2) NOT NULL DEFAULT 0,
    spoilage_loss NUMERIC(14,2) NOT NULL DEFAULT 0,
    used_estimates BOOLEAN NOT NULL DEFAULT FALSE,
    closed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (shop_id, business_date)
);

CREATE INDEX IF NOT EXISTS shop_daily_closings_shop_date_idx
ON shop_daily_closings (shop_id, business_date DESC);
