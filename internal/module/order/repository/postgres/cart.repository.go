package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
	"github.com/chand-magar/gamenglam-server/internal/module/order/repository"
)

type cartRepository struct {
	db *pgxpool.Pool
}

func NewCartRepository(db *pgxpool.Pool) repository.CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetByCustomerId(ctx context.Context, customerId uuid.UUID) (*models.Cart, error) {
	var c models.Cart
	err := r.db.QueryRow(ctx, `
		SELECT c.cart_no, c.cart_id, c.customer_no, c.created_at, c.updated_at
		FROM orders.carts c
		JOIN customer.customers cu ON cu.customer_no = c.customer_no
		WHERE cu.customer_id = $1
	`, customerId).Scan(
		&c.CartNo, &c.CartId, &c.CustomerNo, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("cartRepository.GetByCustomerId: %w", err)
	}
	return &c, nil
}

func (r *cartRepository) GetCartItem(ctx context.Context, cartNo uint32, variantNo uint32) (*models.CartItem, error) {
	var ci models.CartItem
	err := r.db.QueryRow(ctx, `
		SELECT cart_item_no, cart_item_id, cart_no, variant_no, quantity, unit_price, created_at, updated_at
		FROM orders.cart_items
		WHERE cart_no = $1 AND variant_no = $2
	`, cartNo, variantNo).Scan(
		&ci.CartItemNo, &ci.CartItemId, &ci.CartNo, &ci.VariantNo,
		&ci.Quantity, &ci.UnitPrice, &ci.CreatedAt, &ci.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("cartRepository.GetCartItem: %w", err)
	}
	return &ci, nil
}

func (r *cartRepository) GetCartWithItems(ctx context.Context, customerId uuid.UUID) (*models.Cart, error) {
	cart, err := r.GetByCustomerId(ctx, customerId)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, nil
	}

	rows, err := r.db.Query(ctx, `
		SELECT cart_item_no, cart_item_id, cart_no, variant_no, quantity, unit_price, created_at, updated_at
		FROM orders.cart_items
		WHERE cart_no = $1
		ORDER BY created_at ASC
	`, cart.CartNo)
	if err != nil {
		return nil, fmt.Errorf("cartRepository.GetCartWithItems: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ci models.CartItem
		if err := rows.Scan(
			&ci.CartItemNo, &ci.CartItemId, &ci.CartNo, &ci.VariantNo,
			&ci.Quantity, &ci.UnitPrice, &ci.CreatedAt, &ci.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("cartRepository.GetCartWithItems: scan: %w", err)
		}
		cart.Items = append(cart.Items, ci)
	}

	return cart, rows.Err()
}

func (r *cartRepository) SaveCart(ctx context.Context, customerId uuid.UUID) (*models.Cart, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("cartRepository.SaveCart: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var customerNo uint32
	err = tx.QueryRow(ctx, `
		SELECT customer_no FROM customer.customers
		WHERE customer_id = $1 AND status != 'D'
	`, customerId).Scan(&customerNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("cartRepository.SaveCart: customer not found")
		}
		return nil, fmt.Errorf("cartRepository.SaveCart: resolve customer: %w", err)
	}

	var c models.Cart
	err = tx.QueryRow(ctx, `
		INSERT INTO orders.carts (cart_id, customer_no)
		VALUES ($1, $2)
		RETURNING cart_no, cart_id, customer_no, created_at, updated_at
	`, uuid.New(), customerNo).Scan(
		&c.CartNo, &c.CartId, &c.CustomerNo, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("cartRepository.SaveCart: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("cartRepository.SaveCart: commit: %w", err)
	}
	return &c, nil
}

// AddItem — variantNo is resolved by the service from variant_id
func (r *cartRepository) AddItem(ctx context.Context, cartNo uint32, req *models.AddCartItemRequest, unitPrice float64) (*models.CartItem, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("cartRepository.AddItem: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Resolve variant_no from variant_id
	var variantNo uint32
	err = tx.QueryRow(ctx, `
		SELECT variant_no FROM catalog.product_variants
		WHERE variant_id = $1 AND status != 'D'
	`, req.VariantId).Scan(&variantNo)
	if err != nil {
		return nil, fmt.Errorf("cartRepository.AddItem: resolve variant: %w", err)
	}

	var ci models.CartItem
	err = tx.QueryRow(ctx, `
		INSERT INTO orders.cart_items (cart_item_id, cart_no, variant_no, quantity, unit_price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING cart_item_no, cart_item_id, cart_no, variant_no, quantity, unit_price, created_at, updated_at
	`, uuid.New(), cartNo, variantNo, req.Quantity, unitPrice).Scan(
		&ci.CartItemNo, &ci.CartItemId, &ci.CartNo, &ci.VariantNo,
		&ci.Quantity, &ci.UnitPrice, &ci.CreatedAt, &ci.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("cartRepository.AddItem: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("cartRepository.AddItem: commit: %w", err)
	}
	return &ci, nil
}

func (r *cartRepository) UpdateItem(ctx context.Context, cartItemId uuid.UUID, req *models.UpdateCartItemRequest) (*models.CartItem, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("cartRepository.UpdateItem: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT cart_item_no FROM orders.cart_items
		WHERE cart_item_id = $1
		FOR UPDATE
	`, cartItemId).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("cartRepository.UpdateItem: lock row: %w", err)
	}

	var ci models.CartItem
	err = tx.QueryRow(ctx, `
		UPDATE orders.cart_items
		SET quantity = $1, updated_at = NOW()
		WHERE cart_item_id = $2
		RETURNING cart_item_no, cart_item_id, cart_no, variant_no, quantity, unit_price, created_at, updated_at
	`, req.Quantity, cartItemId).Scan(
		&ci.CartItemNo, &ci.CartItemId, &ci.CartNo, &ci.VariantNo,
		&ci.Quantity, &ci.UnitPrice, &ci.CreatedAt, &ci.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("cartRepository.UpdateItem: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("cartRepository.UpdateItem: commit: %w", err)
	}
	return &ci, nil
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartItemId uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("cartRepository.RemoveItem: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT cart_item_no FROM orders.cart_items
		WHERE cart_item_id = $1
		FOR UPDATE
	`, cartItemId).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("cartRepository.RemoveItem: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM orders.cart_items WHERE cart_item_id = $1`, cartItemId)
	if err != nil {
		return fmt.Errorf("cartRepository.RemoveItem: delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("cartRepository.RemoveItem: commit: %w", err)
	}
	return nil
}

func (r *cartRepository) ClearCart(ctx context.Context, cartNo uint32) error {
	_, err := r.db.Exec(ctx, `DELETE FROM orders.cart_items WHERE cart_no = $1`, cartNo)
	if err != nil {
		return fmt.Errorf("cartRepository.ClearCart: %w", err)
	}
	return nil
}
