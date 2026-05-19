package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/repository"
)

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) repository.ProductRepository {
	return &productRepository{db: db}
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *productRepository) GetAll(ctx context.Context) ([]models.Product, error) {
	query := `
		SELECT
			product_no,
			product_id,
			category_no,
			sku,
			product_name,
			slug,
			description,
			base_price,
			status,
			is_featured,
			stock_qty,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM product.products
		WHERE status != 'D'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("productRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.Product
	for rows.Next() {
		var d models.Product

		if err := rows.Scan(
			&d.ProductNo,
			&d.ProductId,
			&d.CategoryNo,
			&d.SKU,
			&d.ProductName,
			&d.Slug,
			&d.Description,
			&d.BasePrice,
			&d.Status,
			&d.IsFeatured,
			&d.StockQty,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("productRepository.GetAll: scan: %w", err)
		}

		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("productRepository.GetAll: rows: %w", err)
	}

	return data, nil
}

func (r *productRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT
			product_no,
			product_id,
			category_no,
			sku,
			product_name,
			slug,
			description,
			base_price,
			status,
			is_featured,
			stock_qty,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM product.products
		WHERE product_id = $1
		AND status != 'D'
	`

	var d models.Product

	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.ProductNo,
		&d.ProductId,
		&d.CategoryNo,
		&d.SKU,
		&d.ProductName,
		&d.Slug,
		&d.Description,
		&d.BasePrice,
		&d.Status,
		&d.IsFeatured,
		&d.StockQty,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("productRepository.GetById: %w", err)
	}

	return &d, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

func (r *productRepository) Save(ctx context.Context, product *models.Product) (*models.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("productRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO product.products (
			product_id,
			category_no,
			sku,
			product_name,
			slug,
			description,
			base_price,
			status,
			is_featured,
			stock_qty,
			created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING
			product_no,
			product_id,
			category_no,
			sku,
			product_name,
			slug,
			description,
			base_price,
			status,
			is_featured,
			stock_qty,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Product

	err = tx.QueryRow(ctx, query,
		product.ProductId,
		product.CategoryNo,
		product.SKU,
		product.ProductName,
		product.Slug,
		product.Description,
		product.BasePrice,
		models.ProductStatusDraft,
		product.IsFeatured,
		product.StockQty,
		product.CreatedBy,
	).Scan(
		&d.ProductNo,
		&d.ProductId,
		&d.CategoryNo,
		&d.SKU,
		&d.ProductName,
		&d.Slug,
		&d.Description,
		&d.BasePrice,
		&d.Status,
		&d.IsFeatured,
		&d.StockQty,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("productRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("productRepository.Save: commit: %w", err)
	}

	return &d, nil
}

func (r *productRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateProductRequest) (*models.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("productRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT product_no FROM product.products
		WHERE product_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("productRepository.Modify: lock row: %w", err)
	}

	query := `
		UPDATE product.products
		SET
			category_no  = $1,
			sku          = COALESCE(NULLIF($2, ''), sku),
			product_name = COALESCE(NULLIF($3, ''), product_name),
			slug         = COALESCE(NULLIF($4, ''), slug),
			description  = COALESCE(NULLIF($5, ''), description),
			base_price   = $6,
			status       = $7,
			is_featured  = $8,
			stock_qty    = $9,
			updated_at   = NOW(),
			updated_by   = $10
		WHERE product_id = $11
		RETURNING
			product_no,
			product_id,
			category_no,
			sku,
			product_name,
			slug,
			description,
			base_price,
			status,
			is_featured,
			stock_qty,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Product

	err = tx.QueryRow(ctx, query,
		req.CategoryNo,
		req.SKU,
		req.ProductName,
		req.Slug,
		req.Description,
		req.BasePrice,
		req.Status,
		req.IsFeatured,
		req.StockQty,
		req.UpdatedBy,
		Id,
	).Scan(
		&d.ProductNo,
		&d.ProductId,
		&d.CategoryNo,
		&d.SKU,
		&d.ProductName,
		&d.Slug,
		&d.Description,
		&d.BasePrice,
		&d.Status,
		&d.IsFeatured,
		&d.StockQty,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("productRepository.Modify: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("productRepository.Modify: commit: %w", err)
	}

	return &d, nil
}

func (r *productRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("productRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT product_no FROM product.products
		WHERE product_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("productRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE product.products
		SET
			status     = 'D',
			updated_at = NOW()
		WHERE product_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("productRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("productRepository.Remove: commit: %w", err)
	}

	return nil
}