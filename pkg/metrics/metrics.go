package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Đo số lượng đơn hàng (Business)
	OrdersCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flashsale_orders_total",
			Help: "Total number of orders created",
		},
		[]string{"status"},
	)

	// Đo thời gian xử lý API (Performance)
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"method", "route", "status_code"},
	)

	// Đo tồn kho (Infrastructure/Business)
	InventoryStock = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "flashsale_inventory_stock_current",
			Help: "Current stock level of products",
		},
		[]string{"product_id"},
	)

	//đo độ trễ của reddis
	RedisLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "flashsale_redis_latency_seconds",
			Help:    "Latency of Redis operations",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"command"}, // label: "deduct", "get", "set"
	)

	// đo độ trễ xử lý (queue log)
	WorkerLag = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "flashsale_worker_lag_seconds",
			Help:    "Time difference between order creation and processing",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"status"}, // label: "processed", "failed"
	)

	// [NEW] Đếm số lần chặn đơn trùng (Idempotency)
	IdempotencyDuplicates = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "flashsale_idempotency_duplicate_total",
			Help: "Total number of duplicate orders detected and skipped",
		},
	)
)
