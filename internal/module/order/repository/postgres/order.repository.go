package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
	"github.com/chand-magar/gamenglam-server/internal/module/order/repository"
)

type orderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) repository.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetAll(ctx context.Context, req *models.OrderListRequest) (*models.OrderListResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.Limit

	// Build dynamic WHERE clause
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1

	if req.Status != "" {
		where += fmt.Sprintf(" AND o.order_status = $%d", n)
		args = append(args, req.Status)
		n++
	}
	if req.Search != "" {
		where += fmt.Sprintf(" AND o.order_ref ILIKE $%d", n)
		args = append(args, "%"+req.Search+"%")
		n++
	}

	// Count query
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := r.db.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM orders.orders o %s`, where),
		countArgs...,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.GetAll: count: %w", err)
	}

	// Data query — JOIN customer for name + subquery for item count
	args = append(args, req.Limit, offset)
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT
			o.order_no, o.order_id, o.customer_no, o.address_no,
			COALESCE(cu.first_name || ' ' || cu.last_name, '') AS customer_name,
			o.order_ref, o.subtotal, o.discount, o.tax,
			o.delivery_fee, o.total, o.order_status, o.notes,
			o.created_at, o.updated_at,
			(SELECT COUNT(*) FROM orders.order_items oi WHERE oi.order_no = o.order_no) AS item_count
		FROM orders.orders o
		LEFT JOIN customer.customers cu ON cu.customer_no = o.customer_no
		%s
		ORDER BY o.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, n, n+1), args...)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.GetAll: query: %w", err)
	}
	defer rows.Close()

	var data []models.Order
	for rows.Next() {
		var d models.Order
		if err := rows.Scan(
			&d.OrderNo, &d.OrderId, &d.CustomerNo, &d.AddressNo,
			&d.CustomerName,
			&d.OrderRef, &d.Subtotal, &d.Discount, &d.Tax,
			&d.DeliveryFee, &d.Total, &d.OrderStatus, &d.Notes,
			&d.CreatedAt, &d.UpdatedAt,
			&d.ItemCount,
		); err != nil {
			return nil, fmt.Errorf("orderRepository.GetAll: scan: %w", err)
		}
		data = append(data, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("orderRepository.GetAll: rows: %w", err)
	}

	totalPages := (total + req.Limit - 1) / req.Limit

	return &models.OrderListResponse{
		Orders:     data,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

func scanOrder(row pgx.Row) (*models.Order, error) {
	var o models.Order
	err := row.Scan(
		&o.OrderNo, &o.OrderId, &o.CustomerNo, &o.AddressNo,
		&o.OrderRef, &o.Subtotal, &o.Discount, &o.Tax,
		&o.DeliveryFee, &o.Total, &o.OrderStatus, &o.Notes,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) Save(ctx context.Context, order *models.Order) (*models.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// ── Insert order ──────────────────────────────────────────────────────────
	var o models.Order
	err = tx.QueryRow(ctx, `
		INSERT INTO orders.orders (
			order_id, customer_no, address_no, order_ref,
			subtotal, discount, tax, delivery_fee, total,
			order_status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING
			order_no, order_id, customer_no, address_no,
			order_ref, subtotal, discount, tax,
			delivery_fee, total, order_status, notes,
			created_at, updated_at
	`,
		order.OrderId, order.CustomerNo, order.AddressNo, order.OrderRef,
		order.Subtotal, order.Discount, order.Tax, order.DeliveryFee, order.Total,
		models.OrderStatusPending, order.Notes,
	).Scan(
		&o.OrderNo, &o.OrderId, &o.CustomerNo, &o.AddressNo,
		&o.OrderRef, &o.Subtotal, &o.Discount, &o.Tax,
		&o.DeliveryFee, &o.Total, &o.OrderStatus, &o.Notes,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.Save: insert order: %w", err)
	}

	// ── Insert order items + decrement stock atomically ───────────────────────
	for _, item := range order.Items {
		attributesJSON, err := json.Marshal(item.Attributes)
		if err != nil {
			return nil, fmt.Errorf("orderRepository.Save: marshal attributes: %w", err)
		}

		var itemNo uint32
		err = tx.QueryRow(ctx, `
			INSERT INTO orders.order_items (
				order_item_id, order_no, variant_no,
				product_name, sku, attributes,
				quantity, unit_price, total_price
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING order_item_no
		`,
			uuid.New(), o.OrderNo, item.VariantNo,
			item.ProductName, item.SKU, attributesJSON,
			item.Quantity, item.UnitPrice, item.TotalPrice,
		).Scan(&itemNo)
		if err != nil {
			return nil, fmt.Errorf("orderRepository.Save: insert order item: %w", err)
		}

		// Decrement stock atomically — same transaction
		result, err := tx.Exec(ctx, `
			UPDATE catalog.product_variants
			SET stock_qty = stock_qty - $1, updated_at = NOW()
			WHERE variant_no = $2 AND stock_qty >= $1
		`, item.Quantity, item.VariantNo)
		if err != nil {
			return nil, fmt.Errorf("orderRepository.Save: decrement stock: %w", err)
		}
		if result.RowsAffected() == 0 {
			return nil, fmt.Errorf("orderRepository.Save: insufficient stock for variant_no %d", item.VariantNo)
		}

		o.Items = append(o.Items, item)
	}

	// ── Insert initial status history ─────────────────────────────────────────
	_, err = tx.Exec(ctx, `
		INSERT INTO orders.order_status_history (history_id, order_no, from_status, to_status, note)
		VALUES ($1, $2, NULL, $3, $4)
	`, uuid.New(), o.OrderNo, models.OrderStatusPending, "Order placed")
	if err != nil {
		return nil, fmt.Errorf("orderRepository.Save: insert status history: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("orderRepository.Save: commit: %w", err)
	}

	return &o, nil
}

func (r *orderRepository) GetByCustomerId(ctx context.Context, customerId uuid.UUID) ([]models.Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			o.order_no, o.order_id, o.customer_no, o.address_no,
			o.order_ref, o.subtotal, o.discount, o.tax,
			o.delivery_fee, o.total, o.order_status, o.notes,
			o.created_at, o.updated_at
		FROM orders.orders o
		JOIN customer.customers cu ON cu.customer_no = o.customer_no
		WHERE cu.customer_id = $1
		ORDER BY o.created_at DESC
	`, customerId)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.GetByCustomerId: %w", err)
	}
	defer rows.Close()

	var data []models.Order
	for rows.Next() {
		var d models.Order
		if err := rows.Scan(
			&d.OrderNo, &d.OrderId, &d.CustomerNo, &d.AddressNo,
			&d.OrderRef, &d.Subtotal, &d.Discount, &d.Tax,
			&d.DeliveryFee, &d.Total, &d.OrderStatus, &d.Notes,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("orderRepository.GetByCustomerId: scan: %w", err)
		}
		data = append(data, d)
	}
	return data, rows.Err()
}

