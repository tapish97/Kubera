package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"kubera_backend/internal/requestlog"
)

type traceKey string

const (
	queryTraceKey traceKey = "database-query-trace"
	batchTraceKey traceKey = "database-batch-trace"
)

type traceState struct {
	started   time.Time
	operation string
	count     int
}

type queryTracer struct {
	slowThreshold time.Duration
}

func (t queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, queryTraceKey, traceState{started: time.Now(), operation: sqlOperation(data.SQL)})
}

func (t queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	state, ok := ctx.Value(queryTraceKey).(traceState)
	if !ok {
		return
	}
	duration := time.Since(state.started)
	isFailure := data.Err != nil && !errors.Is(data.Err, pgx.ErrNoRows)
	if duration < t.slowThreshold && !isFailure {
		return
	}
	fields := []any{"operation", state.operation, "duration_ms", duration.Milliseconds()}
	if requestID := requestlog.RequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, "request_id", requestID)
	}
	if isFailure {
		fields = append(fields, "error", data.Err)
		slog.Error("database query failed", fields...)
		return
	}
	slog.Warn("slow database query", fields...)
}

func (t queryTracer) TraceBatchStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	return context.WithValue(ctx, batchTraceKey, traceState{started: time.Now(), operation: "BATCH", count: data.Batch.Len()})
}

func (t queryTracer) TraceBatchQuery(context.Context, *pgx.Conn, pgx.TraceBatchQueryData) {}

func (t queryTracer) TraceBatchEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceBatchEndData) {
	state, ok := ctx.Value(batchTraceKey).(traceState)
	if !ok {
		return
	}
	duration := time.Since(state.started)
	if duration < t.slowThreshold && data.Err == nil {
		return
	}
	fields := []any{"operation", state.operation, "query_count", state.count, "duration_ms", duration.Milliseconds()}
	if requestID := requestlog.RequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, "request_id", requestID)
	}
	if data.Err != nil {
		fields = append(fields, "error", data.Err)
		slog.Error("database batch failed", fields...)
		return
	}
	slog.Warn("slow database batch", fields...)
}

func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	config.ConnConfig.Tracer = queryTracer{slowThreshold: slowQueryThreshold()}

	host := strings.ToLower(config.ConnConfig.Host)
	isNeon := strings.HasSuffix(host, ".neon.tech")
	isNeonPooler := isNeon && strings.Contains(host, "-pooler.")
	slog.Info("database configuration", "neon", isNeon, "neon_pooler", isNeonPooler, "client_pool_max_connections", config.MaxConns)
	if isNeon && !isNeonPooler {
		slog.Warn("Neon direct connection detected; use the pooled DATABASE_URL on Render")
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func slowQueryThreshold() time.Duration {
	const fallback = 250 * time.Millisecond
	value := strings.TrimSpace(os.Getenv("SLOW_QUERY_MS"))
	if value == "" {
		return fallback
	}
	milliseconds, err := strconv.Atoi(value)
	if err != nil || milliseconds <= 0 {
		return fallback
	}
	return time.Duration(milliseconds) * time.Millisecond
}

func sqlOperation(sql string) string {
	fields := strings.Fields(strings.TrimSpace(sql))
	if len(fields) == 0 {
		return "UNKNOWN"
	}
	return strings.ToUpper(fields[0])
}
