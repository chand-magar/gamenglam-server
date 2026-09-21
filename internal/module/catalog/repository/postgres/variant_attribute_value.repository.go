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

type variantAttributeValueRepository struct {
	db *pgxpool.Pool
}

func NewVariantAttributeValueRepository(db *pgxpool.Pool) repository.VariantAttributeValueRepository {
	return &variantAttributeValueRepository{db: db}
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *variantAttributeValueRepository) GetAll(ctx context.Context) ([]models.VariantAttributeValue, error) {
	query := `
		SELECT
			variant_attribute_value_no,
			variant_attribute_value_id,
			variant_no,
			attribute_value_no,
			created_at,
			created_by
		FROM catalog.variant_attribute_values
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("variantAttributeValueRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.VariantAttributeValue
	for rows.Next() {
		var d models.VariantAttributeValue

		if err := rows.Scan(
			&d.VariantAttributeValueNo,
			&d.VariantAttributeValueId,
			&d.VariantNo,
			&d.AttributeValueNo,
			&d.CreatedAt,
			&d.CreatedBy,
		); err != nil {
			return nil, fmt.Errorf("variantAttributeValueRepository.GetAll: scan: %w", err)
		}

		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("variantAttributeValueRepository.GetAll: rows: %w", err)
	}

	return data, nil
}

func (r *variantAttributeValueRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.VariantAttributeValue, error) {
	query := `
		SELECT
			variant_attribute_value_no,
			variant_attribute_value_id,
			variant_no,
			attribute_value_no,
			created_at,
			created_by
		FROM catalog.variant_attribute_values
		WHERE variant_attribute_value_id = $1
	`

	var d models.VariantAttributeValue

	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.VariantAttributeValueNo,
		&d.VariantAttributeValueId,
		&d.VariantNo,
		&d.AttributeValueNo,
		&d.CreatedAt,
		&d.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("variantAttributeValueRepository.GetById: %w", err)
	}

	return &d, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

func (r *variantAttributeValueRepository) Save(ctx context.Context, vav *models.VariantAttributeValue) (*models.VariantAttributeValue, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("variantAttributeValueRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO catalog.variant_attribute_values (
			variant_attribute_value_id,
			variant_no,
			attribute_value_no,
			created_by
		) VALUES ($1, $2, $3, $4)
		RETURNING
			variant_attribute_value_no,
			variant_attribute_value_id,
			variant_no,
			attribute_value_no,
			created_at,
			created_by
	`

	var d models.VariantAttributeValue

	err = tx.QueryRow(ctx, query,
		vav.VariantAttributeValueId,
		vav.VariantNo,
		vav.AttributeValueNo,
		vav.CreatedBy,
	).Scan(
		&d.VariantAttributeValueNo,
		&d.VariantAttributeValueId,
		&d.VariantNo,
		&d.AttributeValueNo,
		&d.CreatedAt,
		&d.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("variantAttributeValueRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("variantAttributeValueRepository.Save: commit: %w", err)
	}

	return &d, nil
}

func (r *variantAttributeValueRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("variantAttributeValueRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT variant_attribute_value_no FROM catalog.variant_attribute_values
		WHERE variant_attribute_value_id = $1
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("variantAttributeValueRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM catalog.variant_attribute_values
		WHERE variant_attribute_value_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("variantAttributeValueRepository.Remove: hard delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("variantAttributeValueRepository.Remove: commit: %w", err)
	}

	return nil
}

func (r *variantAttributeValueRepository) GetByVariantId(ctx context.Context, variantId uuid.UUID) ([]models.VariantAttributeValue, error) {
	query := `
		SELECT
			vav.variant_attribute_value_no,
			vav.variant_attribute_value_id,
			vav.variant_no,
			vav.attribute_value_no,
			vav.created_at,
			vav.created_by
		FROM catalog.variant_attribute_values vav
		JOIN catalog.product_variants pv ON pv.variant_no = vav.variant_no
		WHERE pv.variant_id = $1
		ORDER BY vav.attribute_value_no ASC
	`

	rows, err := r.db.Query(ctx, query, variantId)
	if err != nil {
		return nil, fmt.Errorf("variantAttributeValueRepository.GetByVariantId: %w", err)
	}
	defer rows.Close()

	var data []models.VariantAttributeValue
	for rows.Next() {
		var d models.VariantAttributeValue
		if err := rows.Scan(
			&d.VariantAttributeValueNo,
			&d.VariantAttributeValueId,
			&d.VariantNo,
			&d.AttributeValueNo,
			&d.CreatedAt,
			&d.CreatedBy,
		); err != nil {
			return nil, fmt.Errorf("variantAttributeValueRepository.GetByVariantId: scan: %w", err)
		}
		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("variantAttributeValueRepository.GetByVariantId: rows: %w", err)
	}

	return data, nil
}

func (r *variantAttributeValueRepository) GetAttributeValueNoById(ctx context.Context, attributeValueId uuid.UUID) (uint32, error) {
	var valueNo uint32
	err := r.db.QueryRow(ctx, `
		SELECT attribute_value_no FROM catalog.attribute_values
		WHERE attribute_value_id = $1 AND status != 'D'
	`, attributeValueId).Scan(&valueNo)
	if err != nil {
		return 0, fmt.Errorf("variantAttributeValueRepository.GetAttributeValueNoById: %w", err)
	}
	return valueNo, nil
}
