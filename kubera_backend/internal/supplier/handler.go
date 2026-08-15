package supplier

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"kubera_backend/internal/auth"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

type Supplier struct {
	ID                  string   `json:"id"`
	ShopID              string   `json:"shop_id"`
	Name                string   `json:"name"`
	Mark                string   `json:"mark"`
	Phone               string   `json:"phone,omitempty"`
	DefaultLeadTimeDays *int     `json:"default_lead_time_days,omitempty"`
	LocationLabel       string   `json:"location_label"`
	Latitude            *float64 `json:"latitude"`
	Longitude           *float64 `json:"longitude"`
	Notes               string   `json:"notes"`
}

type CreateRequest struct {
	Name                string   `json:"name"`
	Mark                string   `json:"mark"`
	Phone               string   `json:"phone"`
	DefaultLeadTimeDays *int     `json:"default_lead_time_days"`
	LocationLabel       string   `json:"location_label"`
	Latitude            *float64 `json:"latitude"`
	Longitude           *float64 `json:"longitude"`
	Notes               string   `json:"notes"`
}

type UpdateRequest struct {
	Name          string   `json:"name"`
	Mark          string   `json:"mark"`
	Phone         string   `json:"phone"`
	LocationLabel string   `json:"location_label"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	Notes         string   `json:"notes"`
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

	if req.Name == "" || req.Mark == "" {
		writeError(
			w,
			"name and mark are required",
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
			, location_label, latitude, longitude, notes
		)
		VALUES (
			$1,
			$2,
			$3,
			NULLIF($4, ''),
			$5, NULLIF($6, ''), $7, $8, NULLIF($9, '')
		)
		RETURNING
			id,
			shop_id,
			name,
			mark,
			COALESCE(phone, ''),
			default_lead_time_days
			, COALESCE(location_label, ''), latitude, longitude, COALESCE(notes, '')
		`,
		shopID,
		req.Name,
		req.Mark,
		req.Phone,
		req.DefaultLeadTimeDays,
		req.LocationLabel,
		req.Latitude,
		req.Longitude,
		req.Notes,
	).Scan(
		&supplier.ID,
		&supplier.ShopID,
		&supplier.Name,
		&supplier.Mark,
		&supplier.Phone,
		&supplier.DefaultLeadTimeDays,
		&supplier.LocationLabel,
		&supplier.Latitude,
		&supplier.Longitude,
		&supplier.Notes,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, supplier)
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
			id,
			shop_id,
			name,
			mark,
			COALESCE(phone, ''),
			default_lead_time_days
			, COALESCE(location_label, ''), latitude, longitude, COALESCE(notes, '')
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
			&supplier.LocationLabel,
			&supplier.Latitude,
			&supplier.Longitude,
			&supplier.Notes,
		); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		suppliers = append(suppliers, supplier)
	}

	writeJSON(w, http.StatusOK, suppliers)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	var req UpdateRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Name = strings.Join(strings.Fields(req.Name), " ")
	req.Mark = strings.Join(strings.Fields(req.Mark), " ")
	req.Phone = strings.TrimSpace(req.Phone)
	req.LocationLabel = strings.Join(strings.Fields(req.LocationLabel), " ")
	req.Notes = strings.TrimSpace(req.Notes)
	if req.Name == "" || req.Mark == "" || req.LocationLabel == "" || req.Latitude == nil || req.Longitude == nil {
		writeError(w, "name, mark, locality, and map pin are required", http.StatusBadRequest)
		return
	}
	if len([]rune(req.Name)) > 120 || len([]rune(req.Mark)) > 80 || len([]rune(req.LocationLabel)) > 180 || len([]rune(req.Notes)) > 1000 || *req.Latitude < -90 || *req.Latitude > 90 || *req.Longitude < -180 || *req.Longitude > 180 {
		writeError(w, "invalid supplier details", http.StatusBadRequest)
		return
	}
	command, err := h.db.Exec(r.Context(), `UPDATE suppliers SET name=$1, mark=$2, phone=NULLIF($3,''), location_label=$4, latitude=$5, longitude=$6, notes=NULLIF($7,''), updated_at=NOW() WHERE id=$8 AND shop_id=$9 AND is_active=TRUE`, req.Name, req.Mark, req.Phone, req.LocationLabel, req.Latitude, req.Longitude, req.Notes, r.PathValue("id"), shopID)
	if err != nil || command.RowsAffected() != 1 {
		writeError(w, "could not update mark", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
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
