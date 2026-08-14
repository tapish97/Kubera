package fruit

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

type Fruit struct {
	ID          string `json:"id"`
	ShopID      string `json:"shop_id"`
	Name        string `json:"name"`
	DefaultUnit string `json:"default_unit"`
}

type CreateRequest struct {
	ShopID      string `json:"shop_id"`
	Name        string `json:"name"`
	DefaultUnit string `json:"default_unit"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ShopID == "" || req.Name == "" {
		writeError(w, "shop_id and name are required", http.StatusBadRequest)
		return
	}

	if req.DefaultUnit == "" {
		req.DefaultUnit = "box"
	}

	var fruit Fruit

	err := h.db.QueryRow(
		r.Context(),
		`
		INSERT INTO fruits (
			shop_id,
			name,
			default_unit
		)
		VALUES ($1, $2, $3)
		RETURNING id, shop_id, name, default_unit
		`,
		req.ShopID,
		req.Name,
		req.DefaultUnit,
	).Scan(
		&fruit.ID,
		&fruit.ShopID,
		&fruit.Name,
		&fruit.DefaultUnit,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, fruit)
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
		SELECT id, shop_id, name, default_unit
		FROM fruits
		WHERE shop_id = $1
		  AND is_active = TRUE
		ORDER BY name
		`,
		shopID,
	)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	fruits := make([]Fruit, 0)

	for rows.Next() {
		var fruit Fruit

		if err := rows.Scan(
			&fruit.ID,
			&fruit.ShopID,
			&fruit.Name,
			&fruit.DefaultUnit,
		); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fruits = append(fruits, fruit)
	}

	writeJSON(w, http.StatusOK, fruits)
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
