package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/models"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/repository"
)

type customerCredentialRepository struct {
	db *pgxpool.Pool
}

func NewCustomerCredentialRepository(db *pgxpool.Pool) repository.CredentialRepository {
	return &customerCredentialRepository{db: db}
}

func (r *customerCredentialRepository) GetByCustomerNo(ctx context.Context, customerNo uint32) (*models.Credentials, error) {
	var c models.Credentials
	err := r.db.QueryRow(ctx, `
		SELECT credential_no, credential_id, customer_no, username, password, status, created_at, updated_at
		FROM customer.credentials
		WHERE customer_no = $1 AND status != 'D'
	`, customerNo).Scan(
		&c.CredentialNo, &c.CredentialId, &c.CustomerNo,
		&c.Username, &c.Password, &c.Status,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("customerCredentialRepository.GetByCustomerNo: %w", err)
	}
	return &c, nil
}

func (r *customerCredentialRepository) GetByEmail(ctx context.Context, email string) (*models.Credentials, error) {
	var c models.Credentials
	err := r.db.QueryRow(ctx, `
		SELECT c.credential_no, c.credential_id, c.customer_no, c.username, c.password, c.status, c.created_at, c.updated_at
		FROM customer.credentials c
		JOIN customer.customers cu ON cu.customer_no = c.customer_no
		WHERE cu.email_id = $1 AND c.status != 'D'
	`, email).Scan(
		&c.CredentialNo, &c.CredentialId, &c.CustomerNo,
		&c.Username, &c.Password, &c.Status,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("customerCredentialRepository.GetByEmail: %w", err)
	}
	return &c, nil
}

func (r *customerCredentialRepository) Save(ctx context.Context, cred *models.Credentials) (*models.Credentials, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("customerCredentialRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var c models.Credentials
	err = tx.QueryRow(ctx, `
		INSERT INTO customer.credentials (credential_id, customer_no, username, password, status)
		VALUES ($1, $2, $3, $4, 'A')
		RETURNING credential_no, credential_id, customer_no, username, password, status, created_at, updated_at
	`,
		cred.CredentialId,
		cred.CustomerNo,
		cred.Username,
		cred.Password,
	).Scan(
		&c.CredentialNo, &c.CredentialId, &c.CustomerNo,
		&c.Username, &c.Password, &c.Status,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("customerCredentialRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("customerCredentialRepository.Save: commit: %w", err)
	}

	return &c, nil
}

func (r *customerCredentialRepository) UpdatePassword(ctx context.Context, customerNo uint32, hashedPassword string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE customer.credentials
		SET password = $1, updated_at = NOW()
		WHERE customer_no = $2
	`, hashedPassword, customerNo)
	if err != nil {
		return fmt.Errorf("customerCredentialRepository.UpdatePassword: %w", err)
	}
	return nil
}
