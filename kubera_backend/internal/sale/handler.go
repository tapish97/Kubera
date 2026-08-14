package sale

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

type SaleItemRequest struct {
	BatchID             string  `json:"batch_id"`
	Quantity            float64 `json:"quantity"`
	SellingPricePerUnit float64 `json:"selling_price_per_unit"`
}

type CreateSaleRequest struct {
	ShopID string            `json:"shop_id"`
	Items  []SaleItemRequest `json:"items"`
	Notes  string            `json:"notes"`
}

type SaleItemResponse struct {
	ID                  string  `json:"id"`
	BatchID             string  `json:"batch_id"`
	Quantity            float64 `json:"quantity"`
	CostPricePerUnit    float64 `json:"cost_price_per_unit"`
	SellingPricePerUnit float64 `json:"selling_price_per_unit"`
	TotalSaleAmount     float64 `json:"total_sale_amount"`
	GrossProfit         float64 `json:"gross_profit"`
}

type SaleResponse struct {
	ID     string             `json:"id"`
	ShopID string             `json:"shop_id"`
	Items  []SaleItemResponse `json:"items"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSaleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ShopID == "" {
		writeError(w, "shop_id is required", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		writeError(w, "at least one sale item is required", http.StatusBadRequest)
		return
	}

	for _, item := range req.Items {
		if item.BatchID == "" {
			writeError(w, "batch_id is required", http.StatusBadRequest)
			return
		}

		if item.Quantity <= 0 {
			writeError(w, "quantity must be greater than 0", http.StatusBadRequest)
			return
		}

		if item.SellingPricePerUnit < 0 {
			writeError(w, "selling price cannot be negative", http.StatusBadRequest)
			return
		}
	}

	tx, err := h.db.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		writeError(w, "could not start transaction", http.StatusInternalServerError)
		return
	}

	defer tx.Rollback(r.Context())

	var saleID string

	err = tx.QueryRow(
		r.Context(),
		`
		INSERT INTO sales (
			shop_id,
			notes
		)
		VALUES ($1, NULLIF($2, ''))
		RETURNING id
		`,
		req.ShopID,
		req.Notes,
	).Scan(&saleID)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := SaleResponse{
		ID:     saleID,
		ShopID: req.ShopID,
		Items:  make([]SaleItemResponse, 0),
	}

	for _, item := range req.Items {

		var (
			quantityRemaining float64
			costPrice         *float64
			batchShopID       string
		)

		// FOR UPDATE locks this batch until the transaction finishes.
		err := tx.QueryRow(
			r.Context(),
			`
			SELECT
				shop_id,
				quantity_remaining,
				purchase_price_per_unit
			FROM inventory_batches
			WHERE id = $1
			FOR UPDATE
			`,
			item.BatchID,
		).Scan(
			&batchShopID,
			&quantityRemaining,
			&costPrice,
		)

		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, "inventory batch not found", http.StatusNotFound)
			return
		}

		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Prevent one shop from selling another shop's stock.
		if batchShopID != req.ShopID {
			writeError(w, "batch does not belong to this shop", http.StatusBadRequest)
			return
		}

		if quantityRemaining < item.Quantity {
			writeError(
				w,
				"not enough inventory available for batch "+item.BatchID,
				http.StatusBadRequest,
			)
			return
		}

		if costPrice == nil {
			writeError(
				w,
				"batch does not have a purchase price",
				http.StatusBadRequest,
			)
			return
		}

		var saleItem SaleItemResponse

		err = tx.QueryRow(
			r.Context(),
			`
			INSERT INTO sale_items (
				sale_id,
				batch_id,
				quantity,
				cost_price_per_unit,
				selling_price_per_unit
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING
				id,
				batch_id,
				quantity,
				cost_price_per_unit,
				selling_price_per_unit,
				total_sale_amount,
				gross_profit
			`,
			saleID,
			item.BatchID,
			item.Quantity,
			*costPrice,
			item.SellingPricePerUnit,
		).Scan(
			&saleItem.ID,
			&saleItem.BatchID,
			&saleItem.Quantity,
			&saleItem.CostPricePerUnit,
			&saleItem.SellingPricePerUnit,
			&saleItem.TotalSaleAmount,
			&saleItem.GrossProfit,
		)

		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(
			r.Context(),
			`
			UPDATE inventory_batches
			SET quantity_remaining = quantity_remaining - $1
			WHERE id = $2
			`,
			item.Quantity,
			item.BatchID,
		)

		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response.Items = append(response.Items, saleItem)
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, "could not save sale", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	shopID := r.URL.Query().Get("shop_id")

	if shopID == "" {
		writeError(w, "shop_id is required", http.StatusBadRequest)
		return
	}

	rows, err := h.db.Query(
		r.Context(),
		`
		SELECT
			s.id,
			s.sold_at,
			f.name,
			sup.name,
			sup.mark,
			ib.quality,
			ib.size,
			si.quantity,
			ib.unit,
			si.cost_price_per_unit,
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
		`,
		shopID,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	type SaleHistoryItem struct {
		SaleID              string  `json:"sale_id"`
		SoldAt              any     `json:"sold_at"`
		Fruit               string  `json:"fruit"`
		Supplier            string  `json:"supplier"`
		Mark                string  `json:"mark"`
		Quality             *string `json:"quality"`
		Size                string  `json:"size"`
		Quantity            float64 `json:"quantity"`
		Unit                string  `json:"unit"`
		CostPricePerUnit    float64 `json:"cost_price_per_unit"`
		SellingPricePerUnit float64 `json:"selling_price_per_unit"`
		TotalSaleAmount     float64 `json:"total_sale_amount"`
		GrossProfit         float64 `json:"gross_profit"`
	}

	results := make([]SaleHistoryItem, 0)

	for rows.Next() {
		var item SaleHistoryItem

		err := rows.Scan(
			&item.SaleID,
			&item.SoldAt,
			&item.Fruit,
			&item.Supplier,
			&item.Mark,
			&item.Quality,
			&item.Size,
			&item.Quantity,
			&item.Unit,
			&item.CostPricePerUnit,
			&item.SellingPricePerUnit,
			&item.TotalSaleAmount,
			&item.GrossProfit,
		)

		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		results = append(results, item)
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
