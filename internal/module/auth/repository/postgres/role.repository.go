package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/repository"
)

type roleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) repository.RoleRepository {
	return &roleRepository{db: db}
}

// scanRole handles JSONB → []RolePermission conversion consistently
func scanRole(row interface {
	Scan(dest ...any) error
}) (*models.Role, error) {
	var d models.Role
	var roleDetailsRaw []byte

	if err := row.Scan(
		&d.RoleNo,
		&d.RoleId,
		&d.RoleName,
		&roleDetailsRaw,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(roleDetailsRaw, &d.RoleDetails); err != nil {
		return nil, fmt.Errorf("scanRole: unmarshal role_details: %w", err)
	}

	return &d, nil
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *roleRepository) GetAll(ctx context.Context) ([]models.Role, error) {
	query := `
		SELECT
			role_no,
			role_id,
			role_name,
			role_details,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM auth.roles
		WHERE status != 'D'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("roleRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.Role
	for rows.Next() {
		var d models.Role
		var roleDetailsRaw []byte

		if err := rows.Scan(
			&d.RoleNo,
			&d.RoleId,
			&d.RoleName,
			&roleDetailsRaw,
			&d.Status,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("roleRepository.GetAll: scan: %w", err)
		}

		if err := json.Unmarshal(roleDetailsRaw, &d.RoleDetails); err != nil {
			return nil, fmt.Errorf("roleRepository.GetAll: unmarshal role_details: %w", err)
		}

		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("roleRepository.GetAll: rows: %w", err)
	}

	return data, nil
}

func (r *roleRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.Role, error) {
	query := `
		SELECT
			role_no,
			role_id,
			role_name,
			role_details,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM auth.roles
		WHERE role_id = $1
		AND status != 'D'
	`

	var d models.Role
	var roleDetailsRaw []byte

	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.RoleNo,
		&d.RoleId,
		&d.RoleName,
		&roleDetailsRaw,
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
		return nil, fmt.Errorf("roleRepository.GetById: %w", err)
	}

	if err := json.Unmarshal(roleDetailsRaw, &d.RoleDetails); err != nil {
		return nil, fmt.Errorf("roleRepository.GetById: unmarshal role_details: %w", err)
	}

	return &d, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

func (r *roleRepository) Save(ctx context.Context, role *models.Role) (*models.Role, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("roleRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	roleDetailsJSON, err := json.Marshal(role.RoleDetails)
	if err != nil {
		return nil, fmt.Errorf("roleRepository.Save: marshal role_details: %w", err)
	}

	query := `
		INSERT INTO auth.roles (
			role_id,
			role_name,
			role_details,
			status,
			created_by
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING
			role_no,
			role_id,
			role_name,
			role_details,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Role
	var roleDetailsRaw []byte

	err = tx.QueryRow(ctx, query,
		role.RoleId,
		role.RoleName,
		roleDetailsJSON,
		models.StatusActive,
		role.CreatedBy,
	).Scan(
		&d.RoleNo,
		&d.RoleId,
		&d.RoleName,
		&roleDetailsRaw,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("roleRepository.Create: insert: %w", err)
	}

	if err = json.Unmarshal(roleDetailsRaw, &d.RoleDetails); err != nil {
		return nil, fmt.Errorf("roleRepository.Save: unmarshal role_details: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("roleRepository.Save: commit: %w", err)
	}

	return &d, nil
}

func (r *roleRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateRoleRequest) (*models.Role, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("roleRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock row — prevents lost updates
	var lockedNo uint8
	err = tx.QueryRow(ctx, `
		SELECT role_no FROM auth.roles
		WHERE role_id = $1 AND status != 'D'
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("roleRepository.Modify: lock row: %w", err)
	}

	roleDetailsJSON, err := json.Marshal(req.RoleDetails)
	if err != nil {
		return nil, fmt.Errorf("roleRepository.Modify: marshal role_details: %w", err)
	}

	query := `
		UPDATE auth.roles
		SET
			role_name    = COALESCE(NULLIF($1, ''), role_name),
			role_details = CASE WHEN $2::text = '[]' THEN role_details ELSE $2 END,
			updated_at   = NOW(),
			updated_by   = $3
		WHERE role_id = $4
		RETURNING
			role_no,
			role_id,
			role_name,
			role_details,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Role
	var roleDetailsRaw []byte

	err = tx.QueryRow(ctx, query,
		req.RoleName,
		roleDetailsJSON,
		req.UpdatedBy,
		Id,
	).Scan(
		&d.RoleNo,
		&d.RoleId,
		&d.RoleName,
		&roleDetailsRaw,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("roleRepository.Modify: update: %w", err)
	}

	if err = json.Unmarshal(roleDetailsRaw, &d.RoleDetails); err != nil {
		return nil, fmt.Errorf("roleRepository.Modify: unmarshal role_details: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("roleRepository.Modify: commit: %w", err)
	}

	return &d, nil
}

func (r *roleRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("roleRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint8
	err = tx.QueryRow(ctx, `
		SELECT role_no FROM auth.roles
		WHERE role_id = $1 AND status != 'D'
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("roleRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE auth.roles
		SET
			status     = 'D',
			updated_at = NOW()
		WHERE role_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("roleRepository.Remove: soft Remove: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("roleRepository.Remove: commit: %w", err)
	}

	return nil
}