package supplier

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

type Supplier struct {
	ID                  string `json:"id"`
	ShopID              string `json:"shop_id"`
	Name                string `json:"name"`
	Mark                string `json:"mark"`
	Phone               string `json:"phone,omitempty"`
	DefaultLeadTimeDays *int   `json:"default_lead_time_days,omitempty"`
}

type CreateRequest struct {
	ShopID              string `json:"shop_id"`
	Name                string `json:"name"`
	Mark                string `json:"mark"`
	Phone               string `json:"phone"`
	DefaultLeadTimeDays *int   `json:"default_lead_time_days"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ShopID == "" || req.Name == "" || req.Mark == "" {
		writeError(
			w,
			"shop_id, name and mark are required",
			http.StatusBadRequest,
		)
		return
	}

	var supplier Supplier

	err := h.db.QueryRow(
		r.Context(),
		`
		INSERT INTO suppliers (
			shop_id,
			name,
			mark,
			phone,
			default_lead_time_days
		)
		VALUES (
			$1,
			$2,
			$3,
			NULLIF($4, ''),
			$5
		)
		RETURNING
			id,
			shop_id,
			name,
			mark,
			COALESCE(phone, ''),
			default_lead_time_days
		`,
		req.ShopID,
		req.Name,
		req.Mark,
		req.Phone,
		req.DefaultLeadTimeDays,
	).Scan(
		&supplier.ID,
		&supplier.ShopID,
		&supplier.Name,
		&supplier.Mark,
		&supplier.Phone,
		&supplier.DefaultLeadTimeDays,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, supplier)
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
			id,
			shop_id,
			name,
			mark,
			COALESCE(phone, ''),
			default_lead_time_days
		FROM suppliers
		WHERE shop_id = $1
		  AND is_active = TRUE
		ORDER BY mark
		`,
		shopID,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	suppliers := make([]Supplier, 0)

	for rows.Next() {
		var supplier Supplier

		if err := rows.Scan(
			&supplier.ID,
			&supplier.ShopID,
			&supplier.Name,
			&supplier.Mark,
			&supplier.Phone,
			&supplier.DefaultLeadTimeDays,
		); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		suppliers = append(suppliers, supplier)
	}

	writeJSON(w, http.StatusOK, suppliers)
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
