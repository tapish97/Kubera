package fruit

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"kubera_backend/internal/auth"
)

type Handler struct{ db *pgxpool.Pool }

func NewHandler(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

type Fruit struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DefaultUnit string `json:"default_unit"`
	IsActive    bool   `json:"is_active"`
}
type CreateRequest struct {
	Name        string `json:"name"`
	DefaultUnit string `json:"default_unit"`
}
type UpdateRequest struct {
	Name        string `json:"name"`
	DefaultUnit string `json:"default_unit"`
}
type StatusRequest struct {
	IsActive *bool `json:"is_active"`
}

var allowedUnits = map[string]bool{"box": true, "kg": true, "piece": true, "crate": true, "dozen": true}

func normalize(name, unit string) (string, string, string) {
	name = strings.Join(strings.Fields(name), " ")
	unit = strings.ToLower(strings.TrimSpace(unit))
	if unit == "" {
		unit = "box"
	}
	if name == "" {
		return "", "", "name is required"
	}
	if len([]rune(name)) > 80 {
		return "", "", "name must be 80 characters or fewer"
	}
	if !allowedUnits[unit] {
		return "", "", "default_unit must be box, kg, piece, crate, or dozen"
	}
	return name, unit, ""
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	shop, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	name, unit, message := normalize(req.Name, req.DefaultUnit)
	if message != "" {
		writeError(w, message, http.StatusBadRequest)
		return
	}
	var fruit Fruit
	err := h.db.QueryRow(r.Context(), `INSERT INTO fruits (shop_id, name, default_unit)
		SELECT $1, $2, $3 WHERE NOT EXISTS (SELECT 1 FROM fruits WHERE shop_id = $1 AND lower(name) = lower($2))
		RETURNING id, name, default_unit, is_active`, shop, name, unit).Scan(&fruit.ID, &fruit.Name, &fruit.DefaultUnit, &fruit.IsActive)
	if errors.Is(err, pgx.ErrNoRows) || isUniqueViolation(err) {
		writeError(w, "a fruit with this name already exists", http.StatusConflict)
		return
	}
	if err != nil {
		log.Printf("create fruit for shop %s: %v", shop, err)
		writeError(w, "could not create fruit", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, fruit)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	shop, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	includeInactive := r.URL.Query().Get("include_inactive") == "true"
	rows, err := h.db.Query(r.Context(), `SELECT id, name, default_unit, is_active FROM fruits
		WHERE shop_id = $1 AND ($2 OR is_active = TRUE) AND ($3 = '' OR name ILIKE '%' || $3 || '%')
		ORDER BY is_active DESC, lower(name)`, shop, includeInactive, search)
	if err != nil {
		writeError(w, "could not load fruits", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	fruits := make([]Fruit, 0)
	for rows.Next() {
		var fruit Fruit
		if err := rows.Scan(&fruit.ID, &fruit.Name, &fruit.DefaultUnit, &fruit.IsActive); err != nil {
			writeError(w, "could not load fruits", http.StatusInternalServerError)
			return
		}
		fruits = append(fruits, fruit)
	}
	if err := rows.Err(); err != nil {
		writeError(w, "could not load fruits", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, fruits)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	shop, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	var fruit Fruit
	err := h.db.QueryRow(r.Context(), `SELECT id, name, default_unit, is_active FROM fruits WHERE id = $1 AND shop_id = $2`, r.PathValue("id"), shop).Scan(&fruit.ID, &fruit.Name, &fruit.DefaultUnit, &fruit.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, "fruit not found", http.StatusNotFound)
		return
	}
	if err != nil {
		writeError(w, "could not load fruit", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, fruit)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	shop, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	name, unit, message := normalize(req.Name, req.DefaultUnit)
	if message != "" {
		writeError(w, message, http.StatusBadRequest)
		return
	}
	var fruit Fruit
	err := h.db.QueryRow(r.Context(), `UPDATE fruits SET name = $1, default_unit = $2 WHERE id = $3 AND shop_id = $4
		AND NOT EXISTS (SELECT 1 FROM fruits other WHERE other.shop_id = $4 AND lower(other.name) = lower($1) AND other.id <> $3)
		RETURNING id, name, default_unit, is_active`, name, unit, r.PathValue("id"), shop).Scan(&fruit.ID, &fruit.Name, &fruit.DefaultUnit, &fruit.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		if checkErr := h.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM fruits WHERE id = $1 AND shop_id = $2)`, r.PathValue("id"), shop).Scan(&exists); checkErr != nil {
			writeError(w, "could not update fruit", http.StatusInternalServerError)
			return
		}
		if exists {
			writeError(w, "a fruit with this name already exists", http.StatusConflict)
		} else {
			writeError(w, "fruit not found", http.StatusNotFound)
		}
		return
	}
	if isUniqueViolation(err) {
		writeError(w, "a fruit with this name already exists", http.StatusConflict)
		return
	}
	if err != nil {
		writeError(w, "could not update fruit", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, fruit)
}

func (h *Handler) SetStatus(w http.ResponseWriter, r *http.Request) {
	var req StatusRequest
	if err := decodeJSON(w, r, &req); err != nil || req.IsActive == nil {
		writeError(w, "is_active is required", http.StatusBadRequest)
		return
	}
	shop, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	var fruit Fruit
	err := h.db.QueryRow(r.Context(), `UPDATE fruits SET is_active = $1 WHERE id = $2 AND shop_id = $3 RETURNING id, name, default_unit, is_active`, *req.IsActive, r.PathValue("id"), shop).Scan(&fruit.ID, &fruit.Name, &fruit.DefaultUnit, &fruit.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, "fruit not found", http.StatusNotFound)
		return
	}
	if err != nil {
		writeError(w, "could not update fruit status", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, fruit)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
func writeError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, status, map[string]string{"error": message})
}
