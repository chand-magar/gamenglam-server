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

type attributeValueRepository struct {
	db *pgxpool.Pool
}

func NewAttributeValueRepository(db *pgxpool.Pool) repository.AttributeValueRepository {
	return &attributeValueRepository{db: db}
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *attributeValueRepository) GetAll(ctx context.Context) ([]models.AttributeValue, error) {
	query := `
		SELECT
			attribute_value_no,
			attribute_value_id,
			attribute_no,
			value,
			display_label,
			color_hex,
			sort_order,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM product.attribute_values
		WHERE status != 'D'
		ORDER BY sort_order ASC, created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("attributeValueRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.AttributeValue
	for rows.Next() {
		var d models.AttributeValue

		if err := rows.Scan(
			&d.AttributeValueNo,
			&d.AttributeValueId,
			&d.AttributeNo,
			&d.Value,
			&d.DisplayLabel,
			&d.ColorHex,
			&d.SortOrder,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("attributeValueRepository.GetAll: scan: %w", err)
		}

		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("attributeValueRepository.GetAll: rows: %w", err)
	}

	return data, nil
}

func (r *attributeValueRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.AttributeValue, error) {
	query := `
		SELECT
			attribute_value_no,
			attribute_value_id,
			attribute_no,
			value,
			display_label,
			color_hex,
			sort_order,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM product.attribute_values
		WHERE attribute_value_id = $1
		AND status != 'D'
	`

	var d models.AttributeValue

	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.AttributeValueNo,
		&d.AttributeValueId,
		&d.AttributeNo,
		&d.Value,
		&d.DisplayLabel,
		&d.ColorHex,
		&d.SortOrder,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("attributeValueRepository.GetById: %w", err)
	}

	return &d, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

func (r *attributeValueRepository) Save(ctx context.Context, attributeValue *models.AttributeValue) (*models.AttributeValue, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("attributeValueRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO product.attribute_values (
			attribute_value_id,
			attribute_no,
			value,
			display_label,
			color_hex,
			sort_order,
			created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			attribute_value_no,
			attribute_value_id,
			attribute_no,
			value,
			display_label,
			color_hex,
			sort_order,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.AttributeValue

	err = tx.QueryRow(ctx, query,
		attributeValue.AttributeValueId,
		attributeValue.AttributeNo,
		attributeValue.Value,
		attributeValue.DisplayLabel,
		attributeValue.ColorHex,
		attributeValue.SortOrder,
		attributeValue.CreatedBy,
	).Scan(
		&d.AttributeValueNo,
		&d.AttributeValueId,
		&d.AttributeNo,
		&d.Value,
		&d.DisplayLabel,
		&d.ColorHex,
		&d.SortOrder,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("attributeValueRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("attributeValueRepository.Save: commit: %w", err)
	}

	return &d, nil
}

func (r *attributeValueRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateAttributeValueRequest) (*models.AttributeValue, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("attributeValueRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT attribute_value_no FROM product.attribute_values
		WHERE attribute_value_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("attributeValueRepository.Modify: lock row: %w", err)
	}

	query := `
		UPDATE product.attribute_values
		SET
			attribute_no  = $1,
			value         = COALESCE(NULLIF($2, ''), value),
			display_label = COALESCE(NULLIF($3, ''), display_label),
			color_hex     = $4,
			sort_order    = $5,
			updated_at    = NOW(),
			updated_by    = $6
		WHERE attribute_value_id = $7
		RETURNING
			attribute_value_no,
			attribute_value_id,
			attribute_no,
			value,
			display_label,
			color_hex,
			sort_order,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.AttributeValue

	err = tx.QueryRow(ctx, query,
		req.AttributeNo,
		req.Value,
		req.DisplayLabel,
		req.ColorHex,
		req.SortOrder,
		req.UpdatedBy,
		Id,
	).Scan(
		&d.AttributeValueNo,
		&d.AttributeValueId,
		&d.AttributeNo,
		&d.Value,
		&d.DisplayLabel,
		&d.ColorHex,
		&d.SortOrder,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("attributeValueRepository.Modify: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("attributeValueRepository.Modify: commit: %w", err)
	}

	return &d, nil
}

func (r *attributeValueRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("attributeValueRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT attribute_value_no FROM product.attribute_values
		WHERE attribute_value_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("attributeValueRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE product.attribute_values
		SET
			status     = 'D',
			updated_at = NOW()
		WHERE attribute_value_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("attributeValueRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("attributeValueRepository.Remove: commit: %w", err)
	}

	return nil
}