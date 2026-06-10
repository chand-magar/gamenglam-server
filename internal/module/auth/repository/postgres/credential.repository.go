package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/repository"
)

type credentialRepository struct {
	db *pgxpool.Pool
}

func NewCredentialRepository(db *pgxpool.Pool) repository.CredentialRepository {
	return &credentialRepository{db: db}
}

func (r *credentialRepository) GetByUsername(ctx context.Context, username string) (*models.Credentials, error) {
	query := `
		SELECT
			credential_no, credential_id, profile_no,
			username, password, status,
			created_at, created_by, updated_at, updated_by
		FROM auth.credentials
		WHERE username = $1
		AND status != 'D'
	`

	var c models.Credentials
	err := r.db.QueryRow(ctx, query, username).Scan(
		&c.CredentialNo, &c.CredentialId, &c.ProfileNo,
		&c.Username, &c.Password, &c.Status,
		&c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("credentialRepository.GetByUsername: %w", err)
	}

	return &c, nil
}

func (r *credentialRepository) GetByProfileNo(ctx context.Context, profileNo uint32) (*models.Credentials, error) {
	query := `
		SELECT
			credential_no, credential_id, profile_no,
			username, password, status,
			created_at, created_by, updated_at, updated_by
		FROM auth.credentials
		WHERE profile_no = $1
		AND status != 'D'
	`

	var c models.Credentials
	err := r.db.QueryRow(ctx, query, profileNo).Scan(
		&c.CredentialNo, &c.CredentialId, &c.ProfileNo,
		&c.Username, &c.Password, &c.Status,
		&c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("credentialRepository.GetByProfileNo: %w", err)
	}

	return &c, nil
}

func (r *credentialRepository) UpdatePassword(ctx context.Context, profileNo uint32, hashedPassword string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("credentialRepository.UpdatePassword: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock row before update
	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT credential_no FROM auth.credentials
		WHERE profile_no = $1 AND status != 'D'
		FOR UPDATE
	`, profileNo).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("credentialRepository.UpdatePassword: credentials not found")
		}
		return fmt.Errorf("credentialRepository.UpdatePassword: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE auth.credentials
		SET
			password   = $1,
			updated_at = NOW()
		WHERE profile_no = $2
	`, hashedPassword, profileNo)
	if err != nil {
		return fmt.Errorf("credentialRepository.UpdatePassword: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("credentialRepository.UpdatePassword: commit: %w", err)
	}

	return nil
}
