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

type productVariantRepository struct {
	db *pgxpool.Pool
}

func NewProductVariantRepository(db *pgxpool.Pool) repository.ProductVariantRepository {
	return &productVariantRepository{db: db}
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *productVariantRepository) GetAll(ctx context.Context) ([]models.ProductVariant, error) {
	query := `
		SELECT
			variant_no,
			variant_id,
			product_no,
			sku,
			price_override,
			stock_qty,
			weight_kg,
			image_url,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM product.product_variants
		WHERE status != 'D'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.ProductVariant
	for rows.Next() {
		var d models.ProductVariant

		if err := rows.Scan(
			&d.VariantNo,
			&d.VariantId,
			&d.ProductNo,
			&d.SKU,
			&d.PriceOverride,
			&d.StockQty,
			&d.WeightKg,
			&d.ImageURL,
			&d.Status,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("productVariantRepository.GetAll: scan: %w", err)
		}

		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("productVariantRepository.GetAll: rows: %w", err)
	}

	return data, nil
}

func (r *productVariantRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.ProductVariant, error) {
	query := `
		SELECT
			variant_no,
			variant_id,
			product_no,
			sku,
			price_override,
			stock_qty,
			weight_kg,
			image_url,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM product.product_variants
		WHERE variant_id = $1
		AND status != 'D'
	`

	var d models.ProductVariant

	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.VariantNo,
		&d.VariantId,
		&d.ProductNo,
		&d.SKU,
		&d.PriceOverride,
		&d.StockQty,
		&d.WeightKg,
		&d.ImageURL,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("productVariantRepository.GetById: %w", err)
	}

	return &d, nil
}

func (r *productVariantRepository) FindOneByNo(ctx context.Context, variantNo uint32) (*models.EnrichedVariant, error) {
	query := `
		SELECT
			pv.variant_no,
			pv.variant_id,
			p.product_name,
			pv.sku,
			COALESCE(pv.price_override, p.base_price) AS price,
			pv.stock_qty
		FROM product.product_variants pv
		JOIN product.products p ON p.product_no = pv.product_no
		WHERE pv.variant_no = $1
		AND pv.status = 'active'
	`

	var ev models.EnrichedVariant
	err := r.db.QueryRow(ctx, query, variantNo).Scan(
		&ev.VariantNo,
		&ev.VariantId,
		&ev.ProductName,
		&ev.SKU,
		&ev.Price,
		&ev.StockQty,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("productVariantRepository.FindOneByNo: %w", err)
	}

	// Fetch variant attributes
	rows, err := r.db.Query(ctx, `
		SELECT av.code, av.value
		FROM product.variant_attribute_values vav
		JOIN product.attribute_values av ON av.attribute_value_no = vav.attribute_value_no
		WHERE vav.variant_no = $1
	`, variantNo)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.FindOneByNo: fetch attributes: %w", err)
	}
	defer rows.Close()

	ev.Attributes = make(map[string]interface{})
	for rows.Next() {
		var code, value string
		if err := rows.Scan(&code, &value); err != nil {
			return nil, fmt.Errorf("productVariantRepository.FindOneByNo: scan attribute: %w", err)
		}
		ev.Attributes[code] = value
	}

	return &ev, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

func (r *productVariantRepository) Save(ctx context.Context, productVariant *models.ProductVariant) (*models.ProductVariant, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO product.product_variants (
			variant_id,
			product_no,
			sku,
			price_override,
			stock_qty,
			weight_kg,
			image_url,
			status,
			created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			variant_no,
			variant_id,
			product_no,
			sku,
			price_override,
			stock_qty,
			weight_kg,
			image_url,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.ProductVariant

	err = tx.QueryRow(ctx, query,
		productVariant.VariantId,
		productVariant.ProductNo,
		productVariant.SKU,
		productVariant.PriceOverride,
		productVariant.StockQty,
		productVariant.WeightKg,
		productVariant.ImageURL,
		models.ProductStatusDraft,
		productVariant.CreatedBy,
	).Scan(
		&d.VariantNo,
		&d.VariantId,
		&d.ProductNo,
		&d.SKU,
		&d.PriceOverride,
		&d.StockQty,
		&d.WeightKg,
		&d.ImageURL,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("productVariantRepository.Save: commit: %w", err)
	}

	return &d, nil
}

func (r *productVariantRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateProductVariantRequest) (*models.ProductVariant, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT variant_no FROM product.product_variants
		WHERE variant_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("productVariantRepository.Modify: lock row: %w", err)
	}

	query := `
		UPDATE product.product_variants
		SET
			product_no     = $1,
			sku            = COALESCE(NULLIF($2, ''), sku),
			price_override = $3,
			stock_qty      = $4,
			weight_kg      = $5,
			image_url      = $6,
			status         = $7,
			updated_at     = NOW(),
			updated_by     = $8
		WHERE variant_id = $9
		RETURNING
			variant_no,
			variant_id,
			product_no,
			sku,
			price_override,
			stock_qty,
			weight_kg,
			image_url,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.ProductVariant

	err = tx.QueryRow(ctx, query,
		req.ProductNo,
		req.SKU,
		req.PriceOverride,
		req.StockQty,
		req.WeightKg,
		req.ImageURL,
		req.Status,
		req.UpdatedBy,
		Id,
	).Scan(
		&d.VariantNo,
		&d.VariantId,
		&d.ProductNo,
		&d.SKU,
		&d.PriceOverride,
		&d.StockQty,
		&d.WeightKg,
		&d.ImageURL,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.Modify: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("productVariantRepository.Modify: commit: %w", err)
	}

	return &d, nil
}

func (r *productVariantRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("productVariantRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT variant_no FROM product.product_variants
		WHERE variant_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("productVariantRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE product.product_variants
		SET
			status     = 'D',
			updated_at = NOW()
		WHERE variant_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("productVariantRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("productVariantRepository.Remove: commit: %w", err)
	}

	return nil
}