package dashboard

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"kubera_backend/internal/auth"
)

type Handler struct {
	db *pgxpool.Pool
}

type closeDayRequest struct {
	AllowEstimates bool `json:"allow_estimates"`
}

func (h *Handler) CloseDay(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	date := r.PathValue("date")
	if _, err := time.Parse("2006-01-02", date); err != nil {
		writeError(w, "date must use YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	var req closeDayRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tz := p.Timezone
	if tz == "" {
		tz = "UTC"
	}
	var estimated int
	err := h.db.QueryRow(r.Context(), `SELECT COUNT(*) FROM sale_items si JOIN sales s ON s.id=si.sale_id WHERE s.shop_id=$1 AND (s.sold_at AT TIME ZONE $2)::date=$3::date AND si.cost_price_is_estimated`, p.ShopID, tz, date).Scan(&estimated)
	if err != nil {
		writeError(w, "could not close day", 500)
		return
	}
	if estimated > 0 && !req.AllowEstimates {
		writeError(w, "buying prices are required or estimates must be accepted", http.StatusConflict)
		return
	}
	_, err = h.db.Exec(r.Context(), `INSERT INTO shop_daily_closings(shop_id,business_date,sales_transactions,sales_revenue,gross_profit,purchase_value,spoilage_loss,used_estimates)
	SELECT $1,$3::date,
	 (SELECT COUNT(*) FROM sales WHERE shop_id=$1 AND (sold_at AT TIME ZONE $2)::date=$3::date),
	 COALESCE((SELECT SUM(si.total_sale_amount) FROM sale_items si JOIN sales s ON s.id=si.sale_id WHERE s.shop_id=$1 AND (s.sold_at AT TIME ZONE $2)::date=$3::date),0),
	 COALESCE((SELECT SUM(si.gross_profit) FROM sale_items si JOIN sales s ON s.id=si.sale_id WHERE s.shop_id=$1 AND (s.sold_at AT TIME ZONE $2)::date=$3::date),0),
	 COALESCE((SELECT SUM(quantity_received*COALESCE(purchase_price_per_unit,0)) FROM inventory_batches WHERE shop_id=$1 AND (received_at AT TIME ZONE $2)::date=$3::date),0),
	 COALESCE((SELECT SUM(ia.quantity*COALESCE(ib.purchase_price_per_unit,0)) FROM inventory_adjustments ia JOIN inventory_batches ib ON ib.id=ia.batch_id WHERE ia.shop_id=$1 AND ia.adjustment_type IN ('spoilage','damage') AND (ia.adjusted_at AT TIME ZONE $2)::date=$3::date),0),$4
	ON CONFLICT(shop_id,business_date) DO UPDATE SET sales_transactions=EXCLUDED.sales_transactions,sales_revenue=EXCLUDED.sales_revenue,gross_profit=EXCLUDED.gross_profit,purchase_value=EXCLUDED.purchase_value,spoilage_loss=EXCLUDED.spoilage_loss,used_estimates=EXCLUDED.used_estimates,closed_at=NOW()`, p.ShopID, tz, date, estimated > 0)
	if err != nil {
		writeError(w, "could not close day", 500)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "closed", "date": date, "used_estimates": estimated > 0})
}

func (h *Handler) ClosingHistory(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", 401)
		return
	}
	rows, err := h.db.Query(r.Context(), `SELECT business_date,sales_transactions,sales_revenue,gross_profit,purchase_value,spoilage_loss,used_estimates,closed_at FROM shop_daily_closings WHERE shop_id=$1 ORDER BY business_date DESC LIMIT 60`, shopID)
	if err != nil {
		writeError(w, "could not load closing history", 500)
		return
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var date time.Time
		var transactions int
		var sales, profit, purchases, loss float64
		var estimated bool
		var closedAt time.Time
		if err := rows.Scan(&date, &transactions, &sales, &profit, &purchases, &loss, &estimated, &closedAt); err != nil {
			writeError(w, "could not load closing history", 500)
			return
		}
		items = append(items, map[string]any{"date": date, "sales_transactions": transactions, "sales_revenue": sales, "gross_profit": profit, "purchase_value": purchases, "spoilage_loss": loss, "used_estimates": estimated, "closed_at": closedAt})
	}
	writeJSON(w, 200, items)
}

