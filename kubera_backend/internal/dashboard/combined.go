package dashboard

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"kubera_backend/internal/auth"
	"kubera_backend/internal/requestlog"
)

type combinedSummary struct {
	TotalInventoryQuantity float64 `json:"total_inventory_quantity"`
	TodaySales             float64 `json:"today_sales"`
	TodayGrossProfit       float64 `json:"today_gross_profit"`
}

type combinedInventoryItem struct {
	BatchID              string    `json:"batch_id"`
	Fruit                string    `json:"fruit"`
	Supplier             string    `json:"supplier"`
	Mark                 string    `json:"mark"`
	Quality              *string   `json:"quality"`
	Size                 string    `json:"size"`
	QuantityRemaining    float64   `json:"quantity_remaining"`
	Unit                 string    `json:"unit"`
	PurchasePricePerUnit *float64  `json:"purchase_price_per_unit"`
	ReceivedAt           time.Time `json:"received_at"`
}

type combinedRecentSale struct {
	SaleID              string    `json:"sale_id"`
	SoldAt              time.Time `json:"sold_at"`
	ReceivedAt          time.Time `json:"received_at"`
	Fruit               string    `json:"fruit"`
	Mark                string    `json:"mark"`
	Quality             *string   `json:"quality"`
	Size                string    `json:"size"`
	Quantity            float64   `json:"quantity"`
	Unit                string    `json:"unit"`
	SellingPricePerUnit float64   `json:"selling_price_per_unit"`
	TotalSaleAmount     float64   `json:"total_sale_amount"`
	GrossProfit         float64   `json:"gross_profit"`
}

type combinedNotifications struct {
	UnpricedBatchCount int `json:"unpriced_batch_count"`
}

type combinedResponse struct {
	Account       auth.Principal          `json:"account"`
	Summary       combinedSummary         `json:"summary"`
	Inventory     []combinedInventoryItem `json:"inventory"`
	RecentSales   []combinedRecentSale    `json:"recent_sales"`
	Notifications combinedNotifications   `json:"notifications"`
}

// Combined returns everything needed for the mobile dashboard in one HTTP
// request and one PostgreSQL batch round trip.
func (h *Handler) Combined(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.ShopID == "" {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	timezone := principal.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	started := time.Now()
	batch := &pgx.Batch{}
	defer func() {
		duration := time.Since(started)
		requestlog.AddAttributes(r.Context(), "dashboard_db_ms", duration.Milliseconds(), "dashboard_query_count", batch.Len())
		if duration >= requestlog.SlowThreshold() {
			slog.Warn("slow dashboard data load", "request_id", requestlog.RequestIDFromContext(r.Context()), "shop_id", principal.ShopID, "duration_ms", duration.Milliseconds(), "query_count", batch.Len())
		}
	}()
	batch.Queue(`SELECT
		COALESCE((SELECT SUM(quantity_remaining) FROM inventory_batches WHERE shop_id=$1),0),
		COALESCE((SELECT SUM(si.total_sale_amount) FROM sale_items si JOIN sales s ON s.id=si.sale_id WHERE s.shop_id=$1 AND (s.sold_at AT TIME ZONE $2)::date=(NOW() AT TIME ZONE $2)::date),0),
		COALESCE((SELECT SUM(si.gross_profit) FROM sale_items si JOIN sales s ON s.id=si.sale_id WHERE s.shop_id=$1 AND (s.sold_at AT TIME ZONE $2)::date=(NOW() AT TIME ZONE $2)::date),0)`, principal.ShopID, timezone)
	batch.Queue(`SELECT batch_id,fruit,supplier,mark,quality,size,quantity_remaining,unit,purchase_price_per_unit,received_at
		FROM current_inventory WHERE shop_id=$1 ORDER BY received_at DESC`, principal.ShopID)
	batch.Queue(`SELECT s.id,s.sold_at,ib.received_at,f.name,sup.mark,ib.quality,ib.size,si.quantity,ib.unit,si.selling_price_per_unit,si.total_sale_amount,si.gross_profit
		FROM sales s JOIN sale_items si ON si.sale_id=s.id JOIN inventory_batches ib ON ib.id=si.batch_id JOIN fruits f ON f.id=ib.fruit_id JOIN suppliers sup ON sup.id=ib.supplier_id
		WHERE s.shop_id=$1 ORDER BY s.sold_at DESC LIMIT 10`, principal.ShopID)
	batch.Queue(`SELECT COUNT(*) FROM inventory_batches WHERE shop_id=$1 AND purchase_price_per_unit IS NULL`, principal.ShopID)

	results := h.db.SendBatch(r.Context(), batch)
	response := combinedResponse{
		Account:     principal,
		Inventory:   make([]combinedInventoryItem, 0),
		RecentSales: make([]combinedRecentSale, 0),
	}
	if err := results.QueryRow().Scan(&response.Summary.TotalInventoryQuantity, &response.Summary.TodaySales, &response.Summary.TodayGrossProfit); err != nil {
		_ = results.Close()
		writeError(w, "could not load dashboard", http.StatusInternalServerError)
		return
	}

	inventoryRows, err := results.Query()
	if err != nil {
		_ = results.Close()
		writeError(w, "could not load dashboard", http.StatusInternalServerError)
		return
	}
	for inventoryRows.Next() {
		var item combinedInventoryItem
		if err := inventoryRows.Scan(&item.BatchID, &item.Fruit, &item.Supplier, &item.Mark, &item.Quality, &item.Size, &item.QuantityRemaining, &item.Unit, &item.PurchasePricePerUnit, &item.ReceivedAt); err != nil {
			inventoryRows.Close()
			_ = results.Close()
			writeError(w, "could not load dashboard", http.StatusInternalServerError)
			return
		}
		response.Inventory = append(response.Inventory, item)
	}
	if err := inventoryRows.Err(); err != nil {
		inventoryRows.Close()
		_ = results.Close()
		writeError(w, "could not load dashboard", http.StatusInternalServerError)
		return
	}
	inventoryRows.Close()

	salesRows, err := results.Query()
	if err != nil {
		_ = results.Close()
		writeError(w, "could not load dashboard", http.StatusInternalServerError)
		return
	}
	for salesRows.Next() {
		var sale combinedRecentSale
		if err := salesRows.Scan(&sale.SaleID, &sale.SoldAt, &sale.ReceivedAt, &sale.Fruit, &sale.Mark, &sale.Quality, &sale.Size, &sale.Quantity, &sale.Unit, &sale.SellingPricePerUnit, &sale.TotalSaleAmount, &sale.GrossProfit); err != nil {
			salesRows.Close()
			_ = results.Close()
			writeError(w, "could not load dashboard", http.StatusInternalServerError)
			return
		}
		response.RecentSales = append(response.RecentSales, sale)
	}
	if err := salesRows.Err(); err != nil {
		salesRows.Close()
		_ = results.Close()
		writeError(w, "could not load dashboard", http.StatusInternalServerError)
		return
	}
	salesRows.Close()

	if err := results.QueryRow().Scan(&response.Notifications.UnpricedBatchCount); err != nil {
		_ = results.Close()
		writeError(w, "could not load dashboard", http.StatusInternalServerError)
		return
	}
	if err := results.Close(); err != nil {
		writeError(w, "could not load dashboard", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, response)
}
