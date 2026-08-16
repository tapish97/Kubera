CREATE INDEX IF NOT EXISTS inventory_batches_shop_received_idx
ON inventory_batches (shop_id, received_at);

CREATE INDEX IF NOT EXISTS sales_shop_sold_idx
ON sales (shop_id, sold_at);

CREATE INDEX IF NOT EXISTS sale_items_batch_sale_idx
ON sale_items (batch_id, sale_id);

CREATE INDEX IF NOT EXISTS inventory_adjustments_batch_adjusted_idx
ON inventory_adjustments (batch_id, adjusted_at);
