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
	kuberaauth "kubera_backend/internal/auth"
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
	authMiddleware, err := kuberaauth.NewMiddleware(db)
	if err != nil {
		log.Fatal(err)
	}
	authHandler := kuberaauth.NewHandler(db)

	mux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"service":  "kubera-api",
			"database": "connected",
		})
	})

	// Fruits
	protectedMux.HandleFunc("GET /me", authHandler.Me)
	protectedMux.HandleFunc("POST /me/onboarding", authHandler.CompleteOnboarding)
	protectedMux.HandleFunc("PATCH /me/profile", authHandler.UpdateProfile)
	protectedMux.HandleFunc("PATCH /me/shop", authHandler.UpdateShop)

	protectedMux.HandleFunc("POST /fruits", fruitHandler.Create)
	protectedMux.HandleFunc("GET /fruits", fruitHandler.List)

	// Suppliers / Marks
	protectedMux.HandleFunc("POST /suppliers", supplierHandler.Create)
	protectedMux.HandleFunc("GET /suppliers", supplierHandler.List)

	// Inventory
	protectedMux.HandleFunc(
		"POST /inventory/batches",
		inventoryHandler.CreateBatch,
	)

	protectedMux.HandleFunc(
		"GET /inventory",
		inventoryHandler.List,
	)
	//sale
	protectedMux.HandleFunc("POST /sales", saleHandler.Create)
	protectedMux.HandleFunc("GET /sales", saleHandler.List)

	adjustmentHandler := adjustment.NewHandler(db)
	dashboardHandler := dashboard.NewHandler(db)

	protectedMux.HandleFunc(
		"POST /inventory/adjustments",
		adjustmentHandler.Create,
	)

	protectedMux.HandleFunc(
		"GET /inventory/adjustments",
		adjustmentHandler.List,
	)

	protectedMux.HandleFunc(
		"GET /dashboard/summary",
		dashboardHandler.Summary,
	)

	protectedMux.HandleFunc(
		"GET /dashboard/recent-sales",
		dashboardHandler.RecentSales,
	)

	protectedMux.HandleFunc(
		"GET /dashboard/stock-by-fruit",
		dashboardHandler.StockByFruit,
	)

	protectedMux.HandleFunc(
		"GET /dashboard/stock-by-supplier",
		dashboardHandler.StockBySupplier,
	)

	mux.Handle("/", authMiddleware.Protect(protectedMux))

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
