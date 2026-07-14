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

// ── READ operations ───────────────────────────────────────────────────────────

func (r *attributeRepository) GetAll(ctx context.Context) ([]models.Attribute, error) {
	query := `
		SELECT
			attribute_no,
			attribute_id,
			category_no,
			attribute_name,
			code,
			type,
			is_required,
			is_variant,
			sort_order,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM product.attributes
		WHERE status != 'D'
		ORDER BY sort_order ASC, created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("attributeRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.Attribute
	for rows.Next() {
		var d models.Attribute

		if err := rows.Scan(
			&d.AttributeNo,
			&d.AttributeId,
			&d.CategoryNo,
			&d.AttributeName,
			&d.Code,
			&d.Type,
			&d.IsRequired,
			&d.IsVariant,
			&d.SortOrder,
			&d.Status,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
		); err != nil {
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
	query := `
		SELECT
			attribute_no,
			attribute_id,
			category_no,
			attribute_name,
			code,
			type,
			is_required,
			is_variant,
			sort_order,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM product.attributes
		WHERE attribute_id = $1
		AND status != 'D'
	`

	var d models.Attribute

	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.AttributeNo,
		&d.AttributeId,
		&d.CategoryNo,
		&d.AttributeName,
		&d.Code,
		&d.Type,
		&d.IsRequired,
		&d.IsVariant,
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
		return nil, fmt.Errorf("attributeRepository.GetById: %w", err)
	}

	return &d, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

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

	query := `
		INSERT INTO product.attributes (
			attribute_id,
			category_no,
			attribute_name,
			code,
			type,
			is_required,
			is_variant,
			sort_order,
			created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			attribute_no,
			attribute_id,
			category_no,
			attribute_name,
			code,
			type,
			is_required,
			is_variant,
			sort_order,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Attribute

	err = tx.QueryRow(ctx, query,
		attribute.AttributeId,
		attribute.CategoryNo,
		attribute.AttributeName,
		attribute.Code,
		attribute.Type,
		attribute.IsRequired,
		attribute.IsVariant,
		attribute.SortOrder,
		attribute.CreatedBy,
	).Scan(
		&d.AttributeNo,
		&d.AttributeId,
		&d.CategoryNo,
		&d.AttributeName,
		&d.Code,
		&d.Type,
		&d.IsRequired,
		&d.IsVariant,
		&d.SortOrder,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
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
		SELECT attribute_no FROM product.attributes
		WHERE attribute_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("attributeRepository.Modify: lock row: %w", err)
	}

	query := `
		UPDATE product.attributes
		SET
			category_no    = $1,
			attribute_name = COALESCE(NULLIF($2, ''), attribute_name),
			code           = COALESCE(NULLIF($3, ''), code),
			type           = $4,
			is_required    = $5,
			is_variant     = $6,
			sort_order     = $7,
			updated_at     = NOW(),
			updated_by     = $8
		WHERE attribute_id = $9
		RETURNING
			attribute_no,
			attribute_id,
			category_no,
			attribute_name,
			code,
			type,
			is_required,
			is_variant,
			sort_order,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Attribute

	err = tx.QueryRow(ctx, query,
		req.CategoryNo,
		req.AttributeName,
		req.Code,
		req.Type,
		req.IsRequired,
		req.IsVariant,
		req.SortOrder,
		req.UpdatedBy,
		Id,
	).Scan(
		&d.AttributeNo,
		&d.AttributeId,
		&d.CategoryNo,
		&d.AttributeName,
		&d.Code,
		&d.Type,
		&d.IsRequired,
		&d.IsVariant,
		&d.SortOrder,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
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
		SELECT attribute_no FROM product.attributes
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
		UPDATE product.attributes
		SET
			status     = 'D',
			updated_at = NOW()
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
