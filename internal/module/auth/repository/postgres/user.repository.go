package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/repository"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByProfileId(ctx context.Context, profileId uuid.UUID) (*models.User, error) {
	query := `
		SELECT
			profile_no, profile_id, role_no,
			user_fullname, email_id, gender, dob,
			mobile_no, address, x_api_key, secret_key,
			status, created_at, created_by, updated_at, updated_by
		FROM auth.users
		WHERE profile_id = $1
		AND status != 'D'
	`

	var u models.User
	err := r.db.QueryRow(ctx, query, profileId).Scan(
		&u.ProfileNo, &u.ProfileId, &u.RoleNo,
		&u.UserFullName, &u.EmailId, &u.Gender, &u.Dob,
		&u.MobileNo, &u.Address, &u.XApiKey, &u.SecretKey,
		&u.Status, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.GetByProfileId: %w", err)
	}

	return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT
			profile_no, profile_id, role_no,
			user_fullname, email_id, gender, dob,
			mobile_no, address, x_api_key, secret_key,
			status, created_at, created_by, updated_at, updated_by
		FROM auth.users
		WHERE email_id = $1
		AND status != 'D'
	`

	var u models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ProfileNo, &u.ProfileId, &u.RoleNo,
		&u.UserFullName, &u.EmailId, &u.Gender, &u.Dob,
		&u.MobileNo, &u.Address, &u.XApiKey, &u.SecretKey,
		&u.Status, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.GetByEmail: %w", err)
	}

	return &u, nil
}

func (r *userRepository) GetByProfileNo(ctx context.Context, profileNo uint32) (*models.User, error) {
	query := `
		SELECT
			profile_no, profile_id, role_no,
			user_fullname, email_id, gender, dob,
			mobile_no, address, x_api_key, secret_key,
			status, created_at, created_by, updated_at, updated_by
		FROM auth.users
		WHERE profile_no = $1
		AND status != 'D'
	`

	var u models.User
	err := r.db.QueryRow(ctx, query, profileNo).Scan(
		&u.ProfileNo, &u.ProfileId, &u.RoleNo,
		&u.UserFullName, &u.EmailId, &u.Gender, &u.Dob,
		&u.MobileNo, &u.Address, &u.XApiKey, &u.SecretKey,
		&u.Status, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.GetByProfileNo: %w", err)
	}

	return &u, nil
}