func (h *Handler) DailyReport(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.ShopID == "" {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	timezone := principal.Timezone
	if timezone == "" {
		timezone = "UTC"
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		writeError(w, "invalid shop timezone", http.StatusInternalServerError)
		return
	}
	reportDate := r.URL.Query().Get("date")
	if reportDate == "" {
		reportDate = time.Now().In(location).Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", reportDate); err != nil {
		writeError(w, "date must use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	type Line struct {
		BatchID                string     `json:"batch_id"`
		Fruit                  string     `json:"fruit"`
		Mark                   string     `json:"mark"`
		Quality                string     `json:"quality"`
		Size                   string     `json:"size"`
		Unit                   string     `json:"unit"`
		ReceivedAt             time.Time  `json:"received_at"`
		FirstSoldAt            *time.Time `json:"first_sold_at"`
		LastSoldAt             *time.Time `json:"last_sold_at"`
		PurchasedQuantity      float64    `json:"purchased_quantity"`
		PurchasePricePerUnit   *float64   `json:"purchase_price_per_unit"`
		PurchaseValue          float64    `json:"purchase_value"`
		SoldQuantity           float64    `json:"sold_quantity"`
		AverageSellingPrice    *float64   `json:"average_selling_price"`
		SalesRevenue           float64    `json:"sales_revenue"`
		GrossProfit            float64    `json:"gross_profit"`
		ClosingQuantity        float64    `json:"closing_quantity"`
		ProfitIsEstimated      bool       `json:"profit_is_estimated"`
		SuggestedPurchasePrice *float64   `json:"suggested_purchase_price"`
	}
	type Report struct {
		Date                   string  `json:"date"`
		StockBatchesAdded      int     `json:"stock_batches_added"`
		PurchaseValue          float64 `json:"purchase_value"`
		SaleCount              int     `json:"sale_count"`
		SalesRevenue           float64 `json:"sales_revenue"`
		GrossProfit            float64 `json:"gross_profit"`
		SpoilageLoss           float64 `json:"spoilage_loss"`
		EstimatedSaleItemCount int     `json:"estimated_sale_item_count"`
		UnpricedBatchCount     int     `json:"unpriced_batch_count"`
		ClosingBatchCount      int     `json:"closing_batch_count"`
		Lines                  []Line  `json:"lines"`
	}
	result := Report{Date: reportDate, Lines: make([]Line, 0)}
	rows, err := h.db.Query(r.Context(), `WITH bounds AS (
		SELECT ($3::date::timestamp AT TIME ZONE $2) AS start_at, (($3::date + 1)::timestamp AT TIME ZONE $2) AS end_at
	), sales_by_batch AS (
		SELECT si.batch_id,
		 COALESCE(SUM(si.quantity) FILTER (WHERE s.sold_at>=b.start_at AND s.sold_at<b.end_at),0) sold_today,
		 COALESCE(SUM(si.total_sale_amount) FILTER (WHERE s.sold_at>=b.start_at AND s.sold_at<b.end_at),0) revenue_today,
		 COALESCE(SUM(si.gross_profit) FILTER (WHERE s.sold_at>=b.start_at AND s.sold_at<b.end_at),0) profit_today,
		 COALESCE(SUM(si.quantity) FILTER (WHERE s.sold_at<b.end_at),0) sold_through_day,
		 BOOL_OR(si.cost_price_is_estimated AND s.sold_at>=b.start_at AND s.sold_at<b.end_at) estimated,
		 MIN(s.sold_at) FILTER (WHERE s.sold_at>=b.start_at AND s.sold_at<b.end_at) first_sold_at,
		 MAX(s.sold_at) FILTER (WHERE s.sold_at>=b.start_at AND s.sold_at<b.end_at) last_sold_at,
		 COUNT(*) FILTER (WHERE si.cost_price_is_estimated AND s.sold_at>=b.start_at AND s.sold_at<b.end_at) estimated_count
		FROM sale_items si JOIN sales s ON s.id=si.sale_id CROSS JOIN bounds b WHERE s.shop_id=$1 AND s.sold_at<b.end_at GROUP BY si.batch_id
	), adjustments AS (
		SELECT ia.batch_id, COALESCE(SUM(CASE WHEN ia.adjustment_type IN ('customer_return','correction_increase') THEN ia.quantity ELSE -ia.quantity END),0) net
		FROM inventory_adjustments ia CROSS JOIN bounds b WHERE ia.shop_id=$1 AND ia.adjusted_at<b.end_at GROUP BY ia.batch_id
	)
	SELECT ib.id,f.name,s.mark,COALESCE(ib.quality,''),ib.size,ib.unit,ib.received_at,sb.first_sold_at,sb.last_sold_at,
	 CASE WHEN ib.received_at>=b.start_at THEN ib.quantity_received ELSE 0 END,
	 ib.purchase_price_per_unit,
	 CASE WHEN ib.received_at>=b.start_at THEN ib.quantity_received*COALESCE(ib.purchase_price_per_unit,0) ELSE 0 END,
	 COALESCE(sb.sold_today,0), CASE WHEN sb.sold_today>0 THEN sb.revenue_today/sb.sold_today END,
	 COALESCE(sb.revenue_today,0),COALESCE(sb.profit_today,0),
	 GREATEST(ib.quantity_received-COALESCE(sb.sold_through_day,0)+COALESCE(a.net,0),0),COALESCE(sb.estimated,FALSE),
	 CASE WHEN ib.purchase_price_per_unit IS NULL AND sb.sold_today>0 THEN (sb.revenue_today/sb.sold_today)*0.94 END,
	 COALESCE(sb.estimated_count,0)
	FROM inventory_batches ib JOIN fruits f ON f.id=ib.fruit_id JOIN suppliers s ON s.id=ib.supplier_id CROSS JOIN bounds b
	LEFT JOIN sales_by_batch sb ON sb.batch_id=ib.id LEFT JOIN adjustments a ON a.batch_id=ib.id
	WHERE ib.shop_id=$1 AND ib.received_at<b.end_at AND (ib.received_at>=b.start_at OR COALESCE(sb.sold_today,0)>0 OR GREATEST(ib.quantity_received-COALESCE(sb.sold_through_day,0)+COALESCE(a.net,0),0)>0)
	ORDER BY lower(f.name),lower(s.mark),ib.received_at`, principal.ShopID, timezone, reportDate)
	if err != nil {
		writeError(w, "could not load daily report", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var line Line
		var estimatedCount int
		if err := rows.Scan(&line.BatchID, &line.Fruit, &line.Mark, &line.Quality, &line.Size, &line.Unit, &line.ReceivedAt, &line.FirstSoldAt, &line.LastSoldAt, &line.PurchasedQuantity, &line.PurchasePricePerUnit, &line.PurchaseValue, &line.SoldQuantity, &line.AverageSellingPrice, &line.SalesRevenue, &line.GrossProfit, &line.ClosingQuantity, &line.ProfitIsEstimated, &line.SuggestedPurchasePrice, &estimatedCount); err != nil {
			writeError(w, "could not load daily report", http.StatusInternalServerError)
			return
		}
		result.Lines = append(result.Lines, line)
		result.PurchaseValue += line.PurchaseValue
		result.SalesRevenue += line.SalesRevenue
		result.GrossProfit += line.GrossProfit
		result.EstimatedSaleItemCount += estimatedCount
		if line.PurchasedQuantity > 0 {
			result.StockBatchesAdded++
		}
		if line.ClosingQuantity > 0 {
			result.ClosingBatchCount++
		}
		if line.PurchasePricePerUnit == nil && (line.PurchasedQuantity > 0 || line.SoldQuantity > 0) {
			result.UnpricedBatchCount++
		}
	}
	if err := rows.Err(); err != nil {
		writeError(w, "could not load daily report", http.StatusInternalServerError)
		return
	}
	if err := h.db.QueryRow(r.Context(), `SELECT
		(SELECT COUNT(*) FROM sales WHERE shop_id=$1 AND (sold_at AT TIME ZONE $2)::date=$3::date),
		COALESCE((SELECT SUM(ia.quantity*COALESCE(ib.purchase_price_per_unit,0)) FROM inventory_adjustments ia JOIN inventory_batches ib ON ib.id=ia.batch_id WHERE ia.shop_id=$1 AND ia.adjustment_type IN ('spoilage','damage') AND (ia.adjusted_at AT TIME ZONE $2)::date=$3::date),0)`, principal.ShopID, timezone, reportDate).Scan(&result.SaleCount, &result.SpoilageLoss); err != nil {
		writeError(w, "could not load daily report", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) NotificationSummary(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	var count int
	if err := h.db.QueryRow(r.Context(), `SELECT COUNT(*) FROM inventory_batches WHERE shop_id=$1 AND purchase_price_per_unit IS NULL`, shopID).Scan(&count); err != nil {
		writeError(w, "could not load notifications", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"unpriced_batch_count": count})
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.ShopID == "" {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	shopID := principal.ShopID
	timezone := principal.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	type Summary struct {
		TotalInventoryQuantity float64 `json:"total_inventory_quantity"`
		TodaySales             float64 `json:"today_sales"`
		TodayGrossProfit       float64 `json:"today_gross_profit"`
	}

	var result Summary

	err := h.db.QueryRow(
		r.Context(),
		`
		SELECT
			COALESCE(
				(
					SELECT SUM(quantity_remaining)
					FROM inventory_batches
					WHERE shop_id = $1
				),
				0
			),

			COALESCE(
				(
					SELECT SUM(si.total_sale_amount)
					FROM sale_items si
					JOIN sales s
						ON s.id = si.sale_id
					WHERE s.shop_id = $1
					  AND (s.sold_at AT TIME ZONE $2)::date =
						  (NOW() AT TIME ZONE $2)::date
				),
				0
			),

			COALESCE(
				(
					SELECT SUM(si.gross_profit)
					FROM sale_items si
					JOIN sales s
						ON s.id = si.sale_id
					WHERE s.shop_id = $1
					  AND (s.sold_at AT TIME ZONE $2)::date =
						  (NOW() AT TIME ZONE $2)::date
				),
				0
			)
		`,
		shopID,
		timezone,
	).Scan(
		&result.TotalInventoryQuantity,
		&result.TodaySales,
		&result.TodayGrossProfit,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) StockByFruit(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	rows, err := h.db.Query(
		r.Context(),
		`
		SELECT
			f.name,
			ib.unit,
			SUM(ib.quantity_remaining)
		FROM inventory_batches ib
		JOIN fruits f
			ON f.id = ib.fruit_id
		WHERE ib.shop_id = $1
		  AND ib.quantity_remaining > 0
		GROUP BY f.name, ib.unit
		ORDER BY f.name
		`,
		shopID,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := make([]map[string]any, 0)

	for rows.Next() {
		var (
			fruit    string
			unit     string
			quantity float64
		)

		if err := rows.Scan(&fruit, &unit, &quantity); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		results = append(results, map[string]any{
			"fruit":    fruit,
			"quantity": quantity,
			"unit":     unit,
		})
	}

	writeJSON(w, http.StatusOK, results)
}

func (h *Handler) StockBySupplier(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	rows, err := h.db.Query(
		r.Context(),
		`
		SELECT
			s.id,
			s.name,
			s.mark,
			ib.unit,
			SUM(ib.quantity_remaining)
		FROM inventory_batches ib
		JOIN suppliers s
			ON s.id = ib.supplier_id
		WHERE ib.shop_id = $1
		  AND ib.quantity_remaining > 0
		GROUP BY
			s.id,
			s.name,
			s.mark,
			ib.unit
		ORDER BY s.mark
		`,
		shopID,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := make([]map[string]any, 0)

	for rows.Next() {
		var (
			supplierID   string
			supplierName string
			mark         string
			unit         string
			quantity     float64
		)

		if err := rows.Scan(
			&supplierID,
			&supplierName,
			&mark,
			&unit,
			&quantity,
		); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		results = append(results, map[string]any{
			"supplier_id": supplierID,
			"supplier":    supplierName,
			"mark":        mark,
			"quantity":    quantity,
			"unit":        unit,
		})
	}

	writeJSON(w, http.StatusOK, results)
}

func (h *Handler) RecentSales(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	rows, err := h.db.Query(
		r.Context(),
		`
		SELECT
			s.id,
			s.sold_at,
			ib.received_at,
			f.name,
			sup.mark,
			ib.quality,
			ib.size,
			si.quantity,
			ib.unit,
			si.selling_price_per_unit,
			si.total_sale_amount,
			si.gross_profit
		FROM sales s
		JOIN sale_items si
			ON si.sale_id = s.id
		JOIN inventory_batches ib
			ON ib.id = si.batch_id
		JOIN fruits f
			ON f.id = ib.fruit_id
		JOIN suppliers sup
			ON sup.id = ib.supplier_id
		WHERE s.shop_id = $1
		ORDER BY s.sold_at DESC
		LIMIT 10
		`,
		shopID,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := make([]map[string]any, 0)

	for rows.Next() {
		var (
			saleID          string
			soldAt          any
			receivedAt      any
			fruit           string
			mark            string
			quality         *string
			size            string
			quantity        float64
			unit            string
			price           float64
			totalSaleAmount float64
			grossProfit     float64
		)

		if err := rows.Scan(
			&saleID,
			&soldAt,
			&receivedAt,
			&fruit,
			&mark,
			&quality,
			&size,
			&quantity,
			&unit,
			&price,
			&totalSaleAmount,
			&grossProfit,
		); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		results = append(results, map[string]any{
			"sale_id":                saleID,
			"sold_at":                soldAt,
			"received_at":            receivedAt,
			"fruit":                  fruit,
			"mark":                   mark,
			"quality":                quality,
			"size":                   size,
			"quantity":               quantity,
			"unit":                   unit,
			"selling_price_per_unit": price,
			"total_sale_amount":      totalSaleAmount,
			"gross_profit":           grossProfit,
		})
	}

	writeJSON(w, http.StatusOK, results)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
