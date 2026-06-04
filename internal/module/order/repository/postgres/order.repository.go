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
			UPDATE product.product_variants
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

func (r *orderRepository) GetAll(ctx context.Context, customerNo uint32) ([]models.Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			order_no, order_id, customer_no, address_no,
			order_ref, subtotal, discount, tax,
			delivery_fee, total, order_status, notes,
			created_at, updated_at
		FROM orders.orders
		WHERE customer_no = $1
		ORDER BY created_at DESC
	`, customerNo)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.OrderNo, &o.OrderId, &o.CustomerNo, &o.AddressNo,
			&o.OrderRef, &o.Subtotal, &o.Discount, &o.Tax,
			&o.DeliveryFee, &o.Total, &o.OrderStatus, &o.Notes,
			&o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("orderRepository.GetAll: scan: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("orderRepository.GetAll: rows: %w", err)
	}

	return orders, nil
}

func (r *orderRepository) GetById(ctx context.Context, orderId uuid.UUID) (*models.Order, error) {
	o, err := scanOrder(r.db.QueryRow(ctx, `
		SELECT
			order_no, order_id, customer_no, address_no,
			order_ref, subtotal, discount, tax,
			delivery_fee, total, order_status, notes,
			created_at, updated_at
		FROM orders.orders
		WHERE order_id = $1
	`, orderId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("orderRepository.GetById: %w", err)
	}

	// Fetch items
	rows, err := r.db.Query(ctx, `
		SELECT
			order_item_no, order_item_id, order_no, variant_no,
			product_name, sku, attributes, quantity, unit_price, total_price, created_at
		FROM orders.order_items
		WHERE order_no = $1
		ORDER BY created_at ASC
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
		); err != nil {
			return nil, fmt.Errorf("orderRepository.GetById: scan item: %w", err)
		}
		item.Attributes = attributesRaw
		o.Items = append(o.Items, item)
	}

	return o, nil
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

	// Lock row
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

	// Update status
	var o models.Order
	err = tx.QueryRow(ctx, `
		UPDATE orders.orders
		SET order_status = $1, updated_at = NOW()
		WHERE order_id = $2
		RETURNING
			order_no, order_id, customer_no, address_no,
			order_ref, subtotal, discount, tax,
			delivery_fee, total, order_status, notes,
			created_at, updated_at
	`, req.ToStatus, orderId).Scan(
		&o.OrderNo, &o.OrderId, &o.CustomerNo, &o.AddressNo,
		&o.OrderRef, &o.Subtotal, &o.Discount, &o.Tax,
		&o.DeliveryFee, &o.Total, &o.OrderStatus, &o.Notes,
		&o.CreatedAt, &o.UpdatedAt,
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