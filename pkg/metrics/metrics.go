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
)
