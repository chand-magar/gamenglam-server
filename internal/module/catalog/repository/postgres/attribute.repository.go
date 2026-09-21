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

type attributeRepository struct {
	db *pgxpool.Pool
}

func NewAttributeRepository(db *pgxpool.Pool) repository.AttributeRepository {
	return &attributeRepository{db: db}
}

// attributeSelect requires JOIN with catalog.categories c
const attributeSelect = `
	a.attribute_no,
	a.attribute_id,
	a.category_no,
	c.category_id,
	a.attribute_name,
	a.code,
	a.type,
	a.is_required,
	a.is_variant,
	a.show_in_filter,
	a.sort_order,
	a.status,
	a.created_at,
	a.created_by,
	a.updated_at,
	a.updated_by
`

func scanAttribute(row interface{ Scan(...any) error }, d *models.Attribute) error {
	return row.Scan(
		&d.AttributeNo,
		&d.AttributeId,
		&d.CategoryNo,
		&d.CategoryId,
		&d.AttributeName,
		&d.Code,
		&d.Type,
		&d.IsRequired,
		&d.IsVariant,
		&d.ShowInFilter, // ← add
		&d.SortOrder,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *attributeRepository) GetAll(ctx context.Context) ([]models.Attribute, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+attributeSelect+`
		FROM catalog.attributes a
		JOIN catalog.categories c ON c.category_no = a.category_no
		WHERE a.status != 'D'
		ORDER BY a.sort_order ASC, a.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("attributeRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.Attribute
	for rows.Next() {
		var d models.Attribute
		if err := scanAttribute(rows, &d); err != nil {
			return nil, fmt.Errorf("attributeRepository.GetAll: scan: %w", err)
		}
		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("attributeRepository.GetAll: rows: %w", err)
	}

	return data, nil
}
func (r *attributeRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.Attribute, error) {
	var d models.Attribute
	err := scanAttribute(r.db.QueryRow(ctx, `
		SELECT `+attributeSelect+`
		FROM catalog.attributes a
		JOIN catalog.categories c ON c.category_no = a.category_no
		WHERE a.attribute_id = $1
		AND a.status != 'D'
	`, Id), &d)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("attributeRepository.GetById: %w", err)
	}
	return &d, nil
}

// ── WRITE operations ──────────────────────────────────────────────────────────

func (r *attributeRepository) Save(ctx context.Context, attribute *models.Attribute) (*models.Attribute, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("attributeRepository.Save: begin tx: %w", err)
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
	`, attribute.CategoryId).Scan(&categoryNo)
	if err != nil {
		return nil, fmt.Errorf("attributeRepository.Save: resolve category: %w", err)
	}

	var d models.Attribute
	err = scanAttribute(tx.QueryRow(ctx, `
		INSERT INTO catalog.attributes (
			attribute_id, category_no, attribute_name, code,
			type, is_required, is_variant, show_in_filter, sort_order, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING
			attribute_no, attribute_id, category_no,
			(SELECT category_id FROM catalog.categories WHERE category_no = $2),
			attribute_name, code, type, is_required, is_variant,
			show_in_filter,  -- ← add between is_variant and sort_order
			sort_order, status, created_at, created_by, updated_at, updated_by
	`,
		attribute.AttributeId,
		categoryNo,
		attribute.AttributeName,
		attribute.Code,
		attribute.Type,
		attribute.IsRequired,
		attribute.IsVariant,
		attribute.ShowInFilter,
		attribute.SortOrder,
		attribute.CreatedBy,
	), &d)
	if err != nil {
		return nil, fmt.Errorf("attributeRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("attributeRepository.Save: commit: %w", err)
	}

	return &d, nil
}

func (r *attributeRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateAttributeRequest) (*models.Attribute, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("attributeRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT attribute_no FROM catalog.attributes
		WHERE attribute_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("attributeRepository.Modify: lock row: %w", err)
	}

	// Resolve category_no from category_id
	var categoryNo uint32
	err = tx.QueryRow(ctx, `
		SELECT category_no FROM catalog.categories
		WHERE category_id = $1 AND status != 'D'
	`, req.CategoryId).Scan(&categoryNo)
	if err != nil {
		return nil, fmt.Errorf("attributeRepository.Modify: resolve category: %w", err)
	}

	var d models.Attribute
	err = scanAttribute(tx.QueryRow(ctx, `
		UPDATE catalog.attributes SET
			category_no    = $1,
			attribute_name = COALESCE(NULLIF($2, ''), attribute_name),
			code           = COALESCE(NULLIF($3, ''), code),
			type           = $4,
			is_required    = $5,
			is_variant     = $6,
			show_in_filter = $7,
			sort_order     = $8,
			updated_at     = NOW(),
			updated_by     = $9
		WHERE attribute_id = $10
		RETURNING
			attribute_no, attribute_id, category_no,
			(SELECT category_id FROM catalog.categories WHERE category_no = $1),
			attribute_name, code, type, is_required, is_variant,
			show_in_filter,  -- ← add
			sort_order, status, created_at, created_by, updated_at, updated_by
	`,
		categoryNo,
		req.AttributeName,
		req.Code,
		req.Type,
		req.IsRequired,
		req.IsVariant,
		req.ShowInFilter,
		req.SortOrder,
		req.UpdatedBy,
		Id,
	), &d)
	if err != nil {
		return nil, fmt.Errorf("attributeRepository.Modify: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("attributeRepository.Modify: commit: %w", err)
	}

	return &d, nil
}

func (r *attributeRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("attributeRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT attribute_no FROM catalog.attributes
		WHERE attribute_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("attributeRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE catalog.attributes
		SET status = 'D', updated_at = NOW()
		WHERE attribute_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("attributeRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("attributeRepository.Remove: commit: %w", err)
	}

	return nil
}

func (r *attributeRepository) GetByCategoryWithValues(ctx context.Context, categoryId uuid.UUID) ([]models.AttributeWithValues, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			a.attribute_id::text,
			a.attribute_name,
			a.code,
			a.type::text,
			av.attribute_value_id::text,
			av.display_label,
			av.value,
			COALESCE(av.color_hex, '') AS color_hex
		FROM catalog.attributes a
		JOIN catalog.categories c ON c.category_no = a.category_no
		LEFT JOIN catalog.attribute_values av ON av.attribute_no = a.attribute_no
			AND av.status != 'D'
		WHERE c.category_id = $1
		AND a.status != 'D'
		AND show_in_filter = TRUE
		ORDER BY a.sort_order ASC, a.attribute_name, av.sort_order ASC
	`, categoryId)
	if err != nil {
		return nil, fmt.Errorf("attributeRepository.GetByCategoryWithValues: %w", err)
	}
	defer rows.Close()

	attrMap := map[string]*models.AttributeWithValues{}
	attrOrder := []string{}

	for rows.Next() {
		var (
			attrId, attrName, code, attrType     string
			valId, displayLabel, value, colorHex *string
		)
		if err := rows.Scan(&attrId, &attrName, &code, &attrType, &valId, &displayLabel, &value, &colorHex); err != nil {
			return nil, fmt.Errorf("attributeRepository.GetByCategoryWithValues: scan: %w", err)
		}

		if _, exists := attrMap[attrId]; !exists {
			attrMap[attrId] = &models.AttributeWithValues{
				AttributeId:   attrId,
				AttributeName: attrName,
				Code:          code,
				Type:          attrType,
				Values:        []models.AttributeValueOption{},
			}
			attrOrder = append(attrOrder, attrId)
		}

		if valId != nil && displayLabel != nil && value != nil {
			hex := ""
			if colorHex != nil {
				hex = *colorHex
			}
			attrMap[attrId].Values = append(attrMap[attrId].Values, models.AttributeValueOption{
				AttributeValueId: *valId,
				DisplayLabel:     *displayLabel,
				Value:            *value,
				ColorHex:         hex,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("attributeRepository.GetByCategoryWithValues: rows: %w", err)
	}

	result := make([]models.AttributeWithValues, 0, len(attrOrder))
	for _, id := range attrOrder {
		result = append(result, *attrMap[id])
	}
	return result, nil
}