func (r *orderRepository) GetById(ctx context.Context, orderId uuid.UUID) (*models.Order, error) {
	var o models.Order
	err := r.db.QueryRow(ctx, `
		SELECT
			o.order_no, o.order_id, o.customer_no, o.address_no,
			COALESCE(cu.first_name || ' ' || cu.last_name, '') AS customer_name,
			o.order_ref, o.subtotal, o.discount, o.tax,
			o.delivery_fee, o.total, o.order_status, o.notes,
			o.created_at, o.updated_at,
			(SELECT COUNT(*) FROM orders.order_items oi WHERE oi.order_no = o.order_no) AS item_count
		FROM orders.orders o
		LEFT JOIN customer.customers cu ON cu.customer_no = o.customer_no
		WHERE o.order_id = $1
	`, orderId).Scan(
		&o.OrderNo, &o.OrderId, &o.CustomerNo, &o.AddressNo,
		&o.CustomerName,
		&o.OrderRef, &o.Subtotal, &o.Discount, &o.Tax,
		&o.DeliveryFee, &o.Total, &o.OrderStatus, &o.Notes,
		&o.CreatedAt, &o.UpdatedAt,
		&o.ItemCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("orderRepository.GetById: %w", err)
	}

	// Fetch items with primary variant image
	rows, err := r.db.Query(ctx, `
		SELECT
			oi.order_item_no, oi.order_item_id, oi.order_no, oi.variant_no,
			oi.product_name, oi.sku, oi.attributes,
			oi.quantity, oi.unit_price, oi.total_price, oi.created_at,
			(SELECT pvi.image_url
			 FROM catalog.product_variant_images pvi
			 WHERE pvi.variant_no = oi.variant_no
			 ORDER BY pvi.is_primary DESC, pvi.sort_order ASC
			 LIMIT 1) AS image_url
		FROM orders.order_items oi
		WHERE oi.order_no = $1
		ORDER BY oi.created_at ASC
	`, o.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.GetById: fetch items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		var attributesRaw []byte
		if err := rows.Scan(
			&item.OrderItemNo, &item.OrderItemId, &item.OrderNo, &item.VariantNo,
			&item.ProductName, &item.SKU, &attributesRaw,
			&item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.CreatedAt,
			&item.ImageUrl,
		); err != nil {
			return nil, fmt.Errorf("orderRepository.GetById: scan item: %w", err)
		}
		item.Attributes = attributesRaw
		o.Items = append(o.Items, item)
	}

	return &o, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, orderId uuid.UUID, req *models.UpdateOrderStatusRequest) (*models.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.UpdateStatus: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock row and capture current status + order_no for history
	var currentStatus models.OrderStatus
	var orderNo uint32
	err = tx.QueryRow(ctx, `
		SELECT order_no, order_status FROM orders.orders
		WHERE order_id = $1
		FOR UPDATE
	`, orderId).Scan(&orderNo, &currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("orderRepository.UpdateStatus: lock row: %w", err)
	}

	// Update + return enriched row via CTE (includes customer name + item count)
	var o models.Order
	err = tx.QueryRow(ctx, `
		WITH updated AS (
			UPDATE orders.orders
			SET order_status = $1, updated_at = NOW()
			WHERE order_id = $2
			RETURNING order_no, order_id, customer_no, address_no,
			          order_ref, subtotal, discount, tax,
			          delivery_fee, total, order_status, notes,
			          created_at, updated_at
		)
		SELECT
			u.order_no, u.order_id, u.customer_no, u.address_no,
			COALESCE(cu.first_name || ' ' || cu.last_name, '') AS customer_name,
			u.order_ref, u.subtotal, u.discount, u.tax,
			u.delivery_fee, u.total, u.order_status, u.notes,
			u.created_at, u.updated_at,
			(SELECT COUNT(*) FROM orders.order_items oi WHERE oi.order_no = u.order_no) AS item_count
		FROM updated u
		LEFT JOIN customer.customers cu ON cu.customer_no = u.customer_no
	`, req.ToStatus, orderId).Scan(
		&o.OrderNo, &o.OrderId, &o.CustomerNo, &o.AddressNo,
		&o.CustomerName,
		&o.OrderRef, &o.Subtotal, &o.Discount, &o.Tax,
		&o.DeliveryFee, &o.Total, &o.OrderStatus, &o.Notes,
		&o.CreatedAt, &o.UpdatedAt,
		&o.ItemCount,
	)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.UpdateStatus: update: %w", err)
	}

	// Append status history
	_, err = tx.Exec(ctx, `
		INSERT INTO orders.order_status_history (history_id, order_no, from_status, to_status, changed_by, note)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), orderNo, currentStatus, req.ToStatus, req.ChangedBy, req.Note)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.UpdateStatus: insert history: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("orderRepository.UpdateStatus: commit: %w", err)
	}

	return &o, nil
}

func (r *orderRepository) GetStatusHistory(ctx context.Context, orderNo uint32) ([]models.OrderStatusHistory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT history_no, history_id, order_no, from_status, to_status, changed_by, note, changed_at
		FROM orders.order_status_history
		WHERE order_no = $1
		ORDER BY changed_at ASC
	`, orderNo)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.GetStatusHistory: %w", err)
	}
	defer rows.Close()

	var history []models.OrderStatusHistory
	for rows.Next() {
		var h models.OrderStatusHistory
		if err := rows.Scan(
			&h.HistoryNo, &h.HistoryId, &h.OrderNo,
			&h.FromStatus, &h.ToStatus, &h.ChangedBy, &h.Note, &h.ChangedAt,
		); err != nil {
			return nil, fmt.Errorf("orderRepository.GetStatusHistory: scan: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}
