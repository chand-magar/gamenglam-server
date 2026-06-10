package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/employee/models"
	"github.com/chand-magar/gamenglam-server/internal/module/employee/repository"
)

type pageRepository struct {
	db *pgxpool.Pool
}

func NewPageRepository(db *pgxpool.Pool) repository.PageRepository {
	return &pageRepository{db: db}
}

func (r *pageRepository) GetAll(ctx context.Context) ([]models.Page, error) {
	query := `
		SELECT
			page_no, page_id, section_no, page_name,
			page_path, page_order, status,
			created_at, created_by, updated_at, updated_by
		FROM auth.pages
		WHERE status != 'D'
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("pageRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.Page
	for rows.Next() {
		var d models.Page
		if err := rows.Scan(
			&d.PageNo, &d.PageId, &d.SectionNo, &d.PageName,
			&d.PagePath, &d.PageOrder, &d.Status,
			&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("pageRepository.GetAll: scan: %w", err)
		}
		data = append(data, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pageRepository.GetAll: rows: %w", err)
	}
	return data, nil
}

func (r *pageRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.Page, error) {
	query := `
		SELECT
			page_no, page_id, section_no, page_name,
			page_path, page_order, status,
			created_at, created_by, updated_at, updated_by
		FROM auth.pages
		WHERE page_id = $1 AND status != 'D'
	`
	var d models.Page
	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.PageNo, &d.PageId, &d.SectionNo, &d.PageName,
		&d.PagePath, &d.PageOrder, &d.Status,
		&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pageRepository.GetById: %w", err)
	}
	return &d, nil
}

func (r *pageRepository) Save(ctx context.Context, page *models.Page) (*models.Page, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("pageRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO auth.pages (
			page_id, section_no, page_name, page_path, page_order, status, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			page_no, page_id, section_no, page_name,
			page_path, page_order, status,
			created_at, created_by, updated_at, updated_by
	`
	var d models.Page
	err = tx.QueryRow(ctx, query,
		page.PageId,
		page.SectionNo,
		page.PageName,
		page.PagePath,
		page.PageOrder,
		models.StatusActive,
		page.CreatedBy,
	).Scan(
		&d.PageNo, &d.PageId, &d.SectionNo, &d.PageName,
		&d.PagePath, &d.PageOrder, &d.Status,
		&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("pageRepository.Save: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("pageRepository.Save: commit: %w", err)
	}
	return &d, nil
}

func (r *pageRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdatePageRequest) (*models.Page, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("pageRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint8
	err = tx.QueryRow(ctx,
		`SELECT page_no FROM auth.pages WHERE page_id = $1 AND status != 'D'`,
		Id,
	).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pageRepository.Modify: lock row: %w", err)
	}

	query := `
		UPDATE auth.pages SET
			page_name  = COALESCE(NULLIF($1, ''), page_name),
			page_path  = COALESCE(NULLIF($2, ''), page_path),
			page_order = COALESCE($3, page_order),
			updated_at = NOW(),
			updated_by = $4
		WHERE page_id = $5
		RETURNING
			page_no, page_id, section_no, page_name,
			page_path, page_order, status,
			created_at, created_by, updated_at, updated_by
	`
	var d models.Page
	err = tx.QueryRow(ctx, query,
		req.PageName,
		req.PagePath,
		req.PageOrder,
		req.UpdatedBy,
		Id,
	).Scan(
		&d.PageNo, &d.PageId, &d.SectionNo, &d.PageName,
		&d.PagePath, &d.PageOrder, &d.Status,
		&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("pageRepository.Modify: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("pageRepository.Modify: commit: %w", err)
	}
	return &d, nil
}

func (r *pageRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pageRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint8
	err = tx.QueryRow(ctx,
		`SELECT page_no FROM auth.pages WHERE page_id = $1 AND status != 'D'`,
		Id,
	).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("pageRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE auth.pages SET status = 'D', updated_at = NOW()
		WHERE page_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("pageRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("pageRepository.Remove: commit: %w", err)
	}
	return nil
}
