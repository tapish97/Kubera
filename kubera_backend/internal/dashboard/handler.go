package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	shopID := r.URL.Query().Get("shop_id")

	if shopID == "" {
		writeError(w, "shop_id is required", http.StatusBadRequest)
		return
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
					  AND (s.sold_at AT TIME ZONE 'Asia/Kolkata')::date =
						  (NOW() AT TIME ZONE 'Asia/Kolkata')::date
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
					  AND (s.sold_at AT TIME ZONE 'Asia/Kolkata')::date =
						  (NOW() AT TIME ZONE 'Asia/Kolkata')::date
				),
				0
			)
		`,
		shopID,
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
	shopID := r.URL.Query().Get("shop_id")

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
	shopID := r.URL.Query().Get("shop_id")

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
	shopID := r.URL.Query().Get("shop_id")

	rows, err := h.db.Query(
		r.Context(),
		`
		SELECT
			s.id,
			s.sold_at,
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
