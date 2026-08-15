package inventory

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"kubera_backend/internal/auth"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

type CreateBatchRequest struct {
	FruitID              string   `json:"fruit_id"`
	SupplierID           string   `json:"supplier_id"`
	Quality              *string  `json:"quality"`
	Size                 string   `json:"size"`
	Quantity             float64  `json:"quantity"`
	Unit                 string   `json:"unit"`
	PurchasePricePerUnit *float64 `json:"purchase_price_per_unit"`
}

type QuickCreateBatchRequest struct {
	FruitID              string   `json:"fruit_id"`
	FruitName            string   `json:"fruit_name"`
	SupplierID           string   `json:"supplier_id"`
	SupplierName         string   `json:"supplier_name"`
	Mark                 string   `json:"mark"`
	Phone                string   `json:"phone"`
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
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	if req.FruitID == "" ||
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
		SELECT
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$6,
			$7,
			$8
		WHERE EXISTS (
			SELECT 1 FROM fruits
			WHERE id = $2 AND shop_id = $1 AND is_active = TRUE
		)
		AND EXISTS (
			SELECT 1 FROM suppliers
			WHERE id = $3 AND shop_id = $1 AND is_active = TRUE
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
		shopID,
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

	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, "fruit or supplier not found for this shop", http.StatusBadRequest)
		return
	}

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, batch)
}

func (h *Handler) CreateQuickBatch(w http.ResponseWriter, r *http.Request) {
	var req QuickCreateBatchRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	req.FruitName = strings.Join(strings.Fields(req.FruitName), " ")
	req.SupplierName = strings.Join(strings.Fields(req.SupplierName), " ")
	req.Mark = strings.Join(strings.Fields(req.Mark), " ")
	req.Unit = strings.ToLower(strings.TrimSpace(req.Unit))
	if req.Unit == "" {
		req.Unit = "box"
	}
	if req.Size == "" {
		req.Size = "normal"
	}
	validUnit := map[string]bool{"box": true, "kg": true, "piece": true, "crate": true, "dozen": true}
	if req.Quantity <= 0 || req.PurchasePricePerUnit == nil || *req.PurchasePricePerUnit < 0 || !validUnit[req.Unit] || (req.FruitID == "" && req.FruitName == "") || (req.SupplierID == "" && req.Mark == "") {
		writeError(w, "fruit, mark, quantity, unit, and buy price are required", http.StatusBadRequest)
		return
	}
	if len([]rune(req.FruitName)) > 80 || len([]rune(req.Mark)) > 80 || len([]rune(req.SupplierName)) > 120 {
		writeError(w, "fruit or mark name is too long", http.StatusBadRequest)
		return
	}

	tx, err := h.db.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		writeError(w, "could not start purchase", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	fruitID := req.FruitID
	if fruitID != "" {
		err = tx.QueryRow(r.Context(), `SELECT id FROM fruits WHERE id = $1 AND shop_id = $2 AND is_active = TRUE`, fruitID, shopID).Scan(&fruitID)
	} else {
		err = tx.QueryRow(r.Context(), `SELECT id FROM fruits WHERE shop_id = $1 AND lower(name) = lower($2) LIMIT 1`, shopID, req.FruitName).Scan(&fruitID)
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(r.Context(), `INSERT INTO fruits (shop_id, name, default_unit) VALUES ($1, $2, $3) RETURNING id`, shopID, req.FruitName, req.Unit).Scan(&fruitID)
		} else if err == nil {
			_, err = tx.Exec(r.Context(), `UPDATE fruits SET is_active = TRUE WHERE id = $1`, fruitID)
		}
	}
	if err != nil {
		writeError(w, "could not resolve fruit", http.StatusBadRequest)
		return
	}

	supplierID := req.SupplierID
	if supplierID != "" {
		err = tx.QueryRow(r.Context(), `SELECT id FROM suppliers WHERE id = $1 AND shop_id = $2 AND is_active = TRUE`, supplierID, shopID).Scan(&supplierID)
	} else {
		err = tx.QueryRow(r.Context(), `SELECT id FROM suppliers WHERE shop_id = $1 AND lower(mark) = lower($2) LIMIT 1`, shopID, req.Mark).Scan(&supplierID)
		if errors.Is(err, pgx.ErrNoRows) {
			if req.SupplierName == "" {
				req.SupplierName = req.Mark
			}
			err = tx.QueryRow(r.Context(), `INSERT INTO suppliers (shop_id, name, mark, phone) VALUES ($1, $2, $3, NULLIF($4, '')) RETURNING id`, shopID, req.SupplierName, req.Mark, strings.TrimSpace(req.Phone)).Scan(&supplierID)
		} else if err == nil {
			_, err = tx.Exec(r.Context(), `UPDATE suppliers SET is_active = TRUE WHERE id = $1`, supplierID)
		}
	}
	if err != nil {
		writeError(w, "could not resolve supplier mark", http.StatusBadRequest)
		return
	}

	var batch Batch
	err = tx.QueryRow(r.Context(), `INSERT INTO inventory_batches (shop_id, fruit_id, supplier_id, quality, size, quantity_received, quantity_remaining, unit, purchase_price_per_unit)
		VALUES ($1,$2,$3,$4,$5,$6,$6,$7,$8)
		RETURNING id, shop_id, fruit_id, supplier_id, quality, size, quantity_received, quantity_remaining, unit, purchase_price_per_unit, received_at`,
		shopID, fruitID, supplierID, req.Quality, req.Size, req.Quantity, req.Unit, req.PurchasePricePerUnit).
		Scan(&batch.ID, &batch.ShopID, &batch.FruitID, &batch.SupplierID, &batch.Quality, &batch.Size, &batch.QuantityReceived, &batch.QuantityRemaining, &batch.Unit, &batch.PurchasePricePerUnit, &batch.ReceivedAt)
	if err != nil {
		writeError(w, "could not save purchased stock", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, "could not save purchased stock", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, batch)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
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
