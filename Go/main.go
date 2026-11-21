package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Item struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderRequest struct {
	CustomerID string `json:"customerId"`
	Items      []Item `json:"items"`
}

type OrderResponse struct {
	OrderID     string    `json:"orderId"`
	CustomerID  string    `json:"customerId"`
	TotalAmount float64   `json:"totalAmount"`
	ItemsCount  int       `json:"itemsCount"`
	ProcessedAt time.Time `json:"processedAt"`
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&pool_max_conns=%s",
		env("DB_USER", "postgres"), env("DB_PASSWORD", "postgres"),
		env("DB_HOST", "localhost"), env("DB_PORT", "5433"),
		env("DB_NAME", "orders_db"), env("DB_POOL_SIZE", "50"))

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/orders", func(w http.ResponseWriter, r *http.Request) {
		createOrder(pool, w, r)
	})

	addr := ":" + env("PORT", "5001")
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func createOrder(pool *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if msg := validate(req); msg != "" {
		httpError(w, http.StatusBadRequest, msg)
		return
	}

	total, count := totals(req.Items)
	orderID := rand.Text() // 26 chars, random, fits VARCHAR(50)

	ctx := r.Context()
	tx, err := pool.Begin(ctx)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(ctx) // no-op after Commit

	// One round trip for the order + all its items.
	batch := &pgx.Batch{}
	var createdAt time.Time
	batch.Queue(`INSERT INTO orders (order_id, customer_id, total_amount, items_count)
	             VALUES ($1, $2, $3, $4) RETURNING created_at`,
		orderID, req.CustomerID, total, count).QueryRow(func(row pgx.Row) error {
		return row.Scan(&createdAt)
	})
	for _, it := range req.Items {
		batch.Queue(`INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)`,
			orderID, it.ProductID, it.Quantity, it.Price)
	}
	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := tx.Commit(ctx); err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(OrderResponse{orderID, req.CustomerID, total, count, createdAt})
}

func validate(req OrderRequest) string {
	if req.CustomerID == "" || len(req.CustomerID) > 50 {
		return "customerId is required (max 50 chars)"
	}
	if len(req.Items) == 0 {
		return "items must not be empty"
	}
	for _, it := range req.Items {
		if it.ProductID == "" || len(it.ProductID) > 50 || it.Quantity <= 0 || it.Price < 0 {
			return "each item needs productId, quantity > 0 and price >= 0"
		}
	}
	return ""
}

// totals returns the order total (rounded to cents) and the number of units.
func totals(items []Item) (float64, int) {
	var total float64
	var count int
	for _, it := range items {
		total += float64(it.Quantity) * it.Price
		count += it.Quantity
	}
	return math.Round(total*100) / 100, count
}

func httpError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
