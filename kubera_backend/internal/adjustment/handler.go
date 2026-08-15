package adjustment

import (
	"encoding/json"
	"errors"
	"net/http"

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

type CreateRequest struct {
	BatchID        string  `json:"batch_id"`
	AdjustmentType string  `json:"adjustment_type"`
	Quantity       float64 `json:"quantity"`
	Reason         string  `json:"reason"`
}

type Response struct {
	ID             string  `json:"id"`
	BatchID        string  `json:"batch_id"`
	AdjustmentType string  `json:"adjustment_type"`
	Quantity       float64 `json:"quantity"`
	Reason         string  `json:"reason,omitempty"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	if req.BatchID == "" {
		writeError(w, "batch_id is required", http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		writeError(w, "quantity must be greater than 0", http.StatusBadRequest)
		return
	}

	validTypes := map[string]bool{
		"spoilage":            true,
		"damage":              true,
		"missing":             true,
		"supplier_return":     true,
		"customer_return":     true,
		"correction_increase": true,
		"correction_decrease": true,
	}

	if !validTypes[req.AdjustmentType] {
		writeError(w, "invalid adjustment_type", http.StatusBadRequest)
		return
	}

	tx, err := h.db.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		writeError(w, "could not start transaction", http.StatusInternalServerError)
		return
	}

	defer tx.Rollback(r.Context())

	var (
		batchShopID       string
		quantityRemaining float64
		quantityReceived  float64
	)

	err = tx.QueryRow(
		r.Context(),
		`
		SELECT
			shop_id,
			quantity_remaining,
			quantity_received
		FROM inventory_batches
		WHERE id = $1
		FOR UPDATE
		`,
		req.BatchID,
	).Scan(
		&batchShopID,
		&quantityRemaining,
		&quantityReceived,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, "batch not found", http.StatusNotFound)
		return
	}

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if batchShopID != shopID {
		writeError(w, "batch does not belong to this shop", http.StatusBadRequest)
		return
	}

	increase := req.AdjustmentType == "customer_return" ||
		req.AdjustmentType == "correction_increase"

	if increase {
		if quantityRemaining+req.Quantity > quantityReceived {
			writeError(
				w,
				"adjustment would exceed original received quantity",
				http.StatusBadRequest,
			)
			return
		}
	} else {
		if req.Quantity > quantityRemaining {
			writeError(
				w,
				"adjustment quantity exceeds remaining stock",
				http.StatusBadRequest,
			)
			return
		}
	}

	var response Response

	err = tx.QueryRow(
		r.Context(),
		`
		INSERT INTO inventory_adjustments (
			shop_id,
			batch_id,
			adjustment_type,
			quantity,
			reason
		)
		VALUES (
			$1, $2, $3, $4, NULLIF($5, '')
		)
		RETURNING
			id,
			batch_id,
			adjustment_type,
			quantity,
			COALESCE(reason, '')
		`,
		shopID,
		req.BatchID,
		req.AdjustmentType,
		req.Quantity,
		req.Reason,
	).Scan(
		&response.ID,
		&response.BatchID,
		&response.AdjustmentType,
		&response.Quantity,
		&response.Reason,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	operator := "-"

	if increase {
		operator = "+"
	}

	_, err = tx.Exec(
		r.Context(),
		`
		UPDATE inventory_batches
		SET quantity_remaining = quantity_remaining `+operator+` $1
		WHERE id = $2
		`,
		req.Quantity,
		req.BatchID,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, "could not save adjustment", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, response)
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
			ia.id,
			ia.adjustment_type,
			ia.quantity,
			COALESCE(ia.reason, ''),
			ia.adjusted_at,
			f.name,
			s.mark,
			ib.quality,
			ib.size,
			ib.unit
		FROM inventory_adjustments ia
		JOIN inventory_batches ib
			ON ib.id = ia.batch_id
		JOIN fruits f
			ON f.id = ib.fruit_id
		JOIN suppliers s
			ON s.id = ib.supplier_id
		WHERE ia.shop_id = $1
		ORDER BY ia.adjusted_at DESC
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
			id             string
			adjustmentType string
			quantity       float64
			reason         string
			adjustedAt     any
			fruit          string
			mark           string
			quality        *string
			size           string
			unit           string
		)

		if err := rows.Scan(
			&id,
			&adjustmentType,
			&quantity,
			&reason,
			&adjustedAt,
			&fruit,
			&mark,
			&quality,
			&size,
			&unit,
		); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		results = append(results, map[string]any{
			"id":              id,
			"adjustment_type": adjustmentType,
			"quantity":        quantity,
			"reason":          reason,
			"adjusted_at":     adjustedAt,
			"fruit":           fruit,
			"mark":            mark,
			"quality":         quality,
			"size":            size,
			"unit":            unit,
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
