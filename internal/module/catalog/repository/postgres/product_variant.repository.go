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
	rows, err := r.db.Query(ctx, `
		SELECT
			pv.variant_no,
			pv.variant_id,
			pv.product_no,
			p.product_id,
			p.product_name,
			p.category_no,
			c.category_name,
			COALESCE(
				(SELECT pvi.image_url FROM catalog.product_variant_images pvi
				 WHERE pvi.variant_no = pv.variant_no AND pvi.is_primary = TRUE LIMIT 1),
				(SELECT pvi.image_url FROM catalog.product_variant_images pvi
				 WHERE pvi.variant_no = pv.variant_no ORDER BY pvi.sort_order ASC LIMIT 1)
			) AS display_image_url,
			pv.sku,
			pv.cost_price,
			pv.mrp,
			pv.selling_price,
			pv.stock_qty,
			pv.weight_kg,
			pv.image_url AS variant_image_url,
			pv.status,
			pv.created_at,
			pv.created_by,
			pv.updated_at,
			pv.updated_by
		FROM catalog.product_variants pv
		JOIN catalog.products   p ON p.product_no  = pv.product_no
		JOIN catalog.categories c ON c.category_no = p.category_no
		WHERE pv.status != 'D'
		ORDER BY pv.created_at DESC
	`)
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
			&d.ProductId,
			&d.ProductName,
			&d.CategoryNo,
			&d.CategoryName,
			&d.ProductImageURL,
			&d.SKU,
			&d.CostPrice,    // ← add
			&d.MRP,          // ← add
			&d.SellingPrice, // ← add
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
	var d models.ProductVariant
	err := r.db.QueryRow(ctx, `
		SELECT
			variant_no,
			variant_id,
			product_no,
			sku,
			cost_price,
			mrp,
			selling_price,
			stock_qty,
			weight_kg,
			image_url,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM catalog.product_variants
		WHERE variant_id = $1
		AND status != 'D'
	`, Id).Scan(
		&d.VariantNo,
		&d.VariantId,
		&d.ProductNo,
		&d.SKU,
		&d.CostPrice,    // ← add
		&d.MRP,          // ← add
		&d.SellingPrice, // ← add
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

func (r *productVariantRepository) FindOneById(ctx context.Context, variantId uuid.UUID) (*models.EnrichedVariant, error) {
	var ev models.EnrichedVariant
	err := r.db.QueryRow(ctx, `
		SELECT
			pv.variant_no,
			pv.variant_id,
			p.product_name,
			pv.sku,
			COALESCE(pv.selling_price, pv.mrp, pv.cost_price, 0) AS price,
			pv.stock_qty
		FROM catalog.product_variants pv
		JOIN catalog.products p ON p.product_no = pv.product_no
		WHERE pv.variant_id = $1
		AND pv.status = 'active'
	`, variantId).Scan(
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
		return nil, fmt.Errorf("productVariantRepository.FindOneById: %w", err)
	}

	// Fetch attributes
	attrRows, err := r.db.Query(ctx, `
		SELECT a.attribute_name, av.display_label
		FROM catalog.variant_attribute_values vav
		JOIN catalog.attribute_values av ON av.attribute_value_no = vav.attribute_value_no
		JOIN catalog.attributes       a  ON a.attribute_no        = av.attribute_no
		WHERE vav.variant_no = $1
	`, ev.VariantNo)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.FindOneById: fetch attributes: %w", err)
	}
	defer attrRows.Close()

	ev.Attributes = make(map[string]interface{})
	for attrRows.Next() {
		var attrName, displayLabel string
		if err := attrRows.Scan(&attrName, &displayLabel); err != nil {
			return nil, fmt.Errorf("productVariantRepository.FindOneById: scan attribute: %w", err)
		}
		ev.Attributes[attrName] = displayLabel
	}

	if err := attrRows.Err(); err != nil {
		return nil, fmt.Errorf("productVariantRepository.FindOneById: rows error: %w", err)
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

	var d models.ProductVariant
	err = tx.QueryRow(ctx, `
		INSERT INTO catalog.product_variants (
			variant_id, product_no, sku,
			cost_price, mrp, selling_price,
			stock_qty, weight_kg, image_url,
			status, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING
			variant_no, variant_id, product_no, sku,
			cost_price, mrp, selling_price,
			stock_qty, weight_kg, image_url,
			status, created_at, created_by, updated_at, updated_by
	`,
		productVariant.VariantId,
		productVariant.ProductNo,
		productVariant.SKU,
		productVariant.CostPrice,
		productVariant.MRP,
		productVariant.SellingPrice,
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
		&d.CostPrice,
		&d.MRP,
		&d.SellingPrice,
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
		SELECT variant_no FROM catalog.product_variants
		WHERE variant_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("productVariantRepository.Modify: lock row: %w", err)
	}

	var pv models.ProductVariant
	err = tx.QueryRow(ctx, `
		UPDATE catalog.product_variants SET
			sku            = COALESCE(NULLIF($1, ''), sku),
			cost_price     = $2,
			mrp            = $3,
			selling_price  = $4,
			stock_qty      = $5,
			weight_kg      = $6,
			image_url      = $7,
			status         = CASE WHEN $8::text = '' THEN status ELSE $8::catalog.product_status_enum END,
			updated_at     = NOW(),
			updated_by     = $9
		WHERE variant_id = $10
		RETURNING
			variant_no, variant_id, product_no,
			sku, cost_price, mrp, selling_price,
			stock_qty, weight_kg, image_url,
			status, created_at, created_by, updated_at, updated_by
	`,
		req.SKU,
		req.CostPrice,
		req.MRP,
		req.SellingPrice,
		req.StockQty,
		req.WeightKg,
		req.ImageURL,
		string(req.Status),
		req.UpdatedBy,
		Id,
	).Scan(
		&pv.VariantNo,
		&pv.VariantId,
		&pv.ProductNo,
		&pv.SKU,
		&pv.CostPrice,
		&pv.MRP,
		&pv.SellingPrice,
		&pv.StockQty,
		&pv.WeightKg,
		&pv.ImageURL,
		&pv.Status,
		&pv.CreatedAt,
		&pv.CreatedBy,
		&pv.UpdatedAt,
		&pv.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.Modify: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("productVariantRepository.Modify: commit: %w", err)
	}

	return &pv, nil
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
		SELECT variant_no FROM catalog.product_variants
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
		UPDATE catalog.product_variants
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

func (r *productVariantRepository) GetByProductId(ctx context.Context, productId uuid.UUID) ([]models.ProductVariant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			pv.variant_no,
			pv.variant_id,
			pv.product_no,
			pv.sku,
			pv.cost_price,
			pv.mrp,
			pv.selling_price,
			pv.stock_qty,
			pv.weight_kg,
			pv.image_url,
			pv.status,
			pv.created_at,
			pv.created_by,
			pv.updated_at,
			pv.updated_by
		FROM catalog.product_variants pv
		JOIN catalog.products p ON p.product_no = pv.product_no
		WHERE p.product_id = $1
		AND pv.status != 'D'
		ORDER BY pv.created_at ASC
	`, productId)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.GetByProductId: %w", err)
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
			&d.CostPrice,    // ← add
			&d.MRP,          // ← add
			&d.SellingPrice, // ← add
			&d.StockQty,
			&d.WeightKg,
			&d.ImageURL,
			&d.Status,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("productVariantRepository.GetByProductId: scan: %w", err)
		}
		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("productVariantRepository.GetByProductId: rows: %w", err)
	}
	return data, nil
}

func (r *productVariantRepository) GetByProductIdWithAttributes(ctx context.Context, productId uuid.UUID) ([]models.VariantWithAttributes, error) {
	query := `
		SELECT
			pv.variant_id,
			pv.sku,
			pv.stock_qty,
			pv.weight_kg,
			pv.status::text,
			a.code,
			av.value
		FROM catalog.product_variants pv
		JOIN catalog.products p ON p.product_no = pv.product_no
		LEFT JOIN catalog.variant_attribute_values vav ON vav.variant_no = pv.variant_no
		LEFT JOIN catalog.attribute_values av           ON av.attribute_value_no = vav.attribute_value_no
		LEFT JOIN catalog.attributes a                  ON a.attribute_no = av.attribute_no
		WHERE p.product_id = $1
		AND pv.status = 'active'
		ORDER BY pv.variant_id, a.code
	`

	rows, err := r.db.Query(ctx, query, productId)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.GetByProductIdWithAttributes: %w", err)
	}
	defer rows.Close()

	variantMap := map[uuid.UUID]*models.VariantWithAttributes{}
	variantOrder := []uuid.UUID{}

	for rows.Next() {
		var (
			variantId uuid.UUID
			sku       string
			stockQty  int32
			weightKg  *float64
			status    string
			code      *string
			value     *string
		)
		if err := rows.Scan(&variantId, &sku, &stockQty, &weightKg, &status, &code, &value); err != nil {
			return nil, fmt.Errorf("productVariantRepository.GetByProductIdWithAttributes: scan: %w", err)
		}

		if _, exists := variantMap[variantId]; !exists {
			variantMap[variantId] = &models.VariantWithAttributes{
				VariantId:  variantId,
				SKU:        sku,
				StockQty:   stockQty,
				WeightKg:   weightKg,
				Status:     models.ProductStatus(status),
				Attributes: map[string]string{},
			}
			variantOrder = append(variantOrder, variantId)
		}

		if code != nil && value != nil {
			variantMap[variantId].Attributes[*code] = *value
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("productVariantRepository.GetByProductIdWithAttributes: rows: %w", err)
	}

	result := make([]models.VariantWithAttributes, 0, len(variantOrder))
	for _, id := range variantOrder {
		result = append(result, *variantMap[id])
	}
	return result, nil
}

func (r *productVariantRepository) GetProductNoByProductId(ctx context.Context, productId uuid.UUID) (uint32, error) {
	var productNo uint32
	err := r.db.QueryRow(ctx, `
		SELECT product_no FROM catalog.products
		WHERE product_id = $1 AND status != 'D'
	`, productId).Scan(&productNo)
	if err != nil {
		return 0, fmt.Errorf("productVariantRepository.GetProductNoByProductId: %w", err)
	}
	return productNo, nil
}

func (r *productVariantRepository) GetListing(ctx context.Context, categoryId *uuid.UUID, attributeValues []string) ([]models.VariantListing, error) {

	args := []any{}

	// Build category filter
	categoryFilter := ""
	if len(attributeValues) > 0 {
		// $1 = attributeValues, $2 = categoryId (if present)
		args = append(args, attributeValues)
		if categoryId != nil {
			categoryFilter = `
				AND p.product_no IN (
					SELECT p2.product_no FROM catalog.products p2
					JOIN catalog.categories c2 ON c2.category_no = p2.category_no
					WHERE c2.category_id = $2
					   OR c2.parent_no IN (
					       SELECT category_no FROM catalog.categories
					       WHERE category_id = $2 AND status != 'D'
					   )
				)`
			args = append(args, *categoryId)
		}
	} else {
		// $1 = categoryId (if present)
		if categoryId != nil {
			categoryFilter = `
				AND p.product_no IN (
					SELECT p2.product_no FROM catalog.products p2
					JOIN catalog.categories c2 ON c2.category_no = p2.category_no
					WHERE c2.category_id = $1
					   OR c2.parent_no IN (
					       SELECT category_no FROM catalog.categories
					       WHERE category_id = $1 AND status != 'D'
					   )
				)`
			args = append(args, *categoryId)
		}
	}

	var query string

	if len(attributeValues) > 0 {
		// With attribute filter:
		// 1. Find products that have ANY variant matching selected values
		// 2. Pick one default variant per product, preferring one that matches
		query = fmt.Sprintf(`
			WITH default_variants AS (
				SELECT DISTINCT ON (p.product_id)
					pv.variant_id,
					p.product_id,
					p.product_name,
					p.slug,
					pv.sku,
					pv.cost_price,
					pv.mrp,
					pv.selling_price,
					pv.stock_qty,
					COALESCE(
						(SELECT pvi.image_url FROM catalog.product_variant_images pvi
						 WHERE pvi.variant_no = pv.variant_no AND pvi.is_primary = TRUE LIMIT 1),
						p.image_url
					) AS image_url,
					c.category_id,
					p.is_featured
				FROM catalog.product_variants pv
				JOIN catalog.products   p ON p.product_no  = pv.product_no
				JOIN catalog.categories c ON c.category_no = p.category_no
				WHERE pv.status = 'active'
				AND   p.status  = 'active'
				AND p.product_no IN (
					SELECT DISTINCT p3.product_no
					FROM catalog.products p3
					JOIN catalog.product_variants pv3 ON pv3.product_no = p3.product_no AND pv3.status = 'active'
					JOIN catalog.variant_attribute_values vav3 ON vav3.variant_no = pv3.variant_no
					JOIN catalog.attribute_values av3 ON av3.attribute_value_no = vav3.attribute_value_no
					WHERE av3.value = ANY($1)
				)
				%s
				ORDER BY
					p.product_id,
					-- Prefer variant that matches the selected attribute values
					CASE WHEN EXISTS (
						SELECT 1
						FROM catalog.variant_attribute_values vav_check
						JOIN catalog.attribute_values av_check ON av_check.attribute_value_no = vav_check.attribute_value_no
						WHERE vav_check.variant_no = pv.variant_no
						AND av_check.value = ANY($1)
					) THEN 0 ELSE 1 END ASC,
					pv.stock_qty DESC,
					COALESCE(pv.selling_price, pv.mrp, 0) ASC
			)
			SELECT
				dv.variant_id,
				dv.product_id,
				dv.product_name,
				dv.slug,
				dv.sku,
				dv.cost_price,
				dv.mrp,
				dv.selling_price,
				dv.stock_qty,
				dv.image_url,
				dv.category_id,
				dv.is_featured,
				a.code,
				av.value
			FROM default_variants dv
			LEFT JOIN catalog.product_variants         pv  ON pv.variant_id          = dv.variant_id
			LEFT JOIN catalog.variant_attribute_values vav ON vav.variant_no         = pv.variant_no
			LEFT JOIN catalog.attribute_values         av  ON av.attribute_value_no  = vav.attribute_value_no
			LEFT JOIN catalog.attributes               a   ON a.attribute_no         = av.attribute_no
			ORDER BY dv.is_featured DESC, dv.stock_qty DESC, dv.variant_id, a.code
		`, categoryFilter)

	} else {
		// No attribute filter — one default variant per product (in-stock first, lowest price)
		query = fmt.Sprintf(`
			WITH default_variants AS (
				SELECT DISTINCT ON (p.product_id)
					pv.variant_id,
					p.product_id,
					p.product_name,
					p.slug,
					pv.sku,
					pv.cost_price,
					pv.mrp,
					pv.selling_price,
					pv.stock_qty,
					COALESCE(
						(SELECT pvi.image_url FROM catalog.product_variant_images pvi
						 WHERE pvi.variant_no = pv.variant_no AND pvi.is_primary = TRUE LIMIT 1),
						p.image_url
					) AS image_url,
					c.category_id,
					p.is_featured
				FROM catalog.product_variants pv
				JOIN catalog.products   p ON p.product_no  = pv.product_no
				JOIN catalog.categories c ON c.category_no = p.category_no
				WHERE pv.status = 'active'
				AND   p.status  = 'active'
				%s
				ORDER BY
					p.product_id,
					pv.stock_qty DESC,
					COALESCE(pv.selling_price, pv.mrp, 0) ASC
			)
			SELECT
				dv.variant_id,
				dv.product_id,
				dv.product_name,
				dv.slug,
				dv.sku,
				dv.cost_price,
				dv.mrp,
				dv.selling_price,
				dv.stock_qty,
				dv.image_url,
				dv.category_id,
				dv.is_featured,
				a.code,
				av.value
			FROM default_variants dv
			LEFT JOIN catalog.product_variants         pv  ON pv.variant_id          = dv.variant_id
			LEFT JOIN catalog.variant_attribute_values vav ON vav.variant_no         = pv.variant_no
			LEFT JOIN catalog.attribute_values         av  ON av.attribute_value_no  = vav.attribute_value_no
			LEFT JOIN catalog.attributes               a   ON a.attribute_no         = av.attribute_no
			ORDER BY dv.is_featured DESC, dv.stock_qty DESC, dv.variant_id, a.code
		`, categoryFilter)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("productVariantRepository.GetListing: %w", err)
	}
	defer rows.Close()

	variantMap := map[uuid.UUID]*models.VariantListing{}
	variantOrder := []uuid.UUID{}

	for rows.Next() {
		var (
			variantId, productId, catId  uuid.UUID
			productName, slug, sku       string
			costPrice, mrp, sellingPrice *float64
			stockQty                     int32
			imageURL                     *string
			isFeatured                   bool
			code, value                  *string
		)
		if err := rows.Scan(
			&variantId, &productId, &productName, &slug, &sku,
			&costPrice, &mrp, &sellingPrice, &stockQty,
			&imageURL, &catId, &isFeatured,
			&code, &value,
		); err != nil {
			return nil, fmt.Errorf("productVariantRepository.GetListing: scan: %w", err)
		}

		if _, exists := variantMap[variantId]; !exists {
			variantMap[variantId] = &models.VariantListing{
				VariantId:    variantId,
				ProductId:    productId,
				ProductName:  productName,
				Slug:         slug,
				SKU:          sku,
				CostPrice:    costPrice,
				MRP:          mrp,
				SellingPrice: sellingPrice,
				StockQty:     stockQty,
				ImageURL:     imageURL,
				CategoryId:   catId,
				IsFeatured:   isFeatured,
				Attributes:   map[string]string{},
			}
			variantOrder = append(variantOrder, variantId)
		}

		if code != nil && value != nil {
			variantMap[variantId].Attributes[*code] = *value
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("productVariantRepository.GetListing: rows: %w", err)
	}

	result := make([]models.VariantListing, 0, len(variantOrder))
	for _, id := range variantOrder {
		result = append(result, *variantMap[id])
	}
	return result, nil
}
