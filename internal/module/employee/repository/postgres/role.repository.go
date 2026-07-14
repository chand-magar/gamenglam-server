package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/employee/models"
	"github.com/chand-magar/gamenglam-server/internal/module/employee/repository"
)

type roleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) repository.RoleRepository {
	return &roleRepository{db: db}
}

// scanRole handles JSONB → []RoleSectionPermission consistently
func scanRole(row interface{ Scan(dest ...any) error }) (*models.Role, error) {
	var ro models.Role
	var roleDetailsRaw []byte

	if err := row.Scan(
		&ro.RoleNo,
		&ro.RoleId,
		&ro.RoleName,
		&roleDetailsRaw,
		&ro.Status,
		&ro.CreatedAt,
		&ro.CreatedBy,
		&ro.UpdatedAt,
		&ro.UpdatedBy,
	); err != nil {
		return nil, err
	}

	if len(roleDetailsRaw) > 0 {
		if err := json.Unmarshal(roleDetailsRaw, &ro.RoleDetails); err != nil {
			return nil, fmt.Errorf("scanRole: unmarshal role_details: %w", err)
		}
	}

	return &ro, nil
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *roleRepository) GetAll(ctx context.Context) ([]models.Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			role_no, role_id, role_name, role_details,
			status, created_at, created_by, updated_at, updated_by
		FROM auth.roles
		WHERE status != 'D'
		ORDER BY role_no ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("roleRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var ro models.Role
		var roleDetailsRaw []byte

		if err := rows.Scan(
			&ro.RoleNo, &ro.RoleId, &ro.RoleName, &roleDetailsRaw,
			&ro.Status, &ro.CreatedAt, &ro.CreatedBy, &ro.UpdatedAt, &ro.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("roleRepository.GetAll: scan: %w", err)
		}

		if len(roleDetailsRaw) > 0 {
			if err := json.Unmarshal(roleDetailsRaw, &ro.RoleDetails); err != nil {
				return nil, fmt.Errorf("roleRepository.GetAll: unmarshal: %w", err)
			}
		}

		roles = append(roles, ro)
	}

	return roles, rows.Err()
}

func (r *roleRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.Role, error) {
	row := r.db.QueryRow(ctx, `
		SELECT
			role_no, role_id, role_name, role_details,
			status, created_at, created_by, updated_at, updated_by
		FROM auth.roles
		WHERE role_id = $1 AND status != 'D'
	`, Id)

	ro, err := scanRole(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("roleRepository.GetById: %w", err)
	}

	return ro, nil
}

// ── WRITE operations ──────────────────────────────────────────────────────────

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

	var roleDetailsRaw []byte
	var ro models.Role

	err = tx.QueryRow(ctx, `
		INSERT INTO auth.roles (
			role_id, role_name, role_details, status, created_by
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING
			role_no, role_id, role_name, role_details,
			status, created_at, created_by, updated_at, updated_by
	`,
		role.RoleId,
		role.RoleName,
		roleDetailsJSON,
		models.StatusActive,
		role.CreatedBy,
	).Scan(
		&ro.RoleNo, &ro.RoleId, &ro.RoleName, &roleDetailsRaw,
		&ro.Status, &ro.CreatedAt, &ro.CreatedBy, &ro.UpdatedAt, &ro.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("roleRepository.Save: insert: %w", err)
	}

	if len(roleDetailsRaw) > 0 {
		if err = json.Unmarshal(roleDetailsRaw, &ro.RoleDetails); err != nil {
			return nil, fmt.Errorf("roleRepository.Save: unmarshal: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("roleRepository.Save: commit: %w", err)
	}

	return &ro, nil
}

func (r *roleRepository) Modify(ctx context.Context, id uuid.UUID, req *models.UpdateRoleRequest) (*models.Role, error) {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("roleRepository.Modify: begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock row
	var roleNo uint8
	err = tx.QueryRow(ctx, `
		SELECT role_no
		FROM auth.roles
		WHERE role_id = $1
		  AND status != 'D'
		FOR UPDATE
	`, id).Scan(&roleNo)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("roleRepository.Modify: lock row: %w", err)
	}

	// Prepare role_details
	var roleDetails any

	if len(req.RoleDetails) > 0 {
		b, err := json.Marshal(req.RoleDetails)
		if err != nil {
			return nil, fmt.Errorf("roleRepository.Modify: marshal role_details: %w", err)
		}

		// Pass string instead of []byte
		roleDetails = string(b)
	} else {
		roleDetails = nil
	}

	var (
		role           models.Role
		roleDetailsRaw []byte
	)

	err = tx.QueryRow(ctx, `
		UPDATE auth.roles
		SET
			role_name = COALESCE(NULLIF($1, ''), role_name),
			role_details = COALESCE($2::jsonb, role_details),
			updated_at = NOW(),
			updated_by = $3
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
	`,
		req.RoleName,
		roleDetails,
		req.UpdatedBy,
		id,
	).Scan(
		&role.RoleNo,
		&role.RoleId,
		&role.RoleName,
		&roleDetailsRaw,
		&role.Status,
		&role.CreatedAt,
		&role.CreatedBy,
		&role.UpdatedAt,
		&role.UpdatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("roleRepository.Modify: update: %w", err)
	}

	if len(roleDetailsRaw) > 0 {
		if err := json.Unmarshal(roleDetailsRaw, &role.RoleDetails); err != nil {
			return nil, fmt.Errorf("roleRepository.Modify: unmarshal role_details: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("roleRepository.Modify: commit: %w", err)
	}

	return &role, nil
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
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("roleRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE auth.roles
		SET status = 'D', updated_at = NOW()
		WHERE role_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("roleRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("roleRepository.Remove: commit: %w", err)
	}

	return nil
}
