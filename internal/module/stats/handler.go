package stats

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type RecentOrder struct {
	OrderId      string    `json:"order_id"`
	OrderRef     string    `json:"order_ref"`
	CustomerName string    `json:"customer_name"`
	Total        float64   `json:"total"`
	OrderStatus  string    `json:"order_status"`
	CreatedAt    time.Time `json:"created_at"`
}

type DashboardStats struct {
	TotalOrders     int64              `json:"total_orders"`
	TotalRevenue    float64            `json:"total_revenue"`
	TotalCustomers  int64              `json:"total_customers"`
	LowStockCount   int64              `json:"low_stock_count"`
	OrdersByStatus  []OrderStatusCount `json:"orders_by_status"`
	RecentOrders    []RecentOrder      `json:"recent_orders"`
}

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()
	stats, err := h.fetchStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (h *Handler) fetchStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{
		OrdersByStatus: []OrderStatusCount{},
		RecentOrders:   []RecentOrder{},
	}

	// total orders
	if err := h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders.orders WHERE status != 'D'`,
	).Scan(&stats.TotalOrders); err != nil {
		return nil, fmt.Errorf("stats: total orders: %w", err)
	}

	// total revenue (all non-deleted orders)
	if err := h.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(total), 0) FROM orders.orders WHERE status != 'D'`,
	).Scan(&stats.TotalRevenue); err != nil {
		return nil, fmt.Errorf("stats: total revenue: %w", err)
	}

	// total customers
	if err := h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM customer.customers WHERE status != 'D'`,
	).Scan(&stats.TotalCustomers); err != nil {
		return nil, fmt.Errorf("stats: total customers: %w", err)
	}

	// low stock variants (active variants with stock_qty < 10)
	if err := h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM catalog.product_variants WHERE status = 'active' AND stock_qty < 10`,
	).Scan(&stats.LowStockCount); err != nil {
		return nil, fmt.Errorf("stats: low stock: %w", err)
	}

	// orders by status
	rows, err := h.db.Query(ctx,
		`SELECT order_status, COUNT(*) FROM orders.orders WHERE status != 'D' GROUP BY order_status ORDER BY order_status`,
	)
	if err != nil {
		return nil, fmt.Errorf("stats: orders by status: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sc OrderStatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, fmt.Errorf("stats: orders by status scan: %w", err)
		}
		stats.OrdersByStatus = append(stats.OrdersByStatus, sc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("stats: orders by status rows: %w", err)
	}

	// recent 5 orders with customer name
	recentRows, err := h.db.Query(ctx, `
		SELECT
			o.order_id, o.order_ref,
			c.first_name || ' ' || c.last_name AS customer_name,
			o.total, o.order_status, o.created_at
		FROM orders.orders o
		JOIN customer.customers c ON c.customer_no = o.customer_no
		WHERE o.status != 'D'
		ORDER BY o.created_at DESC
		LIMIT 5
	`)
	if err != nil {
		return nil, fmt.Errorf("stats: recent orders: %w", err)
	}
	defer recentRows.Close()
	for recentRows.Next() {
		var ro RecentOrder
		if err := recentRows.Scan(
			&ro.OrderId, &ro.OrderRef, &ro.CustomerName,
			&ro.Total, &ro.OrderStatus, &ro.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("stats: recent orders scan: %w", err)
		}
		stats.RecentOrders = append(stats.RecentOrders, ro)
	}
	if err := recentRows.Err(); err != nil {
		return nil, fmt.Errorf("stats: recent orders rows: %w", err)
	}

	return stats, nil
}
