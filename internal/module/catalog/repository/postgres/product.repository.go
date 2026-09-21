package postgres

import (
	"context"
	"encoding/json"
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

// ── shared scan ───────────────────────────────────────────────────────────────

func scanProduct(row interface{ Scan(...any) error }, d *models.Product) error {
	var featuresJSON []byte
	var specificationsJSON []byte

	err := row.Scan(
		&d.ProductNo,
		&d.ProductId,
		&d.CategoryNo,
		&d.CategoryId,
		&d.SKU,
		&d.ProductName,
		&d.Slug,
		&d.Description,
		&d.ImageURL,
		&featuresJSON,
		&specificationsJSON,
		&d.Status,
		&d.IsFeatured,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
		&d.DefaultVariantId,    // ← add
		&d.DefaultSellingPrice, // ← add
		&d.DefaultMRP,          // ← add
		&d.DefaultStockQty,     // ← add
	)
	if err != nil {
		return err
	}

	if len(featuresJSON) > 0 {
		_ = json.Unmarshal(featuresJSON, &d.Features)
	}
	if d.Features == nil {
		d.Features = []string{}
	}

	if len(specificationsJSON) > 0 {
		_ = json.Unmarshal(specificationsJSON, &d.Specifications)
	}
	if d.Specifications == nil {
		d.Specifications = []models.SpecItem{}
	}

	return nil
}

// productSelect — requires JOIN with catalog.categories c
const productSelect = `
	p.product_no,
	p.product_id,
	p.category_no,
	c.category_id,
	p.sku,
	p.product_name,
	p.slug,
	p.description,
	p.image_url,
	p.features,
	p.specifications,
	p.status,
	p.is_featured,
	p.created_at,
	p.created_by,
	p.updated_at,
	p.updated_by,
	(SELECT pv.variant_id FROM catalog.product_variants pv
	 WHERE pv.product_no = p.product_no AND pv.status = 'active'
	 ORDER BY COALESCE(pv.selling_price, pv.mrp, 0) ASC LIMIT 1) AS default_variant_id,
	(SELECT pv.selling_price FROM catalog.product_variants pv
	 WHERE pv.product_no = p.product_no AND pv.status = 'active'
	 ORDER BY COALESCE(pv.selling_price, pv.mrp, 0) ASC LIMIT 1) AS default_selling_price,
	(SELECT pv.mrp FROM catalog.product_variants pv
	 WHERE pv.product_no = p.product_no AND pv.status = 'active'
	 ORDER BY COALESCE(pv.selling_price, pv.mrp, 0) ASC LIMIT 1) AS default_mrp,
	(SELECT pv.stock_qty FROM catalog.product_variants pv
	 WHERE pv.product_no = p.product_no AND pv.status = 'active'
	 ORDER BY COALESCE(pv.selling_price, pv.mrp, 0) ASC LIMIT 1) AS default_stock_qty
`

// ── READ operations ───────────────────────────────────────────────────────────

func (r *productRepository) GetAll(ctx context.Context) ([]models.Product, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+productSelect+`
		FROM catalog.products p
		JOIN catalog.categories c ON c.category_no = p.category_no
		WHERE p.status != 'D'
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("productRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.Product
	for rows.Next() {
		var d models.Product
		if err := scanProduct(rows, &d); err != nil {
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
	var d models.Product
	err := scanProduct(r.db.QueryRow(ctx, `
		SELECT `+productSelect+`
		FROM catalog.products p
		JOIN catalog.categories c ON c.category_no = p.category_no
		WHERE p.product_id = $1
		AND p.status != 'D'
	`, Id), &d)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("productRepository.GetById: %w", err)
	}
	return &d, nil
}

func (r *productRepository) GetByCategory(ctx context.Context, categoryId uuid.UUID) ([]models.Product, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+productSelect+`
		FROM catalog.products p
		JOIN catalog.categories c ON c.category_no = p.category_no
		WHERE c.category_id = $1
		AND p.status != 'D'
		ORDER BY p.created_at DESC
	`, categoryId)
	if err != nil {
		return nil, fmt.Errorf("productRepository.GetByCategory: %w", err)
	}
	defer rows.Close()

	var data []models.Product
	for rows.Next() {
		var d models.Product
		if err := scanProduct(rows, &d); err != nil {
			return nil, fmt.Errorf("productRepository.GetByCategory: scan: %w", err)
		}
		data = append(data, d)
	}

	return data, rows.Err()
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

	// Resolve category_no from category_id
	var categoryNo uint32
	err = tx.QueryRow(ctx, `
		SELECT category_no FROM catalog.categories
		WHERE category_id = $1 AND status != 'D'
	`, product.CategoryId).Scan(&categoryNo)
	if err != nil {
		return nil, fmt.Errorf("productRepository.Save: resolve category: %w", err)
	}

	featuresJSON, _ := json.Marshal(product.Features)
	specificationsJSON, _ := json.Marshal(product.Specifications)

	var d models.Product
	err = scanProduct(tx.QueryRow(ctx, `
		INSERT INTO catalog.products (
			product_id, category_no, sku, product_name, slug,
			description, image_url, features, specifications,
			status, is_featured, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING
			product_no, product_id, category_no,
			(SELECT category_id FROM catalog.categories WHERE category_no = $2),
			sku, product_name, slug, description, image_url,
			features, specifications, status, is_featured,
			created_at, created_by, updated_at, updated_by,
			NULL::uuid,    -- default_variant_id
			NULL::numeric, -- default_selling_price
			NULL::numeric, -- default_mrp
			NULL::integer  -- default_stock_qty
		`,
		product.ProductId,
		categoryNo, // ← resolved
		product.SKU,
		product.ProductName,
		product.Slug,
		product.Description,
		product.ImageURL,
		featuresJSON,
		specificationsJSON,
		models.ProductStatusDraft,
		product.IsFeatured,
		product.CreatedBy,
	), &d)
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
		SELECT product_no FROM catalog.products
		WHERE product_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("productRepository.Modify: lock row: %w", err)
	}

	// Resolve category_no from category_id
	var categoryNo uint32
	err = tx.QueryRow(ctx, `
		SELECT category_no FROM catalog.categories
		WHERE category_id = $1 AND status != 'D'
	`, req.CategoryId).Scan(&categoryNo)
	if err != nil {
		return nil, fmt.Errorf("productRepository.Modify: resolve category: %w", err)
	}

	featuresJSON, _ := json.Marshal(req.Features)
	specificationsJSON, _ := json.Marshal(req.Specifications)

	var d models.Product
	err = scanProduct(tx.QueryRow(ctx, `
		UPDATE catalog.products SET
			category_no    = $1,
			sku            = COALESCE(NULLIF($2, ''), sku),
			product_name   = COALESCE(NULLIF($3, ''), product_name),
			slug           = COALESCE(NULLIF($4, ''), slug),
			description    = COALESCE(NULLIF($5, ''), description),
			features       = $6,
			specifications = $7,
			status         = $8,
			is_featured    = $9,
			updated_at     = NOW(),
			updated_by     = $10
		WHERE product_id = $11
		RETURNING
			product_no, product_id, category_no,
			(SELECT category_id FROM catalog.categories WHERE category_no = $1),
			sku, product_name, slug, description, image_url,
			features, specifications, status, is_featured,
			created_at, created_by, updated_at, updated_by
	`,
		categoryNo, // ← resolved
		req.SKU,
		req.ProductName,
		req.Slug,
		req.Description,
		featuresJSON,
		specificationsJSON,
		req.Status,
		req.IsFeatured,
		req.UpdatedBy,
		Id,
	), &d)
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
		SELECT product_no FROM catalog.products
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
		UPDATE catalog.products
		SET status = 'D', updated_at = NOW()
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

func (r *productRepository) ModifyImage(ctx context.Context, Id uuid.UUID, imageURL string) (*models.Product, error) {
	var d models.Product
	err := scanProduct(r.db.QueryRow(ctx, `
		UPDATE catalog.products p
		SET image_url = $1, updated_at = NOW()
		WHERE p.product_id = $2
		RETURNING
			p.product_no,
			p.product_id,
			p.category_no,
			(SELECT category_id FROM catalog.categories WHERE category_no = p.category_no),
			p.sku, p.product_name, p.slug, p.description, p.base_price, p.image_url,
			p.features, p.specifications, p.status, p.is_featured, p.stock_qty,
			p.created_at, p.created_by, p.updated_at, p.updated_by
	`, imageURL, Id), &d)
	if err != nil {
		return nil, fmt.Errorf("productRepository.ModifyImage: %w", err)
	}
	return &d, nil
}

func (r *productRepository) GetAttributesByProductId(ctx context.Context, productId uuid.UUID) ([]models.ProductAttribute, error) {
	query := `
		SELECT DISTINCT
			a.attribute_id::text,
			a.attribute_name,
			a.code,
			a.type::text,
			av.attribute_value_id::text,
			av.display_label,
			av.value,
			COALESCE(av.color_hex, '') AS color_hex
		FROM catalog.products p
		JOIN catalog.product_variants pv          ON pv.product_no          = p.product_no
		JOIN catalog.variant_attribute_values vav ON vav.variant_no         = pv.variant_no
		JOIN catalog.attribute_values av           ON av.attribute_value_no = vav.attribute_value_no
		JOIN catalog.attributes a                  ON a.attribute_no        = av.attribute_no
		WHERE p.product_id = $1
		AND pv.status = 'active'
		ORDER BY a.attribute_name, av.display_label
	`

	rows, err := r.db.Query(ctx, query, productId)
	if err != nil {
		return nil, fmt.Errorf("productRepository.GetAttributesByProductId: %w", err)
	}
	defer rows.Close()

	attrMap := map[string]*models.ProductAttribute{}
	attrOrder := []string{}

	for rows.Next() {
		var (
			attrId, attrName, code, attrType     string
			valId, displayLabel, value, colorHex string
		)
		if err := rows.Scan(&attrId, &attrName, &code, &attrType, &valId, &displayLabel, &value, &colorHex); err != nil {
			return nil, fmt.Errorf("productRepository.GetAttributesByProductId: scan: %w", err)
		}

		if _, exists := attrMap[attrId]; !exists {
			attrMap[attrId] = &models.ProductAttribute{
				AttributeId:   attrId,
				AttributeName: attrName,
				Code:          code,
				Type:          attrType,
				Values:        []models.AttributeValueOption{},
			}
			attrOrder = append(attrOrder, attrId)
		}

		attrMap[attrId].Values = append(attrMap[attrId].Values, models.AttributeValueOption{
			AttributeValueId: valId,
			DisplayLabel:     displayLabel,
			Value:            value,
			ColorHex:         colorHex,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("productRepository.GetAttributesByProductId: rows: %w", err)
	}

	result := make([]models.ProductAttribute, 0, len(attrOrder))
	for _, id := range attrOrder {
		result = append(result, *attrMap[id])
	}
	return result, nil
}
