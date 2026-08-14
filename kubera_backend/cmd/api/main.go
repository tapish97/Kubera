package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"kubera_backend/internal/adjustment"
	"kubera_backend/internal/dashboard"
	"kubera_backend/internal/database"
	"kubera_backend/internal/fruit"
	"kubera_backend/internal/inventory"
	"kubera_backend/internal/sale"
	"kubera_backend/internal/supplier"
)

func main() {
	// Loads .env locally.
	// On hosted environments, env variables will be injected directly.
	_ = godotenv.Load()

	ctx := context.Background()

	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	fruitHandler := fruit.NewHandler(db)
	supplierHandler := supplier.NewHandler(db)
	inventoryHandler := inventory.NewHandler(db)
	saleHandler := sale.NewHandler(db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"service":  "kubera-api",
			"database": "connected",
		})
	})

	// Fruits
	mux.HandleFunc("POST /fruits", fruitHandler.Create)
	mux.HandleFunc("GET /fruits", fruitHandler.List)

	// Suppliers / Marks
	mux.HandleFunc("POST /suppliers", supplierHandler.Create)
	mux.HandleFunc("GET /suppliers", supplierHandler.List)

	// Inventory
	mux.HandleFunc(
		"POST /inventory/batches",
		inventoryHandler.CreateBatch,
	)

	mux.HandleFunc(
		"GET /inventory",
		inventoryHandler.List,
	)
	//sale
	mux.HandleFunc("POST /sales", saleHandler.Create)
	mux.HandleFunc("GET /sales", saleHandler.List)

	adjustmentHandler := adjustment.NewHandler(db)
	dashboardHandler := dashboard.NewHandler(db)

	mux.HandleFunc(
		"POST /inventory/adjustments",
		adjustmentHandler.Create,
	)

	mux.HandleFunc(
		"GET /inventory/adjustments",
		adjustmentHandler.List,
	)

	mux.HandleFunc(
		"GET /dashboard/summary",
		dashboardHandler.Summary,
	)

	mux.HandleFunc(
		"GET /dashboard/recent-sales",
		dashboardHandler.RecentSales,
	)

	mux.HandleFunc(
		"GET /dashboard/stock-by-fruit",
		dashboardHandler.StockByFruit,
	)

	mux.HandleFunc(
		"GET /dashboard/stock-by-supplier",
		dashboardHandler.StockBySupplier,
	)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           corsMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf(
		"Kubera API running on http://localhost:%s",
		port,
	)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {

		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			// Fine for development.
			// Later we'll replace * with the actual
			// Kubera frontend URL.
			w.Header().Set(
				"Access-Control-Allow-Origin",
				"*",
			)

			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Authorization",
			)

			w.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, PATCH, DELETE, OPTIONS",
			)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}
