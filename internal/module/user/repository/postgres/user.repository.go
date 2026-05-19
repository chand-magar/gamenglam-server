package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/user/models"
	"github.com/chand-magar/gamenglam-server/internal/module/user/repository"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *userRepository) GetAll(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT
			profile_no,
			profile_id,
			role_no,
			user_fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM master.users
		WHERE status != 'D'
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("userRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.User
	for rows.Next() {
		var d models.User
		if err := rows.Scan(
			&d.ProfileNo,
			&d.ProfileId,
			&d.RoleNo,
			&d.UserFullName,
			&d.EmailId,
			&d.Gender,
			&d.Dob,
			&d.MobileNo,
			&d.Address,
			&d.Status,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("userRepository.GetAll: scan: %w", err)
		}
		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("userRepository.GetAll: rows: %w", err)
	}

	return data, nil
}

func (r *userRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.User, error) {
	query := `
		SELECT
			profile_no,
			profile_id,
			role_no,
			user_fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM master.users
		WHERE profile_id = $1
		AND status != 'D'
	`

	var d models.User
	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.ProfileNo,
		&d.ProfileId,
		&d.RoleNo,
		&d.UserFullName,
		&d.EmailId,
		&d.Gender,
		&d.Dob,
		&d.MobileNo,
		&d.Address,
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
		return nil, fmt.Errorf("userRepository.GetById: %w", err)
	}

	return &d, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT
			profile_no,
			profile_id,
			role_no,
			user_fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM master.users
		WHERE email_id = $1
		AND status != 'D'
	`

	var d models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&d.ProfileNo,
		&d.ProfileId,
		&d.RoleNo,
		&d.UserFullName,
		&d.EmailId,
		&d.Gender,
		&d.Dob,
		&d.MobileNo,
		&d.Address,
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
		return nil, fmt.Errorf("userRepository.GetByEmail: %w", err)
	}

	return &d, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

func (r *userRepository) Save(ctx context.Context, user *models.User) (*models.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("userRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO master.users (
			profile_id,
			role_no,
			user_fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING
			profile_no,
			profile_id,
			role_no,
			user_fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.User
	err = tx.QueryRow(ctx, query,
		user.ProfileId,
		user.RoleNo,
		user.UserFullName,
		user.EmailId,
		user.Gender,
		user.Dob,
		user.MobileNo,
		user.Address,
		models.StatusActive,
		user.CreatedBy,
	).Scan(
		&d.ProfileNo,
		&d.ProfileId,
		&d.RoleNo,
		&d.UserFullName,
		&d.EmailId,
		&d.Gender,
		&d.Dob,
		&d.MobileNo,
		&d.Address,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("userRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("userRepository.Save: commit: %w", err)
	}

	return &d, nil
}

func (r *userRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("userRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock row by profile_id — prevents lost updates
	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT profile_no FROM master.users
		WHERE profile_id = $1 AND status != 'D'
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.Modify: lock row: %w", err)
	}

	query := `
		UPDATE master.users
		SET
			user_fullname = COALESCE(NULLIF($1, ''), user_fullname),
			email_id      = COALESCE(NULLIF($2, ''), email_id),
			gender        = COALESCE(NULLIF($3, ''), gender),
			mobile_no     = COALESCE(NULLIF($4, ''), mobile_no),
			address       = CASE WHEN $5::text = '' THEN address ELSE $5::jsonb END,
			updated_at    = NOW(),
			updated_by    = $6
		WHERE profile_id  = $7
		RETURNING
			profile_no,
			profile_id,
			role_no,
			user_fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.User
	err = tx.QueryRow(ctx, query,
		req.UserFullName,
		req.EmailId,
		req.Gender,
		req.MobileNo,
		req.Address,
		req.UpdatedBy,
		Id,
	).Scan(
		&d.ProfileNo,
		&d.ProfileId,
		&d.RoleNo,
		&d.UserFullName,
		&d.EmailId,
		&d.Gender,
		&d.Dob,
		&d.MobileNo,
		&d.Address,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("userRepository.Modify: Modify: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("userRepository.Modify: commit: %w", err)
	}

	return &d, nil
}

func (r *userRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("userRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock row by profile_id before soft delete
	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT profile_no FROM master.users
		WHERE profile_id = $1 AND status != 'D'
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("userRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE master.users
		SET
			status     = 'D',
			updated_at = NOW()
		WHERE profile_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("userRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("userRepository.Remove: commit: %w", err)
	}

	return nil
}