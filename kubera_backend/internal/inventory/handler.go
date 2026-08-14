package inventory

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

type CreateBatchRequest struct {
	ShopID               string   `json:"shop_id"`
	FruitID              string   `json:"fruit_id"`
	SupplierID           string   `json:"supplier_id"`
	Quality              *string  `json:"quality"`
	Size                 string   `json:"size"`
	Quantity             float64  `json:"quantity"`
	Unit                 string   `json:"unit"`
	PurchasePricePerUnit *float64 `json:"purchase_price_per_unit"`
}

type Batch struct {
	ID                   string    `json:"id"`
	ShopID               string    `json:"shop_id"`
	FruitID              string    `json:"fruit_id"`
	SupplierID           string    `json:"supplier_id"`
	Quality              *string   `json:"quality"`
	Size                 string    `json:"size"`
	QuantityReceived     float64   `json:"quantity_received"`
	QuantityRemaining    float64   `json:"quantity_remaining"`
	Unit                 string    `json:"unit"`
	PurchasePricePerUnit *float64  `json:"purchase_price_per_unit"`
	ReceivedAt           time.Time `json:"received_at"`
}

func (h *Handler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var req CreateBatchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ShopID == "" ||
		req.FruitID == "" ||
		req.SupplierID == "" ||
		req.Quantity <= 0 {

		writeError(w, "missing required fields", http.StatusBadRequest)
		return
	}

	if req.Size == "" {
		req.Size = "normal"
	}

	if req.Unit == "" {
		req.Unit = "box"
	}

	var batch Batch

	err := h.db.QueryRow(
		r.Context(),
		`
		INSERT INTO inventory_batches (
			shop_id,
			fruit_id,
			supplier_id,
			quality,
			size,
			quantity_received,
			quantity_remaining,
			unit,
			purchase_price_per_unit
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$6,
			$7,
			$8
		)
		RETURNING
			id,
			shop_id,
			fruit_id,
			supplier_id,
			quality,
			size,
			quantity_received,
			quantity_remaining,
			unit,
			purchase_price_per_unit,
			received_at
		`,
		req.ShopID,
		req.FruitID,
		req.SupplierID,
		req.Quality,
		req.Size,
		req.Quantity,
		req.Unit,
		req.PurchasePricePerUnit,
	).Scan(
		&batch.ID,
		&batch.ShopID,
		&batch.FruitID,
		&batch.SupplierID,
		&batch.Quality,
		&batch.Size,
		&batch.QuantityReceived,
		&batch.QuantityRemaining,
		&batch.Unit,
		&batch.PurchasePricePerUnit,
		&batch.ReceivedAt,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, batch)
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
			batch_id,
			fruit,
			supplier,
			mark,
			quality,
			size,
			quantity_remaining,
			unit,
			purchase_price_per_unit,
			received_at
		FROM current_inventory
		WHERE shop_id = $1
		ORDER BY received_at DESC
		`,
		shopID,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	type InventoryItem struct {
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

	items := make([]InventoryItem, 0)

	for rows.Next() {
		var item InventoryItem

		if err := rows.Scan(
			&item.BatchID,
			&item.Fruit,
			&item.Supplier,
			&item.Mark,
			&item.Quality,
			&item.Size,
			&item.QuantityRemaining,
			&item.Unit,
			&item.PurchasePricePerUnit,
			&item.ReceivedAt,
		); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, items)
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
