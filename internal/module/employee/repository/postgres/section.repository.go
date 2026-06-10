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

type sectionRepository struct {
	db *pgxpool.Pool
}

func NewSectionRepository(db *pgxpool.Pool) repository.SectionRepository {
	return &sectionRepository{db: db}
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *sectionRepository) GetAll(ctx context.Context) ([]models.Section, error) {
	query := `
		SELECT
			section_no,
			section_id,
			section_name,
			section_path,
			section_icon,
			section_order,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM auth.sections
		WHERE status != 'D'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("sectionRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.Section
	for rows.Next() {
		var d models.Section

		if err := rows.Scan(
			&d.SectionNo,
			&d.SectionId,
			&d.SectionName,
			&d.SectionPath,
			&d.SectionIcon,
			&d.SectionOrder,
			&d.Status,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("sectionRepository.GetAll: scan: %w", err)
		}

		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sectionRepository.GetAll: rows: %w", err)
	}

	return data, nil
}

func (r *sectionRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.Section, error) {
	query := `
		SELECT
			section_no,
			section_id,
			section_name,
			section_path,
			section_icon,
			section_order,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM auth.sections
		WHERE section_id = $1
		AND status != 'D'
	`

	var d models.Section

	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.SectionNo,
		&d.SectionId,
		&d.SectionName,
		&d.SectionPath,
		&d.SectionIcon,
		&d.SectionOrder,
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
		return nil, fmt.Errorf("sectionRepository.GetById: %w", err)
	}

	return &d, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

func (r *sectionRepository) Save(ctx context.Context, section *models.Section) (*models.Section, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("sectionRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO auth.sections (
			section_id,
			section_name,
			section_path,
			status,
			created_by
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING
			section_no,
			section_id,
			section_name,
			section_path,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Section

	err = tx.QueryRow(ctx, query,
		section.SectionId,
		section.SectionName,
		section.SectionPath,
		section.SectionIcon,
		section.SectionOrder,
		models.StatusActive,
		section.CreatedBy,
	).Scan(
		&d.SectionNo,
		&d.SectionId,
		&d.SectionName,
		&d.SectionPath,
		&d.SectionIcon,
		&d.SectionOrder,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("sectionRepository.Save: Save: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("sectionRepository.Save: commit: %w", err)
	}

	return &d, nil
}

func (r *sectionRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateSectionRequest) (*models.Section, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("sectionRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock row — prevents lost updates
	var lockedNo uint8
	err = tx.QueryRow(ctx, `
		SELECT section_no FROM auth.sections
		WHERE section_id = $1 AND status != 'D'
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("sectionRepository.Update: lock row: %w", err)
	}

	query := `
		UPDATE auth.sections
		SET
			section_name    = COALESCE(NULLIF($1, ''), section_name),
			section_path 	= COALESCE(NULLIF($2, ''), section_path),
			section_icon 	= COALESCE(NULLIF($3, ''), section_icon),
			section_order 	= COALESCE(NULLIF($4, ''), section_order),
			updated_at   	= NOW(),
			updated_by   	= $5
		WHERE section_id 	= $6
		RETURNING
			section_no,
			section_id,
			section_name,
			section_path,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Section

	err = tx.QueryRow(ctx, query,
		req.SectionName,
		req.SectionPath,
		req.SectionIcon,
		req.SectionOrder,
		req.UpdatedBy,
		Id,
	).Scan(
		&d.SectionNo,
		&d.SectionId,
		&d.SectionName,
		&d.SectionPath,
		&d.SectionIcon,
		&d.SectionOrder,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("sectionRepository.Modify: Modify: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("sectionRepository.Modify: commit: %w", err)
	}

	return &d, nil
}

func (r *sectionRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("sectionRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint8
	err = tx.QueryRow(ctx, `
		SELECT section_no FROM auth.sections
		WHERE section_id = $1 AND status != 'D'
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("sectionRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE auth.sections
		SET
			status     = 'D',
			updated_at = NOW()
		WHERE section_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("sectionRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("sectionRepository.Remove: commit: %w", err)
	}

	return nil
}
