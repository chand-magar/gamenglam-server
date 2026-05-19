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

type categoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) repository.CategoryRepository {
	return &categoryRepository{db: db}
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *categoryRepository) GetAll(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT
			category_no,
			category_id,
			parent_no,
			category_name,
			slug,
			details,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM product.categories
		WHERE status != 'D'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("categoryRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.Category
	for rows.Next() {
		var d models.Category

		if err := rows.Scan(
			&d.CategoryNo,
			&d.CategoryId,
			&d.ParentNo,
			&d.CategoryName,
			&d.Slug,
			&d.Details,
			&d.Status,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("categoryRepository.GetAll: scan: %w", err)
		}

		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("categoryRepository.GetAll: rows: %w", err)
	}

	return data, nil
}

func (r *categoryRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.Category, error) {
	query := `
		SELECT
			category_no,
			category_id,
			parent_no,
			category_name,
			slug,
			details,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM product.categories
		WHERE category_id = $1
		AND status != 'D'
	`

	var d models.Category

	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.CategoryNo,
		&d.CategoryId,
		&d.ParentNo,
		&d.CategoryName,
		&d.Slug,
		&d.Details,
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
		return nil, fmt.Errorf("categoryRepository.GetById: %w", err)
	}

	return &d, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

func (r *categoryRepository) Save(ctx context.Context, category *models.Category) (*models.Category, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("categoryRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO product.categories (
			category_id,
			parent_no,
			category_name,
			slug,
			details,
			status,
			created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			category_no,
			category_id,
			parent_no,
			category_name,
			slug,
			details,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Category

	err = tx.QueryRow(ctx, query,
		category.CategoryId,
		category.ParentNo,
		category.CategoryName,
		category.Slug,
		category.Details,
		models.StatusActive,
		category.CreatedBy,
	).Scan(
		&d.CategoryNo,
		&d.CategoryId,
		&d.ParentNo,
		&d.CategoryName,
		&d.Slug,
		&d.Details,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("categoryRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("categoryRepository.Save: commit: %w", err)
	}

	return &d, nil
}

func (r *categoryRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateCategoryRequest) (*models.Category, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("categoryRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT category_no FROM product.categories
		WHERE category_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("categoryRepository.Modify: lock row: %w", err)
	}

	query := `
		UPDATE product.categories
		SET
			parent_no     = $1,
			category_name = COALESCE(NULLIF($2, ''), category_name),
			slug          = COALESCE(NULLIF($3, ''), slug),
			details       = COALESCE(NULLIF($4, ''), details),
			updated_at    = NOW(),
			updated_by    = $5
		WHERE category_id = $6
		RETURNING
			category_no,
			category_id,
			parent_no,
			category_name,
			slug,
			details,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Category

	err = tx.QueryRow(ctx, query,
		req.ParentNo,
		req.CategoryName,
		req.Slug,
		req.Details,
		req.UpdatedBy,
		Id,
	).Scan(
		&d.CategoryNo,
		&d.CategoryId,
		&d.ParentNo,
		&d.CategoryName,
		&d.Slug,
		&d.Details,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("categoryRepository.Modify: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("categoryRepository.Modify: commit: %w", err)
	}

	return &d, nil
}

func (r *categoryRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("categoryRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT category_no FROM product.categories
		WHERE category_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("categoryRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE product.categories
		SET
			status     = 'D',
			updated_at = NOW()
		WHERE category_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("categoryRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("categoryRepository.Remove: commit: %w", err)
	}

	return nil
}