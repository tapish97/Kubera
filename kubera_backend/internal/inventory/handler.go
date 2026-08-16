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
	LocationLabel        string   `json:"location_label"`
	Latitude             *float64 `json:"latitude"`
	Longitude            *float64 `json:"longitude"`
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

type SupplierOption struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Mark          string   `json:"mark"`
	LocationLabel string   `json:"location_label"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}
type FruitOption struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	DefaultUnit string           `json:"default_unit"`
	Suppliers   []SupplierOption `json:"suppliers"`
}

type SetPurchasePriceRequest struct {
	PurchasePricePerUnit *float64 `json:"purchase_price_per_unit"`
}

type UpdateBatchRequest struct {
	Quality              string   `json:"quality"`
	Size                 string   `json:"size"`
	QuantityRemaining    float64  `json:"quantity_remaining"`
	PurchasePricePerUnit *float64 `json:"purchase_price_per_unit"`
}

func (h *Handler) PurchaseOptions(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	rows, err := h.db.Query(r.Context(), `SELECT f.id, f.name, f.default_unit, s.id, s.name, s.mark, s.location_label, s.latitude, s.longitude
		FROM fruits f LEFT JOIN inventory_batches ib ON ib.fruit_id = f.id AND ib.shop_id = f.shop_id
		LEFT JOIN suppliers s ON s.id = ib.supplier_id AND s.shop_id = f.shop_id AND s.is_active = TRUE
		WHERE f.shop_id = $1 AND f.is_active = TRUE ORDER BY lower(f.name), lower(s.mark)`, shopID)
	if err != nil {
		writeError(w, "could not load purchase choices", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	options := make([]FruitOption, 0)
	positions := make(map[string]int)
	seen := make(map[string]bool)
	for rows.Next() {
		var fruitID, fruitName, unit string
		var supplierID, supplierName, mark, locationLabel *string
		var latitude, longitude *float64
		if err := rows.Scan(&fruitID, &fruitName, &unit, &supplierID, &supplierName, &mark, &locationLabel, &latitude, &longitude); err != nil {
			writeError(w, "could not load purchase choices", http.StatusInternalServerError)
			return
		}
		position, exists := positions[fruitID]
		if !exists {
			position = len(options)
			positions[fruitID] = position
			options = append(options, FruitOption{ID: fruitID, Name: fruitName, DefaultUnit: unit, Suppliers: make([]SupplierOption, 0)})
		}
		if supplierID != nil && !seen[fruitID+":"+*supplierID] {
			options[position].Suppliers = append(options[position].Suppliers, SupplierOption{ID: *supplierID, Name: *supplierName, Mark: *mark, LocationLabel: stringValue(locationLabel), Latitude: latitude, Longitude: longitude})
			seen[fruitID+":"+*supplierID] = true
		}
	}
	if err := rows.Err(); err != nil {
		writeError(w, "could not load purchase choices", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, options)
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
	req.LocationLabel = strings.Join(strings.Fields(req.LocationLabel), " ")
	req.Unit = strings.ToLower(strings.TrimSpace(req.Unit))
	if req.Unit == "" {
		req.Unit = "box"
	}
	if req.Size == "" {
		req.Size = "normal"
	}
	validUnit := map[string]bool{"box": true, "kg": true, "piece": true, "crate": true, "dozen": true}
	if req.Quantity <= 0 || (req.PurchasePricePerUnit != nil && *req.PurchasePricePerUnit < 0) || !validUnit[req.Unit] || (req.FruitID == "" && req.FruitName == "") || (req.SupplierID == "" && req.Mark == "") {
		writeError(w, "fruit, mark, quantity, unit, and buy price are required", http.StatusBadRequest)
		return
	}
	if len([]rune(req.FruitName)) > 80 || len([]rune(req.Mark)) > 80 || len([]rune(req.SupplierName)) > 120 {
		writeError(w, "fruit or mark name is too long", http.StatusBadRequest)
		return
	}
	if len([]rune(req.LocationLabel)) > 180 || (req.Latitude == nil) != (req.Longitude == nil) || (req.Latitude != nil && (*req.Latitude < -90 || *req.Latitude > 90 || *req.Longitude < -180 || *req.Longitude > 180)) {
		writeError(w, "invalid mark location", http.StatusBadRequest)
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
			if req.LocationLabel == "" || req.Latitude == nil || req.Longitude == nil {
				writeError(w, "locality and map pin are required for a new mark", http.StatusBadRequest)
				return
			}
			if req.SupplierName == "" {
				req.SupplierName = req.Mark
			}
			err = tx.QueryRow(r.Context(), `INSERT INTO suppliers (shop_id, name, mark, phone, location_label, latitude, longitude) VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7) RETURNING id`, shopID, req.SupplierName, req.Mark, strings.TrimSpace(req.Phone), req.LocationLabel, req.Latitude, req.Longitude).Scan(&supplierID)
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

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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

func (h *Handler) ListUnpriced(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	rows, err := h.db.Query(r.Context(), `SELECT ib.id, f.name, s.mark, ib.quality, ib.size, ib.unit, ib.quantity_received, ib.quantity_remaining, ib.received_at,
		COALESCE(SUM(si.quantity),0), COALESCE(SUM(si.total_sale_amount),0), CASE WHEN COALESCE(SUM(si.quantity),0)>0 THEN SUM(si.total_sale_amount)/SUM(si.quantity) END,
		CASE WHEN COALESCE(SUM(si.quantity),0)>0 THEN (SUM(si.total_sale_amount)/SUM(si.quantity))*0.94 ELSE NULL END
		FROM inventory_batches ib JOIN fruits f ON f.id=ib.fruit_id JOIN suppliers s ON s.id=ib.supplier_id LEFT JOIN sale_items si ON si.batch_id=ib.id
		WHERE ib.shop_id=$1 AND ib.purchase_price_per_unit IS NULL
		GROUP BY ib.id,f.name,s.mark,ib.quality,ib.size,ib.unit,ib.quantity_received,ib.quantity_remaining,ib.received_at ORDER BY ib.received_at DESC`, shopID)
	if err != nil {
		writeError(w, "could not load unsettled prices", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, fruit, mark, size, unit string
		var quality *string
		var receivedAt time.Time
		var received, remaining, sold, revenue float64
		var averageSelling, suggested *float64
		if err := rows.Scan(&id, &fruit, &mark, &quality, &size, &unit, &received, &remaining, &receivedAt, &sold, &revenue, &averageSelling, &suggested); err != nil {
			writeError(w, "could not load unsettled prices", http.StatusInternalServerError)
			return
		}
		items = append(items, map[string]any{"batch_id": id, "fruit": fruit, "mark": mark, "quality": quality, "size": size, "unit": unit, "quantity_received": received, "quantity_remaining": remaining, "quantity_sold": sold, "sales_revenue": revenue, "average_selling_price": averageSelling, "received_at": receivedAt, "suggested_purchase_price_per_unit": suggested})
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) SetPurchasePrice(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	var req SetPurchasePriceRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil || req.PurchasePricePerUnit == nil || *req.PurchasePricePerUnit < 0 {
		writeError(w, "valid buying price is required", http.StatusBadRequest)
		return
	}
	tx, err := h.db.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		writeError(w, "could not start price settlement", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())
	command, err := tx.Exec(r.Context(), `UPDATE inventory_batches SET purchase_price_per_unit=$1 WHERE id=$2 AND shop_id=$3`, *req.PurchasePricePerUnit, r.PathValue("id"), shopID)
	if err != nil || command.RowsAffected() != 1 {
		writeError(w, "batch not found", http.StatusNotFound)
		return
	}
	_, err = tx.Exec(r.Context(), `UPDATE sale_items si SET cost_price_per_unit=$1, cost_price_is_estimated=FALSE FROM sales s WHERE si.sale_id=s.id AND si.batch_id=$2 AND s.shop_id=$3`, *req.PurchasePricePerUnit, r.PathValue("id"), shopID)
	if err != nil {
		writeError(w, "could not update past sale profits", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, "could not save buying price", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "settled"})
}

func (h *Handler) UpdateBatch(w http.ResponseWriter, r *http.Request) {
	shopID, ok := auth.ShopIDFromContext(r.Context())
	if !ok {
		writeError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	var req UpdateBatchRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Quality = strings.TrimSpace(req.Quality)
	req.Size = strings.ToLower(strings.TrimSpace(req.Size))
	if req.Size == "" {
		req.Size = "normal"
	}
	if req.QuantityRemaining < 0 || (req.PurchasePricePerUnit != nil && *req.PurchasePricePerUnit < 0) || len([]rune(req.Quality)) > 80 || !map[string]bool{"small": true, "normal": true, "large": true}[req.Size] {
		writeError(w, "invalid stock details", http.StatusBadRequest)
		return
	}
	tx, err := h.db.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		writeError(w, "could not start stock update", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())
	var current, received float64
	err = tx.QueryRow(r.Context(), `SELECT quantity_remaining, quantity_received FROM inventory_batches WHERE id=$1 AND shop_id=$2 FOR UPDATE`, r.PathValue("id"), shopID).Scan(&current, &received)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, "batch not found", http.StatusNotFound)
		return
	}
	if err != nil {
		writeError(w, "could not load batch", http.StatusInternalServerError)
		return
	}
	if req.QuantityRemaining > received {
		writeError(w, "remaining quantity cannot exceed quantity received", http.StatusBadRequest)
		return
	}
	difference := req.QuantityRemaining - current
	if difference != 0 {
		adjustmentType := "correction_increase"
		quantity := difference
		if difference < 0 {
			adjustmentType = "correction_decrease"
			quantity = -difference
		}
		_, err = tx.Exec(r.Context(), `INSERT INTO inventory_adjustments (shop_id,batch_id,adjustment_type,quantity,reason) VALUES ($1,$2,$3,$4,'Manual inventory edit')`, shopID, r.PathValue("id"), adjustmentType, quantity)
		if err != nil {
			writeError(w, "could not record stock correction", http.StatusInternalServerError)
			return
		}
	}
	_, err = tx.Exec(r.Context(), `UPDATE inventory_batches SET quality=NULLIF($1,''), size=$2, quantity_remaining=$3, purchase_price_per_unit=$4 WHERE id=$5 AND shop_id=$6`, req.Quality, req.Size, req.QuantityRemaining, req.PurchasePricePerUnit, r.PathValue("id"), shopID)
	if err != nil {
		writeError(w, "could not update stock", http.StatusInternalServerError)
		return
	}
	if req.PurchasePricePerUnit != nil {
		_, err = tx.Exec(r.Context(), `UPDATE sale_items si SET cost_price_per_unit=$1,cost_price_is_estimated=FALSE FROM sales s WHERE si.sale_id=s.id AND si.batch_id=$2 AND s.shop_id=$3`, *req.PurchasePricePerUnit, r.PathValue("id"), shopID)
		if err != nil {
			writeError(w, "could not update sale profits", http.StatusInternalServerError)
			return
		}
	} else {
		_, err = tx.Exec(r.Context(), `UPDATE sale_items si SET cost_price_per_unit=si.selling_price_per_unit*0.94,cost_price_is_estimated=TRUE FROM sales s WHERE si.sale_id=s.id AND si.batch_id=$1 AND s.shop_id=$2`, r.PathValue("id"), shopID)
		if err != nil {
			writeError(w, "could not update provisional sale profits", http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, "could not save stock", http.StatusInternalServerError)
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
