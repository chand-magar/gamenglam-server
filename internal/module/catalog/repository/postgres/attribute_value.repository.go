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
	rows, err := r.db.Query(ctx, `
		SELECT
			av.attribute_value_no,
			av.attribute_value_id,
			av.attribute_no,
			a.attribute_id,
			av.value,
			av.display_label,
			av.color_hex,
			av.sort_order,
			av.created_at,
			av.created_by,
			av.updated_at,
			av.updated_by
		FROM catalog.attribute_values av
		JOIN catalog.attributes a ON a.attribute_no = av.attribute_no
		WHERE av.status != 'D'
		ORDER BY av.sort_order ASC, av.created_at DESC
	`)
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
			&d.AttributeId, // ← add
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
	var d models.AttributeValue
	err := r.db.QueryRow(ctx, `
		SELECT
			av.attribute_value_no,
			av.attribute_value_id,
			av.attribute_no,
			a.attribute_id,
			av.value,
			av.display_label,
			av.color_hex,
			av.sort_order,
			av.created_at,
			av.created_by,
			av.updated_at,
			av.updated_by
		FROM catalog.attribute_values av
		JOIN catalog.attributes a ON a.attribute_no = av.attribute_no
		WHERE av.attribute_value_id = $1
		AND av.status != 'D'
	`, Id).Scan(
		&d.AttributeValueNo,
		&d.AttributeValueId,
		&d.AttributeNo,
		&d.AttributeId, // ← add
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

	// Resolve attribute_no from attribute_id
	var attributeNo uint32
	err = tx.QueryRow(ctx, `
		SELECT attribute_no FROM catalog.attributes
		WHERE attribute_id = $1 AND status != 'D'
	`, attributeValue.AttributeId).Scan(&attributeNo)
	if err != nil {
		return nil, fmt.Errorf("attributeValueRepository.Save: resolve attribute: %w", err)
	}

	var d models.AttributeValue
	err = tx.QueryRow(ctx, `
		INSERT INTO catalog.attribute_values (
			attribute_value_id, attribute_no, value,
			display_label, color_hex, sort_order, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING
			attribute_value_no,
			attribute_value_id,
			attribute_no,
			(SELECT attribute_id FROM catalog.attributes WHERE attribute_no = $2),
			value, display_label, color_hex, sort_order,
			created_at, created_by, updated_at, updated_by
	`,
		attributeValue.AttributeValueId,
		attributeNo,
		attributeValue.Value,
		attributeValue.DisplayLabel,
		attributeValue.ColorHex,
		attributeValue.SortOrder,
		attributeValue.CreatedBy,
	).Scan(
		&d.AttributeValueNo,
		&d.AttributeValueId,
		&d.AttributeNo,
		&d.AttributeId, // ← add
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
		SELECT attribute_value_no FROM catalog.attribute_values
		WHERE attribute_value_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("attributeValueRepository.Modify: lock row: %w", err)
	}

	var d models.AttributeValue
	err = tx.QueryRow(ctx, `
		UPDATE catalog.attribute_values SET
			value         = COALESCE(NULLIF($1, ''), value),
			display_label = COALESCE(NULLIF($2, ''), display_label),
			color_hex     = $3,
			sort_order    = $4,
			updated_at    = NOW(),
			updated_by    = $5
		WHERE attribute_value_id = $6
		RETURNING
			attribute_value_no,
			attribute_value_id,
			attribute_no,
			(SELECT a2.attribute_id FROM catalog.attributes a2 WHERE a2.attribute_no = (
				SELECT attribute_no FROM catalog.attribute_values WHERE attribute_value_id = $6
			)),
			value, display_label, color_hex, sort_order,
			created_at, created_by, updated_at, updated_by
	`,
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
		&d.AttributeId, // ← add
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
		SELECT attribute_value_no FROM catalog.attribute_values
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
		UPDATE catalog.attribute_values
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

func (r *attributeValueRepository) GetByAttributeId(ctx context.Context, attributeId uuid.UUID) ([]models.AttributeValue, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			av.attribute_value_no,
			av.attribute_value_id,
			av.attribute_no,
			a.attribute_id,
			av.value,
			av.display_label,
			av.color_hex,
			av.sort_order,
			av.created_at,
			av.created_by,
			av.updated_at,
			av.updated_by
		FROM catalog.attribute_values av
		JOIN catalog.attributes a ON a.attribute_no = av.attribute_no
		WHERE a.attribute_id = $1
		AND av.status != 'D'
		ORDER BY av.sort_order ASC, av.created_at DESC
	`, attributeId)
	if err != nil {
		return nil, fmt.Errorf("attributeValueRepository.GetByAttributeId: %w", err)
	}
	defer rows.Close()

	var data []models.AttributeValue
	for rows.Next() {
		var d models.AttributeValue
		if err := rows.Scan(
			&d.AttributeValueNo,
			&d.AttributeValueId,
			&d.AttributeNo,
			&d.AttributeId, // ← add
			&d.Value,
			&d.DisplayLabel,
			&d.ColorHex,
			&d.SortOrder,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("attributeValueRepository.GetByAttributeId: scan: %w", err)
		}
		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("attributeValueRepository.GetByAttributeId: rows: %w", err)
	}

	return data, nil
}
