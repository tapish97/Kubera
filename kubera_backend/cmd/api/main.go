package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"kubera_backend/internal/adjustment"
	kuberaauth "kubera_backend/internal/auth"
	"kubera_backend/internal/dashboard"
	"kubera_backend/internal/database"
	"kubera_backend/internal/fruit"
	"kubera_backend/internal/inventory"
	"kubera_backend/internal/requestlog"
	"kubera_backend/internal/sale"
	"kubera_backend/internal/supplier"
)

func main() {
	// Loads .env locally.
	// On hosted environments, env variables will be injected directly.
	_ = godotenv.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx := context.Background()

	db, err := database.Connect(ctx)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	logger.Info("database connected")

	defer db.Close()

	fruitHandler := fruit.NewHandler(db)
	supplierHandler := supplier.NewHandler(db)
	inventoryHandler := inventory.NewHandler(db)
	saleHandler := sale.NewHandler(db)
	authMiddleware, err := kuberaauth.NewMiddleware(db)
	if err != nil {
		logger.Error("authentication configuration failed", "error", err)
		os.Exit(1)
	}
	authHandler := kuberaauth.NewHandler(db)

	mux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"service":  "kubera-api",
			"database": "connected",
		})
	})
	mux.Handle("GET /health", requestlog.Middleware(logger, nil, healthHandler))

	// Fruits
	protectedMux.HandleFunc("GET /me", authHandler.Me)
	protectedMux.HandleFunc("POST /me/onboarding", authHandler.CompleteOnboarding)
	protectedMux.HandleFunc("PATCH /me/profile", authHandler.UpdateProfile)
	protectedMux.HandleFunc("PATCH /me/shop", authHandler.UpdateShop)

	protectedMux.HandleFunc("POST /fruits", fruitHandler.Create)
	protectedMux.HandleFunc("GET /fruits", fruitHandler.List)
	protectedMux.HandleFunc("GET /fruits/{id}", fruitHandler.Get)
	protectedMux.HandleFunc("PATCH /fruits/{id}", fruitHandler.Update)
	protectedMux.HandleFunc("PATCH /fruits/{id}/status", fruitHandler.SetStatus)

	// Suppliers / Marks
	protectedMux.HandleFunc("POST /suppliers", supplierHandler.Create)
	protectedMux.HandleFunc("GET /suppliers", supplierHandler.List)
	protectedMux.HandleFunc("PATCH /suppliers/{id}", supplierHandler.Update)

	// Inventory
	protectedMux.HandleFunc(
		"POST /inventory/batches",
		inventoryHandler.CreateBatch,
	)
	protectedMux.HandleFunc("POST /inventory/batches/quick", inventoryHandler.CreateQuickBatch)
	protectedMux.HandleFunc("GET /purchase-options", inventoryHandler.PurchaseOptions)
	protectedMux.HandleFunc("GET /inventory/unpriced", inventoryHandler.ListUnpriced)
	protectedMux.HandleFunc("PATCH /inventory/batches/{id}/purchase-price", inventoryHandler.SetPurchasePrice)
	protectedMux.HandleFunc("PATCH /inventory/batches/{id}", inventoryHandler.UpdateBatch)

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
	protectedMux.HandleFunc("GET /reports/daily", dashboardHandler.DailyReport)
	protectedMux.HandleFunc("POST /closings/{date}", dashboardHandler.CloseDay)
	protectedMux.HandleFunc("GET /closings", dashboardHandler.ClosingHistory)
	protectedMux.HandleFunc("GET /notifications/summary", dashboardHandler.NotificationSummary)

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

	logger.Info("server starting", "port", port, "service", "kubera-api")
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case received := <-shutdownSignals:
		logger.Info("shutdown signal received", "signal", received.String())
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			return
		}
		logger.Info("server stopped")
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
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
